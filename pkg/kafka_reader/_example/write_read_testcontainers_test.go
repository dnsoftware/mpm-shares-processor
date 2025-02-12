package _example

import (
	"context"
	"testing"
	"time"

	"github.com/IBM/sarama"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

func TestKafkaWithTestcontainers(t *testing.T) {
	// Настройка Kafka через testcontainers
	ctx := context.Background()
	req := testcontainers.ContainerRequest{
		Image:        "confluentinc/cp-kafka:latest",
		ExposedPorts: []string{"9092/tcp"},
		Env: map[string]string{
			"KAFKA_BROKER_ID":                        "1",
			"KAFKA_ZOOKEEPER_CONNECT":                "localhost:2181",
			"KAFKA_ADVERTISED_LISTENERS":             "PLAINTEXT://localhost:9092",
			"KAFKA_OFFSETS_TOPIC_REPLICATION_FACTOR": "1",
		},
		WaitingFor: wait.ForLog("Kafka Server started"),
	}
	kafkaContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		t.Fatalf("Ошибка запуска Kafka-контейнера: %v", err)
	}
	defer kafkaContainer.Terminate(ctx)

	// Настройка Kafka клиента
	brokers := []string{"localhost:9092"}
	topic := "test-topic"
	message := "test message"

	// Отправка сообщения
	producer, err := sarama.NewSyncProducer(brokers, nil)
	if err != nil {
		t.Fatalf("Ошибка создания продюсера: %v", err)
	}
	defer producer.Close()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(message),
	})
	if err != nil {
		t.Fatalf("Ошибка отправки сообщения: %v", err)
	}

	// Чтение сообщения
	consumer, err := sarama.NewConsumer(brokers, nil)
	if err != nil {
		t.Fatalf("Ошибка создания потребителя: %v", err)
	}
	defer consumer.Close()

	partitionConsumer, err := consumer.ConsumePartition(topic, 0, sarama.OffsetOldest)
	if err != nil {
		t.Fatalf("Ошибка создания партиционного потребителя: %v", err)
	}
	defer partitionConsumer.Close()

	select {
	case msg := <-partitionConsumer.Messages():
		if string(msg.Value) != message {
			t.Fatalf("Ожидалось: %s, получено: %s", message, string(msg.Value))
		}
	case <-time.After(10 * time.Second):
		t.Fatal("Тест завершился по таймауту, сообщение не получено")
	}
}
