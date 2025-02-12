package _example

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

func example() {
	// Настройки для подключения к Kafka
	config := sarama.NewConfig()

	// Создание группы потребителей
	consumerGroup, err := sarama.NewConsumerGroup([]string{"localhost:9092"}, "example-group", config)
	if err != nil {
		log.Fatal("Error creating consumer group:", err)
	}
	defer func() {
		if err := consumerGroup.Close(); err != nil {
			log.Fatal("Error closing consumer group:", err)
		}
	}()

	// Обработчик сообщений
	go func() {
		for err := range consumerGroup.Errors() {
			log.Println("Error from consumer group:", err)
		}
	}()

	consumer := &Consumer{}

	// Чтение из топика
	for {
		err := consumerGroup.Consume(nil, []string{"example-topic"}, consumer)
		if err != nil {
			log.Fatal("Error from consumer group:", err)
		}
	}
}

// Определение интерфейса для обработки сообщений
type Consumer struct{}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		fmt.Printf("Received message: %s\n", string(msg.Value))
		session.MarkMessage(msg, "")
	}
	return nil
}
