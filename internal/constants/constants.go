package constants

const (
	ProjectRootAnchorFile = ".env"
	AppLogFile            = "app.log"
	TestLogFile           = "test.log"
	StartConfigFilename   = "/startconf.yaml"             // название файла стартового конфига (с доступами к etcd основного конфига)
	LocalConfigPath       = "/config.yaml"                // Путь к локальному файлу конфига (сюда сохраняется удаленный конфиг)
	ServiceDiscoveryPath  = "/service_discovery/services" // Папка в etcd где хранятся конфиги микросервисов

	CaPath      = "/certs/ca.crt"     // путь к корневому сертификату
	PublicPath  = "/certs/client.crt" // путь к сертификату
	PrivatePath = "/certs/client.key" // путь к приватному ключу

)

// Работа с шарами
const (
	KafkaSharesGroup              = "sharesGroup"
	KafkaSharesTopic              = "shares"
	KafkaSharesAutocommitInterval = 5
)

// Postgresql
const (
	QueryDealine = 5 // время в секундах, после которого прерывать контекст выполнения Postgresql запроса
)

const WorkerSeparator = "."      // символ разделитель имени воркера от имени кошелька
const MigrationDir = "migration" // папка с миграциями относительно корня проекта
const ClickhouseCluster = "clickhouse_cluster"
