package kafka_reader

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/IBM/sarama"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
)

type Config struct {
	Brokers            []string
	Group              string
	Topic              string
	AutoCommitEnable   bool
	AutoCommitInterval int
}

type KafkaReader struct {
	brokers       []string
	group         string
	topic         string
	consumerGroup sarama.ConsumerGroup
	ctx           context.Context
	cancel        context.CancelFunc
	wg            sync.WaitGroup
	logger        logger.MPMLogger
	admin         sarama.ClusterAdmin
}

func NewKafkaReader(cfg Config, logger logger.MPMLogger) (*KafkaReader, error) {
	// Настройка конфигурации
	config := sarama.NewConfig()
	config.Consumer.Offsets.Initial = sarama.OffsetOldest                                             // Чтение с самого начала (если нет сохраненного оффсета)
	config.Consumer.Offsets.AutoCommit.Enable = cfg.AutoCommitEnable                                  // Включаем автоматическое сохранение оффсетов
	config.Consumer.Offsets.AutoCommit.Interval = time.Duration(cfg.AutoCommitInterval) * time.Second // Интервал для сохранения оффсетов - 10 секунд

	// Создаем клиента для consumer группы
	consumerGroup, err := sarama.NewConsumerGroup(cfg.Brokers, cfg.Group, config)
	if err != nil {
		return nil, err
	}

	// Обработка ошибок в отдельной горутине
	go func() {
		for err := range consumerGroup.Errors() {
			log.Printf("Error: %v", err)
		}
	}()

	admin, err := sarama.NewClusterAdmin(cfg.Brokers, config)
	if err != nil {
		log.Fatalf("Ошибка при создании ClusterAdmin: %v", err)
	}

	// Создание контекста для управления остановкой
	ctx, cancel := context.WithCancel(context.Background())

	return &KafkaReader{
		brokers:       cfg.Brokers,
		group:         cfg.Group,
		topic:         cfg.Topic,
		consumerGroup: consumerGroup,
		ctx:           ctx,
		cancel:        cancel,
		logger:        logger,
		admin:         admin,
	}, nil
}

// ConsumeMessages - запуск чтения сообщений из топика
func (r *KafkaReader) ConsumeMessages(handler sarama.ConsumerGroupHandler) {
	r.wg.Add(1)
	go func() {
		defer r.wg.Done()

		// Бесконечный цикл чтения сообщений
		for {
			// Если контекст завершён, выходим из цикла
			if r.ctx.Err() != nil {
				return
			}

			if err := r.consumerGroup.Consume(r.ctx, []string{r.topic}, handler); err != nil {
				r.logger.Error(fmt.Sprintf("Ошибка при чтении сообщений, Group: %s, Topic: %s: %v", r.group, r.topic, err))
			}

		}
	}()

	return
}

func (r *KafkaReader) Close() {
	// Отмена контекста
	r.cancel()
	// Ожидание завершения горутин
	r.wg.Wait()
	// Закрытие Consumer Group
	if err := r.consumerGroup.Close(); err != nil {
		r.logger.Error(fmt.Sprintf("Ошибка закрытия Consumer Group, Group: %s, Topic: %s: %v", r.group, r.topic, err))
	}

	if err := r.admin.Close(); err != nil {
		r.logger.Error(fmt.Sprintf("Ошибка закрытия ClusterAdmin: %v", err))
	}

	r.logger.Info(fmt.Sprintf("KafkaConsumer завершён, Group: %s, Topic: %s", r.group, r.topic))
}

// SetGroupOffset Сбрасывает для текущей группы смещение топика в начало
func (r *KafkaReader) SetGroupOffset(offset int64) error {
	// Создание нового клиента
	client, err := sarama.NewClient(r.brokers, nil)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create Kafka client: %s", err))
		return err
	}
	defer client.Close()

	// Создание OffsetManager
	offsetManager, err := sarama.NewOffsetManagerFromClient(r.group, client)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to create offset manager: %s", err))
		return err
	}
	defer offsetManager.Close()

	// Получение информации о партициях топика
	partitions, err := client.Partitions(r.topic)
	if err != nil {
		r.logger.Error(fmt.Sprintf("Failed to get partitions: %s", err))
	}

	// Установка конкретных смещений по партициям
	for _, partition := range partitions {
		partitionManager, err := offsetManager.ManagePartition(r.topic, partition)
		if err != nil {
			r.logger.Error(fmt.Sprintf("Failed to manage partition %d: %v", partition, err))
			continue
		}
		defer partitionManager.Close()

		// Устанавливаем смещение на начало
		desiredOffset := offset

		// Устанавливаем смещение
		partitionManager.ResetOffset(desiredOffset, "")

		r.logger.Info(fmt.Sprintf("Set offset for partition %d to %d\n", partition, desiredOffset))
	}

	return nil
}
