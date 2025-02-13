package kafka_writer

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/pkg/kafka_reader"
	"github.com/dnsoftware/mpm-shares-processor/pkg/logger"
	otelpkg "github.com/dnsoftware/mpm-shares-processor/pkg/otel"
)

// Тестирование отправки шар в кафку
// кафка должна быть запущена
// + трассировка
// используем запущенные на локалке kafka и otel collector
func TestWriter(t *testing.T) {

	ctx := context.Background()

	otelConfig := otelpkg.Config{
		ServiceName:        "TestService",
		CollectorEndpoint:  "localhost:4317",
		BatchTimeout:       1 * time.Second,
		MaxExportBatchSize: 100,
		MaxQueueSize:       500,
	}
	cleanup := otelpkg.InitTracer(otelConfig)
	defer cleanup()

	tracer := otel.Tracer("share-trace")

	ctxAction, spanAction := tracer.Start(ctx, "share-trace-start")

	filePath, err := logger.GetLoggerMainLogPath()
	require.NoError(t, err)
	logger.InitLogger(logger.LogLevelDebug, filePath)
	//logger.InitLogger(logger.LogLevelProduction, filePath)

	cfg := Config{
		Brokers: []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Topic:   "test_topic",
	}

	writer, err := NewKafkaWriter(cfg, logger.Log())
	assert.NoError(t, err)
	defer writer.Close()

	spanAction.End()

	_, spanTopicDelete := tracer.Start(ctxAction, "kafka-topic-delete")

	err = writer.DeleteTopic(writer.topic)
	//assert.NoError(t, err)

	spanTopicDelete.End()

	_, spanProducerStart := tracer.Start(ctxAction, "kafka-producer-start")
	// Запуск продюсера
	writer.Start()
	spanProducerStart.End()

	// Отправка сообщений
	_, spanSend := tracer.Start(ctxAction, "send-message")
	msgSend := fmt.Sprintf("%v", time.Now().Nanosecond())
	writer.SendMessage(ctxAction, "test_write", msgSend)
	spanSend.End()

	//////////////////////////////////////// читаем из топика
	_, spanRead := tracer.Start(ctxAction, "topic-read")
	cfgReader := kafka_reader.Config{
		Brokers:            []string{"localhost:9092", "localhost:9093", "localhost:9094"},
		Group:              constants.KafkaSharesGroup,
		Topic:              "test_topic", // constants.KafkaSharesTopic
		AutoCommitInterval: constants.KafkaSharesAutocommitInterval,
		AutoCommitEnable:   true,
	}

	reader, err := kafka_reader.NewKafkaReader(cfgReader, logger.Log())
	assert.NoError(t, err)
	defer reader.Close()

	err = reader.SetGroupOffset(sarama.OffsetNewest)
	assert.NoError(t, err)

	// Читаем сообщение
	msgChan := make(chan *sarama.ConsumerMessage)
	handler := &testConsumerGroupHandler{msgChan: msgChan}

	go func() {
		reader.ConsumeMessages(handler)
	}()
	spanRead.End()

	// Получаем сообщения
	select {
	case msg := <-msgChan:
		assert.Equal(t, msgSend, string(msg.Value))
	case <-time.After(6 * time.Second):
		t.Fatal("Таймаут при получении сообщения")
	}

}

// Тестовый обработчик ConsumerGroup
type testConsumerGroupHandler struct {
	msgChan chan *sarama.ConsumerMessage
}

func (h *testConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *testConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h *testConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
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
		h.msgChan <- msg
		sess.MarkMessage(msg, "")

		span.End()
		return nil
	}
	return nil
}
