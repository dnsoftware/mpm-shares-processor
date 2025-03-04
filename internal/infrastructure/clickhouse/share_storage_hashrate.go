package clickhouse

import (
	"context"
	"crypto/rand"
	"encoding/binary"
)

func (c *ClickhouseShareStorage) CoinHashrate(ctx context.Context, coinSymbol string) (uint64, error) {
	var num uint64
	binary.Read(rand.Reader, binary.LittleEndian, &num)
	num = num % 65536

	return num, nil
}

func (c *ClickhouseShareStorage) MinerHashrate(ctx context.Context, walletID int64) (uint64, error) {
	var num uint64
	binary.Read(rand.Reader, binary.LittleEndian, &num)

	return num, nil
}

func (c *ClickhouseShareStorage) WorkerHashrate(ctx context.Context, workerID int64) (uint64, error) {
	var num uint64
	binary.Read(rand.Reader, binary.LittleEndian, &num)
	num = num % 65536

	return num, nil
}
