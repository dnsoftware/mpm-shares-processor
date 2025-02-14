package servicediscovery

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	clientv3 "go.etcd.io/etcd/client/v3"
)

type EtcdConfig struct {
	Nodes       []string // список нод кластера etcd
	Username    string   // логин доступа к etcd
	Password    string   // пароль доступа к etcd
	CertCaPath  string   // путь к файлу корневого сертификата
	CertPath    string   // путь к файлу публичного сертификата
	CertKeyPath string   // путь к файлу приватного ключа сертификата
}

func NewEtcdConfig(cfg EtcdConfig) (*clientv3.Config, error) {
	// Root certificate CA load
	caCert, err := os.ReadFile(cfg.CertCaPath)
	if err != nil {
		return nil, fmt.Errorf("error reading CA certificate: %w", err)
	}

	// Make certs pool
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	// Load cert and key
	cert, err := tls.LoadX509KeyPair(cfg.CertPath, cfg.CertKeyPath)
	if err != nil {
		return nil, fmt.Errorf("error certificates loading: %w", err)
	}
	// Make tls.Config with a configured root certificate
	tlsConfig := &tls.Config{
		RootCAs:      caCertPool,
		Certificates: []tls.Certificate{cert},
		// Enabling verification of the server certificate
		InsecureSkipVerify: false,
	}
	_ = tlsConfig

	conf := clientv3.Config{
		Endpoints:   cfg.Nodes,
		DialTimeout: 5 * time.Second,
		TLS:         tlsConfig,
		Username:    cfg.Username,
		Password:    cfg.Password,
	}
	if err != nil {
		return nil, err
	}

	return &conf, nil
}
