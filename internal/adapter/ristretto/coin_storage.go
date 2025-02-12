package ristretto

import (
	"fmt"

	"github.com/dgraph-io/ristretto"
)

type RistrettoCoinStorage struct {
	cache *ristretto.Cache
}

func NewRistrettoCoinStorage() (*RistrettoCoinStorage, error) {
	cache, err := ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e4,     // Количество счётчиков для элементов
		MaxCost:     1 << 24, // Максимальная стоимость (в байтах)
		BufferItems: 64,      // Количество буферных элементов
	})

	return &RistrettoCoinStorage{
		cache: cache,
	}, err
}

func (c *RistrettoCoinStorage) CreateCoin(key string, value int64) (int64, error) {
	c.cache.Set(key, value, 1)
	c.cache.Wait()

	val, isFound := c.cache.Get(key)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoCoinStorage CreateCoin error: значение не того типа")
	}

	return newID, nil
}

func (c *RistrettoCoinStorage) GetCoinIDByName(coin string) (int64, error) {

	val, isFound := c.cache.Get(coin)
	if !isFound {
		return 0, nil
	}

	newID, ok := val.(int64)
	if !ok {
		return 0, fmt.Errorf("RistrettoCoinStorage GetCoinIDByName error: значение не того типа")
	}

	return newID, nil

}
