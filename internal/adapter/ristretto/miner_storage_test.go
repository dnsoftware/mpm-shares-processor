package ristretto

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/dnsoftware/mpm-shares-processor/entity"
)

func TestRistrettoMinerStorage(t *testing.T) {
	storage, err := NewRistrettoMinerStorage()
	require.NoError(t, err)

	//*** wallet
	wallet := entity.Wallet{
		CoinID:       4,
		Name:         "wallet",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	}

	id, err := storage.GetWalletIDByName(wallet.Name, wallet.CoinID, wallet.RewardMethod)
	require.NoError(t, err)
	require.Equal(t, id, int64(0))

	wallet.ID = 1

	id, err = storage.CreateWallet(wallet)
	require.NoError(t, err)
	require.Equal(t, id, int64(1))

	id, err = storage.GetWalletIDByName(wallet.Name, wallet.CoinID, wallet.RewardMethod)
	require.NoError(t, err)
	require.Equal(t, id, int64(1))

	//*** worker
	worker := entity.Worker{
		CoinID:       4,
		Workerfull:   "workerfull",
		Wallet:       "wallet",
		Worker:       "worker",
		ServerID:     "SERVER",
		IP:           "127.0.0.1",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	}

	id, err = storage.GetWorkerIDByName(worker.Workerfull, worker.CoinID, worker.RewardMethod)
	require.NoError(t, err)
	require.Equal(t, id, int64(0))

	worker.ID = 2

	id, err = storage.CreateWorker(worker)
	require.NoError(t, err)
	require.Equal(t, id, int64(2))

	id, err = storage.GetWorkerIDByName(worker.Workerfull, worker.CoinID, worker.RewardMethod)
	require.NoError(t, err)
	require.Equal(t, id, int64(2))

}
