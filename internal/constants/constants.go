package constants

const (
	ProjectRootAnchorFile = ".env"
	AppLogFile            = "app.log"
	TestLogFile           = "test.log"
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
