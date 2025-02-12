// Package shares реализует обработчик шар, полученных из топика кафки
package shares

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/dnsoftware/mpm-save-get-shares/pkg/kafka_reader"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	"github.com/dnsoftware/mpm-shares-processor/dto"
	"github.com/dnsoftware/mpm-shares-processor/entity"
)

type Config struct {
	BatchSize     int           // Размер буфера для пакетного чтения
	FlushInterval time.Duration // Максимальное время ожидания для заполнения пакета в секундах
}

type Processor interface {
	NormalizeShare(ctxTask context.Context, shareFound dto.ShareFound) (entity.Share, error)
	AddSharesBatch(shares []entity.Share) error
}

// ShareConsumer реализует интерфейс sarama.ConsumerGroupHandler
type ShareConsumer struct {
	cfg         Config
	kafkaReader *kafka_reader.KafkaReader
	msgChan     chan *sarama.ConsumerMessage
	Processor
}

func NewShareConsumer(cfg Config, kafkaReader *kafka_reader.KafkaReader, processor Processor) (*ShareConsumer, error) {
	return &ShareConsumer{
		cfg:         cfg,
		kafkaReader: kafkaReader,
		msgChan:     make(chan *sarama.ConsumerMessage),
		Processor:   processor,
	}, nil
}

// StartConsume Стартует чтение из Кафки
func (consumer *ShareConsumer) StartConsume() {

	go func() {
		consumer.kafkaReader.ConsumeMessages(consumer)
	}()

}

// GetConsumeChan возвращает канал куда пишуться считанные сообщения
//func (consumer *ShareConsumer) GetConsumeChan() chan *sarama.ConsumerMessage {
//	return consumer.msgChan
//}

func (consumer *ShareConsumer) Close() {
	consumer.kafkaReader.Close()
}

// Setup вызывается перед началом обработки (интерфейс ConsumerGroupHandler)
func (consumer *ShareConsumer) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup вызывается после завершения обработки (интерфейс ConsumerGroupHandler)
func (consumer *ShareConsumer) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim обрабатывает сообщения из партиций (интерфейс ConsumerGroupHandler)
// если используем канал consumer.MsgChan (или какой-то подобный) - нужно его вычитывать снаружи
func (consumer *ShareConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {

	var item dto.ShareFound
	var batch []*sarama.ConsumerMessage // Буфер для пакетного чтения
	timer := time.NewTimer(consumer.cfg.FlushInterval * time.Second)

	for msg := range claim.Messages() {

		// Создаем carrier для извлечения заголовков
		carrier := propagation.MapCarrier{}
		// Извлекаем контекст трассировки из заголовков
		// Преобразуем заголовки Kafka в формат map
		for _, header := range msg.Headers {
			carrier[string(header.Key)] = string(header.Value)
		}
		ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)
		tracer := otel.Tracer("consume-share")
		ctx, span := tracer.Start(ctx, "process")

		// Функция для обработки пакета
		processBatch := func() error {
			if len(batch) > 0 {
				sharesBatch := make([]entity.Share, 0, len(batch))

				start := time.Now().UnixMilli()
				logger.Log().Info(fmt.Sprintf("Processing batch of %d messages", len(batch)))
				for _, mess := range batch {
					if mess == nil {
						fmt.Println("batch message nil")
						continue
					}
					err := json.Unmarshal(mess.Value, &item)
					normShare, err := consumer.NormalizeShare(ctx, item)
					if err != nil {
						logger.Log().Error("NormalizeShare error: " + err.Error())
						return fmt.Errorf("NormalizeShare error: %s", err.Error())
					}
					sharesBatch = append(sharesBatch, normShare)
				}
				end := time.Now().UnixMilli()
				fmt.Println(fmt.Sprintf("Batch time: %v", end-start))

				err := consumer.AddSharesBatch(sharesBatch)
				if err != nil {
					return err
				}

				// Помечаем смещения для пакета как прочитанные
				session.MarkOffset(batch[len(batch)-1].Topic, batch[len(batch)-1].Partition, batch[len(batch)-1].Offset+1, "")
				batch = nil // Очищаем пакет после обработки

			}
			timer.Reset(consumer.cfg.FlushInterval * time.Second) // Сбрасываем таймер

			return nil
		}

		for {
			select {
			case message, ok := <-claim.Messages():
				if !ok {
					fmt.Println("Channel claim.Messages() closed")
					return nil
				}

				if message == nil {
					fmt.Println("claim.Messages nil")
					continue
				}
				batch = append(batch, message)
				if len(batch) >= consumer.cfg.BatchSize {
					err := processBatch()
					if err != nil {
						return err
					}
				}
			case <-timer.C:
				// Если сработал таймер, обрабатываем текущий пакет
				err := processBatch()
				if err != nil {
					return err
				}

			case <-session.Context().Done():
				// Завершаем работу при остановке сессии
				return nil
			}
		}

		//consumer.msgChan <- msg

		// Если сообщение успешно обработано - помечаем, как обработанное TODO
		// session.MarkMessage(msg, "")
		// иначе - логируем ошибку и делаем пометку "алерт" TODO
		// ...

		span.End()
	}

	return nil
}
