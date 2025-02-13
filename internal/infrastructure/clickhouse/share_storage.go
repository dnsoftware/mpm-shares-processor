package clickhouse

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
	"github.com/shopspring/decimal"

	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
)

type ShareStorageConfig struct {
	Conn        driver.Conn
	ClusterName string
	Database    string
}

type ClickhouseShareStorage struct {
	conn        driver.Conn
	clusterName string
	database    string
}

func NewClickhouseShareStorage(cfg ShareStorageConfig) (*ClickhouseShareStorage, error) {
	s := &ClickhouseShareStorage{
		conn:        cfg.Conn,
		clusterName: cfg.ClusterName,
		database:    cfg.Database,
	}
	return s, nil
}

// AddShare Добавление единичной шары (для теста, в основном коде не используется, используется пакетная вставка)
func (c *ClickhouseShareStorage) AddShare(ctx context.Context, share entity.Share) error {

	query := fmt.Sprintf(`INSERT INTO %s.shares (uuid, server_id, coin_id, worker_id, wallet_id, share_date, difficulty, sharedif, nonce, is_solo, reward_method, cost) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`, c.database)

	//t.Format("2006-01-02 15:04:05.999") share.ShareDate
	if err := c.conn.Exec(ctx, query, share.UUID, share.ServerID, share.CoinID, share.WorkerID, share.WalletID, share.ShareDate, share.Difficulty, share.Sharedif, share.Nonce, share.IsSolo, share.RewardMethod, share.Cost); err != nil {
		return err
	}

	return nil
}

// GetShareRow Получение единичной записи
func (c *ClickhouseShareStorage) GetShareRow(ctx context.Context, shareUUID string) (*entity.Share, error) {

	q := fmt.Sprintf(`SELECT server_id, coin_id, worker_id, wallet_id, share_date, CAST(difficulty, 'String'), CAST(sharedif, 'String'), nonce, is_solo, reward_method, CAST(cost, 'String') FROM %s.shares WHERE uuid = ?`, c.database)
	share := entity.Share{}
	var shareDate time.Time
	err := c.conn.QueryRow(ctx, q, shareUUID).Scan(&share.ServerID, &share.CoinID, &share.WorkerID, &share.WalletID, &shareDate, &share.Difficulty, &share.Sharedif, &share.Nonce, &share.IsSolo, &share.RewardMethod, &share.Cost)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если запись не найдена
			return nil, nil
		} else {
			return nil, err
		}
	}

	share.ShareDate = shareDate.UnixMilli()

	return &share, nil
}

// AddSharesBatch пакетная вставка
func (c *ClickhouseShareStorage) AddSharesBatch(ctx context.Context, shares []entity.Share) error {

	// Открытие пакетной вставки
	batch, err := c.conn.PrepareBatch(ctx, "INSERT INTO shares (uuid, server_id, coin_id, worker_id, wallet_id, share_date, difficulty, sharedif, nonce, is_solo, reward_method, cost) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)")
	if err != nil {
		return err
	}

	// Добавление данных в пакет
	for _, share := range shares {
		difficulty, err := decimal.NewFromString(share.Difficulty)
		if err != nil {
			return err
		}
		sharedif, err := decimal.NewFromString(share.Sharedif)
		if err != nil {
			return err
		}
		cost, err := decimal.NewFromString(share.Cost)
		if err != nil {
			return err
		}

		if err := batch.Append(share.UUID, share.ServerID, share.CoinID, share.WorkerID, share.WalletID, share.ShareDate, difficulty, sharedif, share.Nonce, share.IsSolo, share.RewardMethod, cost); err != nil {
			return err
		}
	}

	// Попытка выполнить пакетную вставку
	if err := batch.Send(); err != nil {
		return err
	}

	return nil
}
