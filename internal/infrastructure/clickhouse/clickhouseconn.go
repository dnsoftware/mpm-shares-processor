package clickhouse

import (
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type Config struct {
	Addr             []string // список нод
	Database         string   // назхвание БД
	Username         string   // имя пользователя базы
	Password         string   // пароль пользователя
	MaxExecutionTime int
}

func NewClickhouseConnect(cfg Config) (driver.Conn, error) {

	maxExecutionTime := 60
	if cfg.MaxExecutionTime > 0 {
		maxExecutionTime = cfg.MaxExecutionTime
	}

	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Addr,
		Auth: clickhouse.Auth{
			Database: cfg.Database,
			Username: cfg.Username,
			Password: cfg.Password,
		},
		DialTimeout: 5 * time.Second,
		Compression: &clickhouse.Compression{
			Method: clickhouse.CompressionLZ4,
		},
		Settings: map[string]interface{}{
			"max_execution_time": maxExecutionTime,
		},
		MaxOpenConns: 10,
		MaxIdleConns: 5,
		Debug:        true,
	})

	if err != nil {
		return nil, err
	}

	return conn, nil
}
