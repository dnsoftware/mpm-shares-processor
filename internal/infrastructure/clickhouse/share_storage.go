package clickhouse

import (
	"bytes"
	"context"
	"database/sql"
	"fmt"
	"log"
	"strconv"
	"strings"
	"text/template"
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

// RangeDifficultySum сумма полей difficulty из диапазона dateStart < share_date <= dateEnd
// dateStart - дата нахождения предыдущего блока
// dateEnd - дата нахождения текущего блока
// Возвращает сумму м кол-во шар в суммируемом наборе
func (c *ClickhouseShareStorage) RangeDifficultySum(ctx context.Context,
	dateStart time.Time, dateEnd time.Time, coinID int64, rewardMethod string) (float64, uint64, error) {
	query := `SELECT COUNT() cnt, CAST(SUM(difficulty), 'String') AS sumdif 
			  FROM shares WHERE share_date > ? AND share_date <= ? AND coin_id = ? AND reward_method = ? `

	var cnt uint64
	var sumdifStr string
	err := c.conn.QueryRow(ctx, query, dateStart, dateEnd, coinID, rewardMethod).Scan(&cnt, &sumdifStr)
	if err != nil {
		return 0, 0, err
	}

	sumdif, err := strconv.ParseFloat(sumdifStr, 64)
	if err != nil {
		return 0, 0, err
	}

	return sumdif, cnt, err
}

// DifficultyIntervalGroupGlobal суммирование сложностей по интервалам в рамках указанного периода для всех майнеров по монете и reward методу
func (c *ClickhouseShareStorage) DifficultyIntervalGroupGlobal(ctx context.Context,
	periodStart time.Time,
	periodEnd time.Time,
	interval int,
	coinID int64,
	rewardMethod string,
) ([]time.Time, []float64, error) {
	return c.difficultyIntervalGroup(ctx, periodStart, periodEnd, interval, coinID, 0, 0, rewardMethod)
}

// DifficultyIntervalGroupWallet суммирование сложностей по интервалам в рамках указанного периода для конкретного манера по монете и reward методу
func (c *ClickhouseShareStorage) DifficultyIntervalGroupWallet(ctx context.Context,
	periodStart time.Time,
	periodEnd time.Time,
	interval int,
	coinID int64,
	walletID int64,
	rewardMethod string,
) ([]time.Time, []float64, error) {
	return c.difficultyIntervalGroup(ctx, periodStart, periodEnd, interval, coinID, walletID, 0, rewardMethod)
}

// DifficultyIntervalGroupWorker суммирование сложностей по интервалам в рамках указанного периода для конкретного воркера по монете и reward методу
func (c *ClickhouseShareStorage) DifficultyIntervalGroupWorker(ctx context.Context,
	periodStart time.Time,
	periodEnd time.Time,
	interval int,
	coinID int64,
	workerID int64,
	rewardMethod string,
) ([]time.Time, []float64, error) {
	return c.difficultyIntervalGroup(ctx, periodStart, periodEnd, interval, coinID, 0, workerID, rewardMethod)
}

// difficultyIntervalGroup суммирование сложностей по интервалам в рамках указанного периода
// Interval - (start, end] записи с меткой start не включаются, а записи с меткой end включаются в результат
func (c *ClickhouseShareStorage) difficultyIntervalGroup(ctx context.Context,
	periodStart time.Time, // дата начала глобального интервала
	periodEnd time.Time, // дата окончания глобального интервала
	interval int, // интервал группировки в секундах
	coinID int64, // код монеты
	walletID int64, // код кошелька (майнера), 0 - если не учитываем
	workerID int64, // код воркера, 0 - если не учитываем
	rewardMethod string, // метод вознаграждения
) ([]time.Time, []float64, error) {

	var timeLabels []time.Time
	var difSums []float64

	queryTemplate := `SELECT
    			toStartOfInterval(share_date, INTERVAL ? SECOND) AS interval_end,
    			COUNT() cnt, toFloat64(SUM(difficulty)) AS sumdif
			  FROM shares WHERE share_date >= ? AND share_date < ? AND coin_id = ? AND reward_method = ? {{.where}}
			  GROUP BY
    			interval_end
			  ORDER BY
    			interval_end `

	var parts []string                       // динамические части SQL запроса
	var dynamicParams []any                  // динамические параметры в SQL запросе (подставляются вместо знака "?")
	sqlSubstrings := make(map[string]string) // динамические подстроки для SQL запроса

	dynamicParams = append(dynamicParams, interval, periodStart, periodEnd, coinID, rewardMethod)

	if walletID > 0 {
		parts = append(parts, " AND wallet_id = ? ")
		dynamicParams = append(dynamicParams, walletID)
	}

	if workerID > 0 {
		parts = append(parts, " AND worker_id = ? ")
		dynamicParams = append(dynamicParams, workerID)
	}

	sqlSubstrings["where"] = strings.Join(parts, " ")

	// Создание шаблона
	tmpl, err := template.New("query").Parse(queryTemplate)
	if err != nil {
		return nil, nil, err
	}

	var query bytes.Buffer
	err = tmpl.Execute(&query, sqlSubstrings)
	if err != nil {
		log.Fatal(err)
	}

	rows, err := c.conn.Query(ctx, query.String(), dynamicParams...)
	if err != nil {
		return nil, nil, err
	}
	defer rows.Close()

	// Обработка результатов
	for rows.Next() {
		var intervalStart time.Time
		var cnt uint64
		var sumdif float64
		if err := rows.Scan(&intervalStart, &cnt, &sumdif); err != nil {
			return nil, nil, err
		}

		timeLabels = append(timeLabels, intervalStart)
		difSums = append(difSums, sumdif)

	}

	// Проверка на ошибки после выполнения
	if err := rows.Err(); err != nil {
		return nil, nil, err
	}

	return timeLabels, difSums, nil
}
