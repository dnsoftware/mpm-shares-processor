package config

import (
	"log"
	"testing"

	"github.com/dnsoftware/mpmslib/pkg/configloader"
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
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env"

	startConf, err := configloader.LoadStartConfig(basePath + constants.StartConfigFilename)
	if err != nil {
		log.Fatalf("start config load error: %s", err)
	}

	cfg, err := New(configFile, envFile)
	require.NoError(t, err)

	cfg.App.AppID = startConf.AppID
	cfg.EtcdConfig.Endpoints = startConf.Etcd.Endpoints
	cfg.EtcdConfig.Username = startConf.Etcd.Auth.Username
	cfg.EtcdConfig.Password = startConf.Etcd.Auth.Password

	assert.Equal(t, constants.KafkaSharesTopic, cfg.KafkaShareReader.Topic)
	assert.Equal(t, "127.0.0.1:7878", cfg.GRPC.CoinTarget)
	assert.Equal(t, "127.0.0.1:6878", cfg.GRPC.SharesTarget)

}
