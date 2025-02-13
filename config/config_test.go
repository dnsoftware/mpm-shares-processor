package config

import (
	"log"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"

	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"
)

func TestConfigNew(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config_example.yaml"
	envFile := basePath + "/.env_example"

	cfg, err := New(configFile, envFile)
	require.NoError(t, err)

	assert.Equal(t, constants.KafkaSharesTopic, cfg.KafkaShareReader.Topic)
	assert.Equal(t, "127.0.0.1:7878", cfg.GRPC.CoinTarget)
	assert.Equal(t, "127.0.0.1:6878", cfg.GRPC.SharesTarget)

}
