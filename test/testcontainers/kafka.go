package testcontainers

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/kafka"
)

func NewKafkaTestcontainer(t *testing.T) (*kafka.KafkaContainer, error) {
	ctx := context.Background()
	kafkaContainer, err := kafka.Run(ctx, "confluentinc/confluent-local:7.5.0", kafka.WithClusterID("testCluster"))
	if err != nil {
		return nil, err
	}
	testcontainers.CleanupContainer(t, kafkaContainer)

	if !strings.EqualFold(kafkaContainer.ClusterID, "testCluster") {
		return nil, fmt.Errorf("expected clusterID to be %s, got %s", "testCluster", kafkaContainer.ClusterID)
	}

	return kafkaContainer, nil
}
