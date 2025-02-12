package entity

type Worker struct {
	ID           int64
	CoinID       int64
	Workerfull   string // полное имя воркера
	Wallet       string // имя кошелька (майнера)
	Worker       string // имя воркера (без имени кошелька)
	ServerID     string // идентификатор пул-сервера (типа ALEPH-1 и т.п.)
	IP           string // IP адрес воркера
	IsSolo       bool   // оставлено для совместимости TODO убрать
	RewardMethod string // строковый код метода распределения наград
}
