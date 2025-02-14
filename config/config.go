package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v2"
)

type App struct {
	AppID                string
	Name                 string            `yaml:"app_name" envconfig:"APP_NAME"    required:"false"`
	Version              string            `yaml:"app_version" envconfig:"APP_VERSION" required:"false"`
	Dependencies         []string          `yaml:"dependencies"` // Зависимости от других микросервисов (будет ожидать их запуска, отслеживание через Service Discovery)
	ServiceDiscoveryList map[string]string // список текущих сервисов из Service Discovery
}

type ApiBaseUrls struct {
	Grps string `yaml:"grpc"` // host:port для доступа к gRPC API (пустая строка, если нет)
	Rest string `yaml:"rest"` // host:port для доступа к REST API (пустая строка, если нет)
}

type KafkaShareReaderConfig struct {
	Brokers            []string      `yaml:"brokers" envconfig:"KAFKA_SHARE_READER_BROKERS" required:"false"`
	Group              string        `yaml:"group" envconfig:"KAFKA_SHARE_READER_GROUP" required:"false"`
	Topic              string        `yaml:"topic" envconfig:"KAFKA_SHARE_READER_TOPIC" required:"false"`
	AutoCommitEnable   bool          `yaml:"auto_commit_enable" envconfig:"KAFKA_SHARE_AUTO_COMMIT_ENABLE" required:"false"`
	AutoCommitInterval int           `yaml:"auto_commit_interval" envconfig:"KAFKA_SHARE_AUTO_COMMIT_INTERVAL" required:"false"` // в секундах
	ReadBatchSize      int           `yaml:"read_batch_size"`
	ReadFlushInterval  time.Duration `yaml:"read_flush_interval"`
}

type KafkaMetricWriterConfig struct {
	Brokers []string `yaml:"brokers" envconfig:"KAFKA_METRIC_WRITER_BROKERS" required:"false"`
	Topic   string   `yaml:"topic" envconfig:"KAFKA_METRIC_WRITER_TOPIC" required:"false"`
}

type GRPCConfig struct {
	CoinTarget  string `yaml:"coin_target" envconfig:"GRPC_COIN_TARGET" required:"false"`   // ServiceDiscovery ID для адреса сервиса справочника монет
	MinerTarget string `yaml:"miner_target" envconfig:"GRPC_MINER_TARGET" required:"false"` // ServiceDiscovery ID для адреса сервиса работы с майнерами/воркерами
	// Deprecated
	SharesTarget string `yaml:"shares_target" envconfig:"GRPC_SHARES_TARGET" required:"false"` // хост:порт удаленного хранилища shares timeseries
}

type AuthConfig struct {
	JWTServiceName   string   `yaml:"jwt_service_name" envconfig:"JWT_SERVICE_NAME" required:"false"`          // Название сервиса (для сверки с JWTValidServices при авторизаии)
	JWTSecret        string   `yaml:"jwt_secret" envconfig:"AUTH_JWT_SECRET" required:"false"`                 // JWT секрет
	JWTValidServices []string `yaml:"jwt_valid_services" envconfig:"AUTH_JWT_VALID_SERVICES" required:"false"` // список микросервисов (через запятую), которым разрешен доступ
}

type OtelConfig struct {
	Endpoint           string        `yaml:"endpoint"`
	BatchTimeout       time.Duration `yaml:"batch_timeout"`         // таймоут отправки телеметрических пакетов в секундах
	MaxExportBatchSize int           `yaml:"max_export_batch_size"` // максимальное кол-во сообщений в пакете
	MaxQueueSize       int           `yaml:"max_queue_size"`        // максимум спанов в очереди
}

type ClickhouseConfig struct {
	Addr     []string `yaml:"addr" envconfig:"CLICKHOUSE_ADDR" required:"false"`         // хост:порт clickhouse
	Database string   `yaml:"database" envconfig:"CLICKHOUSE_DATABASE" required:"false"` // название базы clickhouse
	Username string   `yaml:"username" envconfig:"CLICKHOUSE_USERNAME" required:"false"` // имя пользователя базы clickhouse
	Password string   `yaml:"password" envconfig:"CLICKHOUSE_PASSWORD" required:"false"` // пароль пользователя базы clickhouse
}

type Etcd struct {
	Endpoints string
	Username  string
	Password  string
}

type Config struct {
	App               App         `yaml:"application"`
	ApiBaseUrls       ApiBaseUrls `yaml:"api_base_urls"`
	EtcdConfig        Etcd
	KafkaShareReader  KafkaShareReaderConfig  `yaml:"kafka_share_reader"`
	KafkaMetricWriter KafkaMetricWriterConfig `yaml:"kafka_metric_writer"`
	GRPC              GRPCConfig              `yaml:"grpc"`
	Auth              AuthConfig              `yaml:"auth"`
	Otel              OtelConfig              `yaml:"otel"`
	Clickhouse        ClickhouseConfig        `yaml:"clickhouse"`
}

func New(filePath string, envFile string) (Config, error) {
	var config Config
	var err error

	config.App.ServiceDiscoveryList = make(map[string]string)

	// 1. Читаем из config.yaml.
	file, err := os.Open(filePath)
	if err == nil {
		defer file.Close()
		decoder := yaml.NewDecoder(file)
		if decodeErr := decoder.Decode(&config); decodeErr != nil {
			return config, fmt.Errorf("Ошибка при чтении config.yaml: %v", decodeErr)
		}
	} else {
		log.Printf("config.yaml не найден, используются значения по умолчанию: %v", err)
	}

	// 2.1 Загрузка переменных окружения из .env
	err = godotenv.Load(envFile)
	if err != nil {
		return config, fmt.Errorf("godotenv.Load: %w", err)
	}

	// 2.2 Переопределяем переменные, полученные из конфиг файла
	err = envconfig.Process("", &config)
	if err != nil {
		return config, fmt.Errorf("envconfig.Process: %w", err)
	}

	// 3. Чтение параметров командной строки
	// Регистрируем флаги
	kafkaShareBrokers := flag.String("kafka_share_brokers", "", "Хост для подключения")
	jwtSecret := flag.String("js", "", "JWT secret")

	// Устанавливаем тестовые аргументы (TODO убрать)
	flag.CommandLine.Parse([]string{"-kafka_share_brokers=localhost:9092", "-js=jwtsecret"})

	flag.Parse()

	// для каждого аргумента проверяем не пустой ли он, и если не пустой - переопределяем переменную конфига
	if *kafkaShareBrokers != "" {
		config.KafkaShareReader.Brokers = strings.Split(*kafkaShareBrokers, ",")
		config.Auth.JWTSecret = *jwtSecret
	}

	return config, nil
}
