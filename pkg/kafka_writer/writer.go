package kafka_writer

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"

	"github.com/dnsoftware/mpm-shares-processor/pkg/logger"
)

type Config struct {
	Brokers []string
	Topic   string
}

// KafkaWriter - структура для асинхронного продюсера
type KafkaWriter struct {
	brokers  []string
	topic    string
	producer sarama.AsyncProducer
	stop     chan os.Signal
	wg       sync.WaitGroup
	logger   logger.MPMLogger
}

// NewKafkaWriter - конструктор для создания нового KafkaProducer
func NewKafkaWriter(cfg Config, logger logger.MPMLogger) (*KafkaWriter, error) {
	// Конфигурация продюсера
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForLocal        // Дождаться подтверждения от лидера
	config.Producer.Retry.Max = 5                             // Количество попыток повторной отправки
	config.Producer.Return.Successes = true                   // Возвращать успешные отправки
	config.Producer.Return.Errors = true                      // Возвращать ошибки отправки
	config.Producer.Partitioner = sarama.NewRandomPartitioner // Случайное распределение по партициям
	config.Producer.Retry.Backoff = 200 * time.Millisecond    // Задержка между попытками

	// Настройка пакетной отправки
	config.Producer.Flush.Frequency = 500 * time.Millisecond // Отправлять каждые 500 мс
	config.Producer.Flush.Bytes = 1024 * 1024                // Отправлять при накоплении 1 МБ данных
	config.Producer.Flush.Messages = 100                     // Отправлять каждые 100 сообщений
	config.Producer.Flush.MaxMessages = 1000                 // Максимум 1000 сообщений в пакете

	// Ограничения по размеру сообщений
	config.Producer.MaxMessageBytes = 10 * 1024 * 1024 // Максимальный размер одного сообщения: 10 МБ

	// Создание асинхронного продюсера
	producer, err := sarama.NewAsyncProducer(cfg.Brokers, config)
	if err != nil {
		return nil, err
	}

	return &KafkaWriter{
		brokers:  cfg.Brokers,
		topic:    cfg.Topic,
		producer: producer,
		stop:     make(chan os.Signal, 1),
		logger:   logger,
	}, nil
}

// Start - метод для запуска продюсера
func (k *KafkaWriter) Start() {
	// Обработка сигналов для завершения работы
	signal.Notify(k.stop, os.Interrupt, syscall.SIGTERM)

	// Горутина для обработки успешных сообщений
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for success := range k.producer.Successes() {
			// TODO удалить для ускорения
			log.Printf("Сообщение успешно отправлено в партицию %d, с оффсетом %d", success.Partition, success.Offset)
		}
	}()

	// Горутина для обработки ошибок
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for err := range k.producer.Errors() {
			k.logger.Error(fmt.Sprintf("Ошибка отправки сообщения: %v", err.Err))
		}
	}()
}

// SendMessage - метод для отправки сообщения в Kafka
func (k *KafkaWriter) SendMessage(ctx context.Context, key string, value string) {

	message := &sarama.ProducerMessage{
		Topic: k.topic,
		Key:   sarama.StringEncoder(key), // Ключ сообщения
		Value: sarama.StringEncoder(value),
	}

	// Заголовки Kafka и инъекция контекста трассировки в них
	headers := SaramaHeadersCarrier(make([]sarama.RecordHeader, 0))
	propagator := otel.GetTextMapPropagator()
	propagator.Inject(ctx, &headers) // Передаём указатель на адаптер
	message.Headers = headers

	// Отправка сообщения
	k.producer.Input() <- message
}

// Close - метод для закрытия продюсера
func (k *KafkaWriter) Close() {
	k.producer.AsyncClose() // завершение работы асинхронного продюсера и закрытие канало Errors() и Successes()
	k.wg.Wait()             // Ожидание завершения горутин обработки Errors() и Successes()

	k.logger.Info(fmt.Sprintf("Работа KafkaWriter завершена"))
}

func (k *KafkaWriter) DeleteTopic(topic string) error {
	// Создаем конфигурацию Sarama
	config := sarama.NewConfig()
	//config.Version = sarama.V3_0_0_0 // Убедитесь, что используемая версия Kafka поддерживает удаление топиков

	// Создаем ClusterAdmin
	admin, err := sarama.NewClusterAdmin(k.brokers, config)
	if err != nil {
		return err
	}
	defer admin.Close()

	// Удаляем топик
	err = admin.DeleteTopic(topic)
	if err != nil {
		return err
	}

	return nil
}
