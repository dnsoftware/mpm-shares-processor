package etcd

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"
)

func TestNewEtcdClient(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	require.NoError(t, err)

	cfg := EtcdConfig{
		Nodes:       strings.Split("127.0.0.1:2379,127.0.0.1:2479,127.0.0.1:2579", ","),
		Username:    "root",
		Password:    "securepassword",
		CertCaPath:  basePath + "/certs/ca.crt",
		CertPath:    basePath + "/certs/client.crt",
		CertKeyPath: basePath + "/certs/client.key",
	}

	cli, err := NewEtcdClient(cfg)
	_ = cli

	_, err = cli.Put(context.Background(), "/test_etcd", "test")
	require.NoError(t, err)
	resp, err := cli.Get(context.Background(), "/test_etcd")
	data := string(resp.Kvs[0].Value)
	require.Equal(t, "test", data)

}
