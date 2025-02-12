package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/dnsoftware/mpm-miners-processor/pkg/certmanager"
	jwtauth "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-save-get-shares/config"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/kafka_reader"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/logger"
	otelpkg "github.com/dnsoftware/mpm-save-get-shares/pkg/otel"
	"github.com/dnsoftware/mpm-save-get-shares/pkg/utils"
	pb "github.com/dnsoftware/mpm-shares-processor/adapter/grpc"
	"github.com/dnsoftware/mpm-shares-processor/adapter/kafka_consumer/shares"
	"github.com/dnsoftware/mpm-shares-processor/adapter/ristretto"
	"github.com/dnsoftware/mpm-shares-processor/constants"
	"github.com/dnsoftware/mpm-shares-processor/usecase/share"
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

	//***** Remote ClickHouse shares timeseries
	connShares, err := grpc.DialContext(ctx,
		cfg.GRPC.SharesTarget, // Адрес:порт
		//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
		grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
		grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	)
	if err != nil {
		logger.Log().Fatal("Remote ClickHouse DialContext error: " + err.Error())
	}

	shareStorage, err := pb.NewShareStorage(connShares)
	if err != nil {
		logger.Log().Fatal("NewShareStorage error: " + err.Error())
	}

	usecase := share.NewShareUseCase(shareStorage, minerStorage, coinStorage, cacheMiner, cacheCoin)

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
