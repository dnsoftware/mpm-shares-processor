package _example

import (
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestKafkaConsumer(t *testing.T) {
	// Настройка Kafka
	brokers := []string{"localhost:9092"}
	topic := "test-topic"
	group := "test-group"

	// Создание продюсера для отправки тестового сообщения
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		t.Fatalf("Ошибка создания продюсера: %v", err)
	}
	defer producer.Close()

	// Отправка тестового сообщения
	message := "test message"
	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	})
	if err != nil {
		t.Fatalf("Ошибка отправки сообщения: %v", err)
	}

	// Создание потребителя
	config := sarama.NewConfig()
	config.Version = sarama.V2_8_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	client, err := sarama.NewConsumerGroup(brokers, group, config)
	if err != nil {
		t.Fatalf("Ошибка создания consumer group: %v", err)
	}
	defer client.Close()

	// Чтение сообщения
	consumer := &testConsumer{
		messages: make(chan string, 1),
	}
	go func() {
		for {
			err := client.Consume(nil, []string{topic}, consumer)
			if err != nil {
				t.Errorf("Ошибка чтения: %v", err)
			}
		}
	}()

	// Ожидание сообщения
	select {
	case msg := <-consumer.messages:
		assert.Equal(t, message, msg, "Сообщение должно совпадать")
	case <-time.After(5 * time.Second):
		t.Fatal("Тест завершился по таймауту, сообщение не получено")
	}
}

// testConsumer реализует интерфейс sarama.ConsumerGroupHandler
type testConsumer struct {
	messages chan string
}

func (c *testConsumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *testConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *testConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for message := range claim.Messages() {
		c.messages <- string(message.Value)
		session.MarkMessage(message, "")
	}
	return nil
}
