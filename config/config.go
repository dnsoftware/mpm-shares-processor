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
	Name    string `yaml:"app_name" envconfig:"APP_NAME"    required:"true"`
	Version string `yaml:"app_version" envconfig:"APP_VERSION" required:"true"`
}

type KafkaShareReaderConfig struct {
	Brokers            []string      `yaml:"brokers" envconfig:"KAFKA_SHARE_READER_BROKERS" required:"true"`
	Group              string        `yaml:"group" envconfig:"KAFKA_SHARE_READER_GROUP" required:"true"`
	Topic              string        `yaml:"topic" envconfig:"KAFKA_SHARE_READER_TOPIC" required:"true"`
	AutoCommitEnable   bool          `yaml:"auto_commit_enable" envconfig:"KAFKA_SHARE_AUTO_COMMIT_ENABLE" required:"true"`
	AutoCommitInterval int           `yaml:"auto_commit_interval" envconfig:"KAFKA_SHARE_AUTO_COMMIT_INTERVAL" required:"true"` // в секундах
	ReadBatchSize      int           `yaml:"read_batch_size"`
	ReadFlushInterval  time.Duration `yaml:"read_flush_interval"`
}

type KafkaMetricWriterConfig struct {
	Brokers []string `yaml:"brokers" envconfig:"KAFKA_METRIC_WRITER_BROKERS" required:"true"`
	Topic   string   `yaml:"topic" envconfig:"KAFKA_METRIC_WRITER_TOPIC" required:"true"`
}

type GRPCConfig struct {
	CoinTarget   string `yaml:"coin_target" envconfig:"GRPC_COIN_TARGET" required:"true"`     // хост:порт удаленного хранилища Coin
	MinerTarget  string `yaml:"miner_target" envconfig:"GRPC_MINER_TARGET" required:"true"`   // хост:порт удаленного хранилища майнер/воркер
	SharesTarget string `yaml:"shares_target" envconfig:"GRPC_SHARES_TARGET" required:"true"` // хост:порт удаленного хранилища shares timeseries
}

type AuthConfig struct {
	JWTServiceName   string   `yaml:"jwt_service_name" envconfig:"JWT_SERVICE_NAME" required:"true"`          // Название сервиса (для сверки с JWTValidServices при авторизаии)
	JWTSecret        string   `yaml:"jwt_secret" envconfig:"AUTH_JWT_SECRET" required:"true"`                 // JWT секрет
	JWTValidServices []string `yaml:"jwt_valid_services" envconfig:"AUTH_JWT_VALID_SERVICES" required:"true"` // список микросервисов (через запятую), которым разрешен доступ
}

type OtelConfig struct {
	Endpoint           string        `yaml:"endpoint"`
	BatchTimeout       time.Duration `yaml:"batch_timeout"`         // таймоут отправки телеметрических пакетов в секундах
	MaxExportBatchSize int           `yaml:"max_export_batch_size"` // максимальное кол-во сообщений в пакете
	MaxQueueSize       int           `yaml:"max_queue_size"`        // максимум спанов в очереди
}

type Config struct {
	App               App                     `yaml:"application"`
	KafkaShareReader  KafkaShareReaderConfig  `yaml:"kafka_share_reader"`
	KafkaMetricWriter KafkaMetricWriterConfig `yaml:"kafka_metric_writer"`
	GRPC              GRPCConfig              `yaml:"grpc"`
	Auth              AuthConfig              `yaml:"auth"`
	Otel              OtelConfig              `yaml:"otel"`
}

func New(filePath string, envFile string) (Config, error) {
	var config Config
	var err error

	// 1. Читаем из config.yaml. Самый низкий приоритет
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
