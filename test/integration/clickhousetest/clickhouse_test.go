package clickhousetest

import (
	"context"
	"fmt"
	"log"
	"testing"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/golang-migrate/migrate/v4"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/dnsoftware/mpm-shares-processor/config"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
	clickhouse2 "github.com/dnsoftware/mpm-shares-processor/internal/infrastructure/clickhouse"
)

// Должен быть запущен кластер Clickhouse
func TestAddShare(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	require.NoError(t, err)
	configFile := basePath + "/config_example.yaml"
	envFile := basePath + "/.env_example"

	cfg, err := config.New(configFile, envFile)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Подключение к базе данных ClickHouse
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Clickhouse.Addr,
		Auth: clickhouse.Auth{
			Database: "default",
			Username: cfg.Clickhouse.Username,
			Password: cfg.Clickhouse.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Debug: true,
	})
	require.NoError(t, err)

	// Проверка подключения
	err = conn.Ping(ctx)
	require.NoError(t, err)

	// Укажите путь к миграциям и строку подключения к базе данных
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s", "default", "", cfg.Clickhouse.Addr[0], "default")
	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir,
		dsn,
	)
	require.NoError(t, err)

	// Сброс миграций
	m.Force(-1)

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		require.NoError(t, err)
	}
	log.Println("Миграции успешно применены")

	conn.Close()

	// Подключаемся к mpmhouse
	conn, err = clickhouse2.NewClickhouseConnect(clickhouse2.Config{
		Addr:             cfg.Clickhouse.Addr,
		Database:         cfg.Clickhouse.Database,
		Username:         cfg.Clickhouse.Username,
		Password:         cfg.Clickhouse.Password,
		MaxExecutionTime: 10,
	})
	require.NoError(t, err)
	defer conn.Close()

	// Проверка подключения
	err = conn.Ping(ctx)
	require.NoError(t, err)

	cfgStore := clickhouse2.ShareStorageConfig{
		Conn:        conn,
		ClusterName: "clickhouse_cluster",
		Database:    "mpmhouse",
	}
	store, err := clickhouse2.NewClickhouseShareStorage(cfgStore)
	require.NoError(t, err)

	// Вставка записи
	ctx = context.Background()
	share := entity.Share{
		UUID:         "9e7b8188-01a6-4989-8805-e4337e31195a",
		ServerID:     "EU-HSHP-ALPH-1",
		CoinID:       4,
		WorkerID:     4,
		WalletID:     2,
		ShareDate:    time.Now().UnixMilli(),
		Difficulty:   "0.008941",
		Sharedif:     "5.14677",
		Nonce:        "9c44010001030201010202030400040402040304915711c0",
		IsSolo:       false,
		RewardMethod: "PPLNS",
		Cost:         "0.00124",
	}

	err = store.AddShare(ctx, share)
	require.NoError(t, err)

	// Получение записи
	shareFrom, err := store.GetShareRow(ctx, "9e7b8188-01a6-4989-8805-e4337e31195a")
	require.NoError(t, err)

	require.Equal(t, shareFrom.ServerID, share.ServerID)

	// Пакетная вставка
	var shares []entity.Share
	for i := 0; i < 100000; i++ {
		share = entity.Share{
			UUID:         uuid.New().String(),
			ServerID:     "EU-HSHP-ALPH-1",
			CoinID:       4,
			WorkerID:     4,
			WalletID:     2,
			ShareDate:    time.Now().UnixMilli(),
			Difficulty:   "0.008941",
			Sharedif:     "5.14677",
			Nonce:        "9c44010001030201010202030400040402040304915711c0",
			IsSolo:       false,
			RewardMethod: "PPLNS",
			Cost:         "0.00124",
		}

		shares = append(shares, share)
	}

	start := time.Now().UnixMicro()
	err = store.AddSharesBatch(ctx, shares)
	end := time.Now().UnixMicro()
	require.NoError(t, err)

	fmt.Println(fmt.Sprintf("%v", end-start))

}
