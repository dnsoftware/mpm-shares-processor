/************** DO NOT EDIT ****************/
package kafka

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

// Эталонный образец от testcontainers
func TestKafkaWriteReadExample(t *testing.T) {
	topic := "some-topic"

	ctx := context.Background()

	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("kraftCluster"))
	testcontainers.CleanupContainer(t, kafkaContainer)
	require.NoError(t, err)

	assertAdvertisedListeners(t, kafkaContainer)

	if !strings.EqualFold(kafkaContainer.ClusterID, "kraftCluster") {
		t.Fatalf("expected clusterID to be %s, got %s", "kraftCluster", kafkaContainer.ClusterID)
	}

	// getBrokers {
	brokers, err := kafkaContainer.Brokers(ctx)
	// }
	require.NoError(t, err)

	config := sarama.NewConfig()
	client, err := sarama.NewConsumerGroup(brokers, "groupName", config)
	require.NoError(t, err)

	consumer, ready, done, cancel := NewTestKafkaConsumer(t)
	defer cancel()
	go func() {
		if err := client.Consume(context.Background(), []string{topic}, consumer); err != nil {
			cancel()
		}
	}()

	// wait for the consumer to be ready
	<-ready

	// perform assertions

	// set config to true because successfully delivered messages will be returned on the Successes channel
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(brokers, config)
	require.NoError(t, err)

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder("key"),
		Value: sarama.StringEncoder("value"),
	})
	require.NoError(t, err)

	<-done

	if !strings.EqualFold(string(consumer.message.Key), "key") {
		t.Fatalf("expected key to be %s, got %s", "key", string(consumer.message.Key))
	}
	if !strings.EqualFold(string(consumer.message.Value), "value") {
		t.Fatalf("expected value to be %s, got %s", "value", string(consumer.message.Value))
	}
}

// assertAdvertisedListeners checks that the advertised listeners are set correctly:
// - The BROKER:// protocol is using the hostname of the Kafka container
func assertAdvertisedListeners(t *testing.T, container testcontainers.Container) {
	t.Helper()
	inspect, err := container.Inspect(context.Background())
	require.NoError(t, err)

	hostname := inspect.Config.Hostname

	code, r, err := container.Exec(context.Background(), []string{"cat", "/usr/sbin/testcontainers_start.sh"})
	require.NoError(t, err)

	require.Zero(t, code)

	bs, err := io.ReadAll(r)
	require.NoError(t, err)

	if !strings.Contains(string(bs), "BROKER://"+hostname+":9092") {
		t.Fatalf("expected advertised listeners to contain %s, got %s", "BROKER://"+hostname+":9092", string(bs))
	}
}

// TestKafkaConsumer is a test consumer for Kafka
type TestKafkaConsumer struct {
	t       *testing.T
	ready   chan bool
	done    chan bool
	cancel  chan bool
	message *sarama.ConsumerMessage
}

func NewTestKafkaConsumer(t *testing.T) (*TestKafkaConsumer, <-chan bool, <-chan bool, func()) {
	t.Helper()
	kc := &TestKafkaConsumer{
		t:      t,
		ready:  make(chan bool, 1),
		done:   make(chan bool, 1),
		cancel: make(chan bool, 1),
	}
	return kc, kc.ready, kc.done, func() {
		kc.cancel <- true
	}
}

func (k *TestKafkaConsumer) Setup(_ sarama.ConsumerGroupSession) error {
	return nil
}

func (k *TestKafkaConsumer) Cleanup(_ sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim is called by the Kafka client library when a message is received
func (k *TestKafkaConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	k.ready <- true
	for {
		select {
		case message := <-claim.Messages():
			k.message = message
			session.MarkMessage(message, "")
			k.done <- true

		case <-k.cancel:
			return nil

		case <-session.Context().Done():
			return nil
		}
	}
}
