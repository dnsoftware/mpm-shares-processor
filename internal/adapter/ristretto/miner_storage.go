package ristretto

import (
	"fmt"

	"github.com/dgraph-io/ristretto"

	"github.com/dnsoftware/mpm-shares-processor/entity"
)

type RistrettoMinerStorage struct {
	cache *ristretto.Cache
}

func NewRistrettoMinerStorage() (*RistrettoMinerStorage, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,     // Количество счётчиков для элементов
		MaxCost:     1 << 30, // Максимальная стоимость (в байтах)
		BufferItems: 64,      // Количество буферных элементов
		//Cost: func(value interface{}) int64 {
		//	if str, ok := value.(string); ok {
		//		return int64(len(str)) // Стоимость — длина строки
		//	}
		//	return 1
		//},
	})

	return &RistrettoMinerStorage{
		cache: cache,
	}, err
}

func (p *RistrettoMinerStorage) makeKey(obj interface{}) (string, error) {

	var key string

	switch v := obj.(type) {
	case entity.Wallet:
		coinIDStr := fmt.Sprintf("%v", v.CoinID)
		key = v.Name + coinIDStr + v.RewardMethod
		return key, nil
	case entity.Worker:
		coinIDStr := fmt.Sprintf("%v", v.CoinID)
		key = v.Workerfull + coinIDStr + v.RewardMethod
		return key, nil
	default: // нет совпадения, v имеет тип значения интерфейса x
		return "", fmt.Errorf("bad cached object")
	}

}

// CreateWallet закэшировать ID кошелька (майнера)
// wallet - должен уже иметь ID
func (p *RistrettoMinerStorage) CreateWallet(wallet entity.Wallet) (int64, error) {
	key, err := p.makeKey(wallet)
	if err != nil {
		return 0, err
	}

	p.cache.Set(key, wallet.ID, 1)
	p.cache.Wait()

	val, isFound := p.cache.Get(key)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoMinerStorage CreateWallet error: значение не того типа")
	}

	return newID, nil
}

// CreateWorker закэшировать ID воркера
// worker - должен уже иметь ID
func (p *RistrettoMinerStorage) CreateWorker(worker entity.Worker) (int64, error) {
	key, err := p.makeKey(worker)
	if err != nil {
		return 0, err
	}

	p.cache.Set(key, worker.ID, 1)
	p.cache.Wait()

	val, isFound := p.cache.Get(key)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoMinerStorage CreateWorker error: значение не того типа")
	}

	return newID, nil
}

func (p *RistrettoMinerStorage) GetWalletIDByName(walletName string, coinID int64, rewardMethod string) (int64, error) {
	coinIDStr := fmt.Sprintf("%v", coinID)
	key := walletName + coinIDStr + rewardMethod
	val, isFound := p.cache.Get(key)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoMinerStorage GetWalletIDByName error: значение не того типа")
	}

	return newID, nil
}

func (p *RistrettoMinerStorage) GetWorkerIDByName(workerName string, coinID int64, rewardMethod string) (int64, error) {
	coinIDStr := fmt.Sprintf("%v", coinID)
	key := workerName + coinIDStr + rewardMethod
	val, isFound := p.cache.Get(key)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoMinerStorage GetWorkerIDByName error: значение не того типа")
	}

	return newID, nil
}
