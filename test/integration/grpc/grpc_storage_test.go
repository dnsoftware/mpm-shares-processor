package grpc

import (
	"context"
	"log"
	"testing"
	"time"

	"github.com/dnsoftware/mpm-miners-processor/pkg/certmanager"
	jwtauth "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	"github.com/dnsoftware/mpm-shares-processor/config"

	pb "github.com/dnsoftware/mpm-shares-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
)

// Должен быть запущен сторонний микросервис к которому идут запросы от тестируемого клиента
func TestGRPCStorageTest(t *testing.T) {

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env_example"

	cfg, err := config.New(configFile, envFile)

	jwt := jwtauth.NewJWTServiceSymmetric(cfg.Auth.JWTServiceName, cfg.Auth.JWTValidServices, cfg.Auth.JWTSecret, 60)

	// Полномочия для TLS соединения
	certMan, err := certmanager.NewCertManager(basePath + "/certs")
	clientCreds, err := certMan.GetClientCredentials()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	conn, err := grpc.DialContext(ctx,
		cfg.GRPC.CoinTarget, // Адрес:порт
		//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
		grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
		grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	)
	if err != nil {
		t.Fatalf("Failed to create gRPC client: %v", err)
	}

	// Coin
	storage, err := pb.NewCoinStorage(conn)
	require.NoError(t, err)

	ctx = context.Background()
	id, err := storage.GetCoinIDByName(ctx, "ALPH")
	require.NoError(t, err)
	require.Equal(t, int64(4), id)

	// Miner
	stor, err := pb.NewMinerStorage(conn)
	require.NoError(t, err)
	newID, err := stor.CreateWallet(ctx, entity.Wallet{
		CoinID:       4,
		Name:         "wallettest",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)

	addedID, err := stor.GetWalletIDByName(ctx, "wallettest", 4, "PPLNS")
	require.NoError(t, err)
	require.Equal(t, newID, addedID)

	// Проверка на повторную вставку
	checkID, err := stor.CreateWallet(ctx, entity.Wallet{
		CoinID:       4,
		Name:         "wallettest",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, newID, checkID)

	// Worker
	newID, err = stor.CreateWorker(ctx, entity.Worker{
		CoinID:       4,
		Workerfull:   "workerfull.test",
		Wallet:       "workerfull",
		Worker:       "test",
		ServerID:     "SERV",
		IP:           "127.0.0.1",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)

	addedID, err = stor.GetWorkerIDByName(ctx, "workerfull.test", 4, "PPLNS")
	require.NoError(t, err)
	require.Equal(t, newID, addedID)

	// Проверка на повторную вставку
	checkID, err = stor.CreateWorker(ctx, entity.Worker{
		CoinID:       4,
		Workerfull:   "workerfull.test",
		Wallet:       "workerfull",
		Worker:       "test",
		ServerID:     "SERV",
		IP:           "127.0.0.1",
		IsSolo:       false,
		RewardMethod: "PPLNS",
	})
	require.NoError(t, err)
	require.Equal(t, newID, checkID)

}
