CREATE TABLE IF NOT EXISTS mpmhouse.shares  ON CLUSTER clickhouse_cluster (
   uuid String, -- уникальный идентификатор
   server_id String, -- идентификатор пул-сервера
   coin_id Int64, -- идентификатор монеты
   worker_id Int64, -- ID воркера
   wallet_id Int64, -- ID майнера (кошелька)
   share_date DateTime64(3), -- время когда найдено в формате timestamp, в миллисекундах
   difficulty Decimal(25, 10), -- сложность майнера
   sharedif Decimal(25, 10), -- сложность шары реальная
   nonce String, -- nonce шары
   is_solo Bool, -- соло режим
   reward_method String, -- метод начисления вознаграждения
   cost Decimal(30, 20), -- награда за шару
   INDEX idx_share_date (share_date) TYPE minmax GRANULARITY 4,
   INDEX idx_server_id (server_id) TYPE set(4) GRANULARITY 4,
   INDEX idx_coin_id (coin_id) TYPE set(4) GRANULARITY 4,
   INDEX idx_worker_id (worker_id) TYPE set(0) GRANULARITY 4,
   INDEX idx_wallet_id (wallet_id) TYPE set(0) GRANULARITY 4,
   INDEX idx_reward_method (reward_method) TYPE set(4) GRANULARITY 4,
   INDEX idx_uuid (uuid) TYPE minmax GRANULARITY 16
)
ENGINE = ReplicatedMergeTree(
    '/clickhouse/tables/{shard}/shares', -- Общий путь для всех реплик
    '{replica}'                           -- Уникальное имя для текущей реплики
)
PARTITION BY toYYYYMM(share_date)
ORDER BY (share_date)
SETTINGS index_granularity = 8192;
