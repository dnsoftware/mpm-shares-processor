package postgres

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
)

type PostgresCoinStorage struct {
	pool *pgxpool.Pool
}

func NewPostgresCoinStorage(pool *pgxpool.Pool) (*PostgresCoinStorage, error) {

	storage := &PostgresCoinStorage{
		pool: pool,
	}

	return storage, nil
}

func (c *PostgresCoinStorage) GetCoinIDByName(ctx context.Context, coin string) (int64, error) {
	var id int64
	err := c.pool.QueryRow(ctx, `SELECT id FROM coins WHERE symbol = $1`,
		coin).Scan(&id)

	return id, err

}
