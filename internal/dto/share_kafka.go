package dto

import (
	"github.com/dnsoftware/mpm-shares-processor/entity"
)

// ShareFound Структура данных шары получаемой из Кафки
type ShareFound struct {
	Uuid         string `json:"uuid"`         // уникальный идентификатор
	BlockType    string `json:"blockType"`    //
	ServerID     string `json:"serverId"`     // идентификатор пул-сервера (типа ALEPH-1 и т.п.)
	CoinSymbol   string `json:"coinSymbol"`   // идентификатор монеты
	Workerfull   string `json:"workerfull"`   // полный идентификатор воркера
	ShareDate    int64  `json:"shareDate"`    // время когда найдено, в миллисекундах
	CHrate       int64  `json:"cHrate"`       // текущий хешрейт
	AHrate       int64  `json:"aHrate"`       // средний хешрейт
	Difficulty   string `json:"difficulty"`   // сложность майнера
	Sharedif     string `json:"sharedif"`     // сложность шары	реальная
	Nonce        string `json:"nonce"`        // nonce шары
	MinerIp      string `json:"minerIp"`      // IP майнера, приславшего шару
	IsSolo       bool   `json:"isSolo"`       // соло режим
	RewardMethod string `json:"rewardMethod"` // метод начисления вознаграждения
	Cost         string `json:"cost"`         // награда за шару
}

// ToShare Частичный маппинг в entity.Share
// CoinID, WorkerID, WalletID - заполняются потом
func (s *ShareFound) ToShare() entity.Share {

	share := entity.Share{
		UUID:         s.Uuid,
		ServerID:     s.ServerID,
		CoinID:       0, // заполняется в usecase
		WorkerID:     0, // заполняется в usecase
		WalletID:     0, // заполняется в usecase
		ShareDate:    s.ShareDate,
		Difficulty:   s.Difficulty,
		Sharedif:     s.Sharedif,
		Nonce:        s.Nonce,
		IsSolo:       s.IsSolo,
		RewardMethod: s.RewardMethod,
		Cost:         s.Cost,
	}

	return share
}
