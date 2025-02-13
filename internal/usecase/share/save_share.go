package share

import (
	"context"
	"time"

	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
)

// AddSharesBatch сохранение шары в базе данных (ClickHouse)
// Возвращает nil, если запись была добавлена успешно
func (u *ShareUseCase) AddSharesBatch(shares []entity.Share) error {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := u.shareStorage.AddSharesBatch(ctx, shares)

	return err
}
