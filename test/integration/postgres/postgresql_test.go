package postgres

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/dnsoftware/mpm-shares-processor/pkg/logger"
	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	tctest "github.com/dnsoftware/mpm-shares-processor/test/testcontainers"

	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/postgres"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
)

func TestPostgresql(t *testing.T) {
	filePath, err := logger.GetLoggerTestLogPath()
	require.NoError(t, err)
	logger.InitLogger(logger.LogLevelDebug, filePath)

	// Подготовка Postgres контейнера
	ctxPG := context.Background()
	postgresContainer, err := tctest.NewPostgresTestcontainer(t)
	if err != nil {
		t.Fatalf(err.Error())
	}

	dsn, err := postgresContainer.ConnectionString(ctxPG, "sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}
	defer pool.Close()

	// Миграции
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	// Укажите путь к миграциям и строку подключения к базе данных
	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir+"/postgresql",
		dsn,
	)
	require.NoError(t, err)

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}
	log.Println("Миграции успешно применены")

	//******** Тестирование

	// Вставка/чтение таблица workers
	t.Run("Test Write/Read workers", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), constants.QueryDealine*time.Second)
		defer cancel()

		minerStorage, err := postgres.NewPostgresMinerStorage(pool)
		require.NoError(t, err)

		// вставка записи
		worker := entity.Worker{
			CoinID:       4,
			Workerfull:   "wallet.worker",
			Wallet:       "wallet",
			Worker:       "worker",
			ServerID:     "TEST-SERVER",
			IP:           "127.0.0.1",
			IsSolo:       false,
			RewardMethod: "PPLNS",
		}

		id, err := minerStorage.CreateWorker(ctx, worker)
		require.NoError(t, err)
		require.Greater(t, id, int64(0))

		// получение записи
		workerID, err := minerStorage.GetWorkerIDByName(ctx, "wallet.worker", 4, "PPLNS")
		require.NoError(t, err)
		require.Equal(t, id, workerID)

		// Повторная попытка вставки того же значения
		id2, err := minerStorage.CreateWorker(ctx, worker)
		require.NoError(t, err)
		require.Equal(t, id, id2)

	})

	// Вставка/чтение таблица wallets
	t.Run("Test Write/Read wallets", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), constants.QueryDealine*time.Second)
		defer cancel()

		minerStorage, err := postgres.NewPostgresMinerStorage(pool)
		require.NoError(t, err)

		// вставка записи
		wallet := entity.Wallet{
			CoinID:       4,
			Name:         "wallet",
			IsSolo:       false,
			RewardMethod: "PPLNS",
		}

		id, err := minerStorage.CreateWallet(ctx, wallet)
		require.NoError(t, err)
		require.Greater(t, id, int64(0))

		// получение записи
		workerID, err := minerStorage.GetWalletIDByName(ctx, "wallet", 4, "PPLNS")
		require.NoError(t, err)
		require.Equal(t, id, workerID)

		// Повторная попытка вставки того же значения
		id2, err := minerStorage.CreateWallet(ctx, wallet)
		require.NoError(t, err)
		require.Equal(t, id, id2)

	})

	t.Run("Test Read coins", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.Background(), constants.QueryDealine*time.Second)
		defer cancel()

		coinStorage, err := postgres.NewPostgresCoinStorage(pool)
		require.NoError(t, err)

		// получение записи
		coinID, err := coinStorage.GetCoinIDByName(ctx, "ALPH")
		require.NoError(t, err)
		require.Equal(t, int64(4), coinID)
	})

}
