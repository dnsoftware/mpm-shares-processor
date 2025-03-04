package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/dnsoftware/mpm-miners-processor/pkg/certmanager"
	jwtauth "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
	"github.com/golang-migrate/migrate/v4"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/dnsoftware/mpmslib/pkg/servicediscovery"

	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/rest"
	"github.com/dnsoftware/mpm-shares-processor/internal/usecase/analitics"
	"github.com/dnsoftware/mpm-shares-processor/pkg/kafka_reader"
	"github.com/dnsoftware/mpm-shares-processor/pkg/logger"
	otelpkg "github.com/dnsoftware/mpm-shares-processor/pkg/otel"
	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	"github.com/dnsoftware/mpm-shares-processor/config"

	_ "github.com/golang-migrate/migrate/v4/database/clickhouse"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	pb "github.com/dnsoftware/mpm-shares-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/kafka_consumer/shares"
	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/ristretto"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	clickhouse2 "github.com/dnsoftware/mpm-shares-processor/internal/infrastructure/clickhouse"
	"github.com/dnsoftware/mpm-shares-processor/internal/usecase/share"
)

type Dependencies struct {
}

func Run(ctx context.Context, cfg config.Config) (err error) {
	var deps Dependencies
	_ = deps

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed (app.Run): %s", err.Error())
	}

	etcdConf, err := servicediscovery.NewEtcdConfig(servicediscovery.EtcdConfig{
		Nodes:       strings.Split(cfg.EtcdConfig.Endpoints, ","),
		Username:    cfg.EtcdConfig.Username,
		Password:    cfg.EtcdConfig.Password,
		CertCaPath:  basePath + constants.CaPath,
		CertPath:    basePath + constants.PublicPath,
		CertKeyPath: basePath + constants.PrivatePath,
	})
	if err != nil {
		log.Fatalf("NewEtcdConfig error: %s", err.Error())
	}

	serviceKey := cfg.App.AppID + ":" + constants.ApiBaseUrlGrpc
	serviceAddr := cfg.ApiBaseUrls.Grps
	sd, err := servicediscovery.NewServiceDiscovery(*etcdConf, constants.ServiceDiscoveryPath, serviceKey, serviceAddr, 5, 10)
	if err != nil {
		log.Fatalf("NewServiceDiscovery error: %s", err.Error())
	}

	err = sd.RegisterService(cfg.App.AppID+":"+constants.ApiBaseUrlRest, cfg.ApiBaseUrls.Rest)
	if err != nil {
		log.Fatalf("Rest service register error: %s", err.Error())
	}

	sd.WaitDependencies(cfg.App.Dependencies)
	cfg.App.ServiceDiscoveryList, err = sd.DiscoverAllServices()
	if err != nil {
		log.Fatalf("DiscoverAllServices error: %s", err.Error())
	}
	logger.Log().Info("All services discovered")

	// инициализируем BaseURLs доступа к API внешних сервисов
	cfg.GRPC.CoinTarget = cfg.App.ServiceDiscoveryList[cfg.GRPC.CoinTarget]
	cfg.GRPC.MinerTarget = cfg.App.ServiceDiscoveryList[cfg.GRPC.MinerTarget]

	// Инициализация трассировщика
	otelConfig := otelpkg.Config{
		ServiceName:        cfg.App.Name,
		CollectorEndpoint:  cfg.Otel.Endpoint,
		BatchTimeout:       cfg.Otel.BatchTimeout * time.Second,
		MaxExportBatchSize: cfg.Otel.MaxExportBatchSize,
		MaxQueueSize:       cfg.Otel.MaxQueueSize,
	}
	_ = otelpkg.InitTracer(otelConfig)
	//defer cleanup()
	tracer := otel.Tracer("share-trace")
	_ = tracer

	// Для работы с JWT токенами
	jwt := jwtauth.NewJWTServiceSymmetric(cfg.Auth.JWTServiceName, cfg.Auth.JWTValidServices, cfg.Auth.JWTSecret, 60)

	// Полномочия для TLS соединения
	certMan, err := certmanager.NewCertManager(basePath + "/certs")
	clientCreds, err := certMan.GetClientCredentials()

	// Клиентское GRPC соединение
	conn, err := grpc.DialContext(ctx,
		cfg.GRPC.CoinTarget, // Адрес:порт
		//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
		grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
		grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	)
	if err != nil {
		logger.Log().Fatal("DialContext error: " + err.Error())
	}

	// Ristretto кэш
	cacheCoin, err := ristretto.NewRistrettoCoinStorage()
	if err != nil {
		logger.Log().Fatal("NewRistrettoCoinStorage error: " + err.Error())
	}
	cacheMiner, err := ristretto.NewRistrettoMinerStorage()
	if err != nil {
		logger.Log().Fatal("NewRistrettoMinerStorage error: " + err.Error())
	}

	// remote API miners processor
	coinStorage, err := pb.NewCoinStorage(conn)
	if err != nil {
		logger.Log().Fatal("NewCoinStorage error: " + err.Error())
	}

	minerStorage, err := pb.NewMinerStorage(conn)
	if err != nil {
		logger.Log().Fatal("NewMinerStorage error: " + err.Error())
	}

	////***** Remote ClickHouse shares timeseries
	//connShares, err := grpc.DialContext(ctx,
	//	cfg.GRPC.SharesTarget, // Адрес:порт
	//	//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
	//	grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
	//	grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	//)
	//if err != nil {
	//	logger.Log().Fatal("Remote ClickHouse DialContext error: " + err.Error())
	//
	//}

	// Подключение к базе данных ClickHouse
	connCH, err := clickhouse.Open(&clickhouse.Options{
		Addr: cfg.Clickhouse.Addr,
		Auth: clickhouse.Auth{
			Database: "default",
			Username: cfg.Clickhouse.Username,
			Password: cfg.Clickhouse.Password,
		},
		Settings: clickhouse.Settings{
			"max_execution_time": 60,
		},
		Debug: true,
	})
	if err != nil {
		logger.Log().Fatal(err.Error())
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Проверка подключения
	err = connCH.Ping(ctx)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}

	// Укажите путь к миграциям и строку подключения к базе данных
	dsn := fmt.Sprintf("clickhouse://%s:%s@%s/%s", "default", "", cfg.Clickhouse.Addr[0], "default")
	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir,
		dsn,
	)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}

	// Сброс миграций
	m.Force(-1)

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		logger.Log().Fatal(err.Error())
	}
	log.Println("Миграции успешно применены")

	connCH.Close()

	// Подключаемся к mpmhouse
	connCH, err = clickhouse2.NewClickhouseConnect(clickhouse2.Config{
		Addr:             cfg.Clickhouse.Addr,
		Database:         cfg.Clickhouse.Database,
		Username:         cfg.Clickhouse.Username,
		Password:         cfg.Clickhouse.Password,
		MaxExecutionTime: 10,
	})
	if err != nil {
		logger.Log().Fatal(err.Error())
	}
	defer connCH.Close()

	// Проверка подключения
	err = connCH.Ping(ctx)
	if err != nil {
		logger.Log().Fatal(err.Error())
	}

	cfgStore := clickhouse2.ShareStorageConfig{
		Conn:        connCH,
		ClusterName: "clickhouse_cluster",
		Database:    "mpmhouse",
	}
	shareStorage, err := clickhouse2.NewClickhouseShareStorage(cfgStore)
	if err != nil {
		logger.Log().Fatal("NewShareStorage error: " + err.Error())
	}

	usecase := share.NewShareUseCase(shareStorage, minerStorage, coinStorage, cacheMiner, cacheCoin)
	analiticsUsecase := analitics.NewAnaliticsUsecase(shareStorage)

	// http сервер
	httpHandler := rest.NewHandler(analiticsUsecase)
	go func() {
		http.ListenAndServe(cfg.ApiBaseUrls.Rest, httpHandler.Routes())
	}()

	// Вычитываем сообщения из Кафки
	cfgReader := kafka_reader.Config{
		Brokers:            cfg.KafkaShareReader.Brokers,
		Group:              cfg.KafkaShareReader.Group,
		Topic:              cfg.KafkaShareReader.Topic,
		AutoCommitInterval: constants.KafkaSharesAutocommitInterval,
		AutoCommitEnable:   true,
	}

	reader, err := kafka_reader.NewKafkaReader(cfgReader, logger.Log())

	cfgConsumer := shares.Config{
		BatchSize:     cfg.KafkaShareReader.ReadBatchSize,
		FlushInterval: cfg.KafkaShareReader.ReadFlushInterval,
	}

	consumer, err := shares.NewShareConsumer(cfgConsumer, reader, usecase)
	if err != nil {
		logger.Log().Fatal("NewShareConsumer error: " + err.Error())
	}
	defer consumer.Close()

	// Стартуем вычитывание сообщений
	consumer.StartConsume()

	// Настройка graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	<-quit
	log.Println("Shutting down service...")

	return nil
}
