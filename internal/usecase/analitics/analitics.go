package analitics

import (
	"context"
	"time"

	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
)

type ShareStorage interface {
	CoinHashrate(ctx context.Context, coinSymbol string) (uint64, error)
	MinerHashrate(ctx context.Context, walletID int64) (uint64, error)
	WorkerHashrate(ctx context.Context, workerID int64) (uint64, error)
}

type AnaliticsUsecase struct {
	shareStorage ShareStorage
}

func NewAnaliticsUsecase(s ShareStorage) *AnaliticsUsecase {
	return &AnaliticsUsecase{
		shareStorage: s,
	}
}

func (a *AnaliticsUsecase) CoinHashrate(coinSymbol string) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout*time.Second)
	defer cancel()

	coinHashrate, err := a.shareStorage.CoinHashrate(ctx, coinSymbol)
	if err != nil {
		return 0, err
	}

	return coinHashrate, nil
}

func (a *AnaliticsUsecase) MinerHashrate(walletID int64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout*time.Second)
	defer cancel()

	walletHashrate, err := a.shareStorage.MinerHashrate(ctx, walletID)
	if err != nil {
		return 0, err
	}

	return walletHashrate, nil
}

func (a *AnaliticsUsecase) WorkerHashrate(workerID int64) (uint64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), constants.ContextTimeout*time.Second)
	defer cancel()

	workerHashrate, err := a.shareStorage.WorkerHashrate(ctx, workerID)
	if err != nil {
		return 0, err
	}

	return workerHashrate, nil
}
