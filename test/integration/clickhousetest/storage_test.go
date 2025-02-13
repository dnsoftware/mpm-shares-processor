package clickhousetest

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	"github.com/dnsoftware/mpm-shares-processor/config"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/internal/infrastructure/clickhouse"
)

func TestStorage(t *testing.T) {
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	require.NoError(t, err)
	configFile := basePath + "/config_example.yaml"
	envFile := basePath + "/.env_example"

	cfg, err := config.New(configFile, envFile)
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// Подключаемся к mpmhouse
	conn, err := clickhouse.NewClickhouseConnect(clickhouse.Config{
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

	cfgStore := clickhouse.ShareStorageConfig{
		Conn:        conn,
		ClusterName: constants.ClickhouseCluster,
		Database:    cfg.Clickhouse.Database,
	}
	store, err := clickhouse.NewClickhouseShareStorage(cfgStore)
	require.NoError(t, err)

	// Тест текущей сложности раунда
	t.Run("Test Round difficulty", func(t *testing.T) {
		ctx = context.Background()
		// Формат, соответствующий строке
		layout := "2006-01-02 15:04:05.999"
		dateStart, err := time.Parse(layout, "2025-02-12 19:09:06.803000000")
		require.NoError(t, err)

		dateEnd := dateStart.Add(2000000 * time.Second)
		start := time.Now().UnixMilli()
		sum, count, err := store.RangeDifficultySum(ctx, dateStart, dateEnd, 4, "PPLNS")
		require.NoError(t, err)

		delta := time.Now().UnixMilli() - start

		fmt.Println(fmt.Sprintf("Sum: %.08f, count: %v, time: %v", sum, count, delta))
	})

	t.Run("Test Interval group sum", func(t *testing.T) {
		ctx = context.Background()
		// Формат, соответствующий строке
		layout := "2006-01-02 15:04:05.999"
		dateStart, err := time.Parse(layout, "2025-01-17 22:23:57.042000")
		require.NoError(t, err)

		dateEnd := dateStart.Add(60000 * time.Second)

		start := time.Now().UnixMilli()
		labels, sums, err := store.DifficultyIntervalGroupGlobal(ctx, dateStart, dateEnd, 60, 4, "SOLO")
		require.NoError(t, err)
		delta := time.Now().UnixMilli() - start
		fmt.Println(labels)
		fmt.Println(sums)
		fmt.Println(fmt.Sprintf("time: %v", delta))

		start = time.Now().UnixMilli()
		labels, sums, err = store.DifficultyIntervalGroupWallet(ctx, dateStart, dateEnd, 60, 4, 211822, "SOLO")
		require.NoError(t, err)
		delta = time.Now().UnixMilli() - start
		fmt.Println(labels)
		fmt.Println(sums)
		fmt.Println(fmt.Sprintf("time: %v", delta))

		start = time.Now().UnixMilli()
		labels, sums, err = store.DifficultyIntervalGroupWorker(ctx, dateStart, dateEnd, 60, 4, 211822, "SOLO")
		require.NoError(t, err)
		delta = time.Now().UnixMilli() - start
		fmt.Println(labels)
		fmt.Println(sums)
		fmt.Println(fmt.Sprintf("time: %v", delta))

	})

}
