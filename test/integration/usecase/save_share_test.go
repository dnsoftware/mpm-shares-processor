package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/dnsoftware/mpm-miners-processor/pkg/certmanager"
	jwtauth "github.com/dnsoftware/mpm-miners-processor/pkg/jwt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/otel"
	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-shares-processor/pkg/logger"
	otelpkg "github.com/dnsoftware/mpm-shares-processor/pkg/otel"
	"github.com/dnsoftware/mpm-shares-processor/pkg/utils"

	"github.com/dnsoftware/mpm-shares-processor/config"
	tctest "github.com/dnsoftware/mpm-shares-processor/test/testcontainers"

	pb "github.com/dnsoftware/mpm-shares-processor/internal/adapter/grpc"
	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/postgres"
	"github.com/dnsoftware/mpm-shares-processor/internal/adapter/ristretto"
	"github.com/dnsoftware/mpm-shares-processor/internal/constants"
	"github.com/dnsoftware/mpm-shares-processor/internal/dto"
	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
	"github.com/dnsoftware/mpm-shares-processor/internal/usecase/share"
)

func setup(t *testing.T) *pgxpool.Pool {
	ctx := context.Background()
	_ = ctx

	filePath, err := logger.GetLoggerTestLogPath()
	require.NoError(t, err)
	logger.InitLogger(logger.LogLevelDebug, filePath)

	otelConfig := otelpkg.Config{
		ServiceName:        "TestService",
		CollectorEndpoint:  "localhost:4317",
		BatchTimeout:       1 * time.Second,
		MaxExportBatchSize: 100,
		MaxQueueSize:       500,
	}
	_ = otelpkg.InitTracer(otelConfig)
	//defer cleanup()
	tracer := otel.Tracer("share-trace")
	_ = tracer

	/********************** Настройка testcontainers ************************/
	// Уровень логирование testcontainers
	testcontainers.Logger = log.New(os.Stderr, ": ", log.LstdFlags)

	//**** POSTGRES ****//
	ctxPG := context.Background()
	postgresContainer, err := tctest.NewPostgresTestcontainer(t)
	if err != nil {
		t.Fatalf(err.Error())
	}

	dsn, err := postgresContainer.ConnectionString(ctxPG, "sslmode=disable")
	if err != nil {
		t.Fatalf(err.Error())
	}

	pool, err := pgxpool.Connect(context.Background(), dsn)
	if err != nil {
		panic("Unable to connect to database: " + err.Error())
	}

	// применяем миграции Postgres
	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	// Укажите путь к миграциям и строку подключения к базе данных
	m, err := migrate.New(
		"file://"+basePath+"/"+constants.MigrationDir+"/postgresql",
		dsn,
	)
	require.NoError(t, err)

	// Применить миграции
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatalf("Ошибка при применении миграций: %v", err)
	}

	return pool
}

// Тестируем работу с локальной базой Postgresql в testcontainers
func TestSaveShareLocal(t *testing.T) {
	pool := setup(t)

	// Ristretto кэш
	cacheCoin, err := ristretto.NewRistrettoCoinStorage()
	require.NoError(t, err)

	cacheMiner, err := ristretto.NewRistrettoMinerStorage()
	require.NoError(t, err)

	// Postgresql
	coinStorage, err := postgres.NewPostgresCoinStorage(pool)
	require.NoError(t, err)

	minerStorage, err := postgres.NewPostgresMinerStorage(pool)
	require.NoError(t, err)

	// ClickHouse
	shareStorage, err := pb.NewShareStorage(&grpc.ClientConn{})
	require.NoError(t, err)

	usecase := share.NewShareUseCase(shareStorage, minerStorage, coinStorage, cacheMiner, cacheCoin)

	// Загружаем тестовые данные
	var sfSlice []dto.ShareFound
	err = json.Unmarshal([]byte(testData), &sfSlice)
	require.NoError(t, err)

	// Тестируем локальный Postgres
	for _, item := range sfSlice {
		ctx := context.Background()
		start := time.Now().UnixMicro()

		normShare, err := usecase.NormalizeShare(ctx, item)
		end := time.Now().UnixMicro() - start
		fmt.Println(fmt.Sprintf("%v", end))
		if err != nil {
			logger.Log().Error(fmt.Sprintf("Normilize share error: %s", err.Error()))
		}
		require.NoError(t, err)

		require.Equal(t, int64(4), normShare.CoinID)

	}

	// Записываем данные через API в ClickHouse (используем буфер), расположенный во внешнем микросервисе

	// Тестируем записанные данные путем их получения и сравнения с отправленными

}

// Тестируем работу с удаленным микросервисом через API
// Используем JWT и TLS
// Удаленный микросервис "Miner processor" должен быть запущен или же запускать его из теста (в контейнере), а после теста тушить
// Удаленный микросервис "Shares timeseries" должен быть запущен или же запускать его из теста (в контейнере), а после теста тушить
func TestSaveShareRemote(t *testing.T) {
	ctx := context.Background()
	_ = ctx

	filePath, err := logger.GetLoggerTestLogPath()
	require.NoError(t, err)
	logger.InitLogger(logger.LogLevelDebug, filePath)

	otelConfig := otelpkg.Config{
		ServiceName:        "TestService",
		CollectorEndpoint:  "localhost:4317",
		BatchTimeout:       1 * time.Second,
		MaxExportBatchSize: 100,
		MaxQueueSize:       500,
	}
	_ = otelpkg.InitTracer(otelConfig)
	//defer cleanup()
	tracer := otel.Tracer("share-trace")
	_ = tracer

	basePath, err := utils.GetProjectRoot(constants.ProjectRootAnchorFile)
	if err != nil {
		log.Fatalf("GetProjectRoot failed: %s", err.Error())
	}
	configFile := basePath + "/config.yaml"
	envFile := basePath + "/.env_example"

	cfg, err := config.New(configFile, envFile)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	jwt := jwtauth.NewJWTServiceSymmetric(cfg.Auth.JWTServiceName, cfg.Auth.JWTValidServices, cfg.Auth.JWTSecret, 60)

	// Полномочия для TLS соединения
	certMan, err := certmanager.NewCertManager(basePath + "/certs")
	clientCreds, err := certMan.GetClientCredentials()

	conn, err := grpc.DialContext(ctx,
		cfg.GRPC.CoinTarget, // Адрес:порт
		//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
		grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
		grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	)
	require.NoError(t, err)

	// Ristretto кэш
	cacheCoin, err := ristretto.NewRistrettoCoinStorage()
	require.NoError(t, err)

	cacheMiner, err := ristretto.NewRistrettoMinerStorage()
	require.NoError(t, err)

	// remote API miners processor
	coinStorage, err := pb.NewCoinStorage(conn)
	require.NoError(t, err)

	minerStorage, err := pb.NewMinerStorage(conn)
	require.NoError(t, err)

	//***** Remote ClickHouse shares timeseries
	connShares, err := grpc.DialContext(ctx,
		cfg.GRPC.SharesTarget, // Адрес:порт
		//grpc.WithTransportCredentials(insecure.NewCredentials()), // Отключаем TLS
		grpc.WithTransportCredentials(*clientCreds), // Включаем TLS
		grpc.WithUnaryInterceptor(jwt.GetClientInterceptor()),
	)
	require.NoError(t, err)

	shareStorage, err := pb.NewShareStorage(connShares)
	require.NoError(t, err)

	usecase := share.NewShareUseCase(shareStorage, minerStorage, coinStorage, cacheMiner, cacheCoin)

	// Загружаем тестовые данные
	var sfSlice []dto.ShareFound
	err = json.Unmarshal([]byte(testData), &sfSlice)
	require.NoError(t, err)

	// Тестируем удаленный API
	sharesBatch := make([]entity.Share, 0, len(sfSlice))
	for _, item := range sfSlice {

		ctx, span := tracer.Start(context.Background(), "ProcessShare")

		start := time.Now().UnixMicro()
		normShare, err := usecase.NormalizeShare(ctx, item)
		end := time.Now().UnixMicro() - start
		fmt.Println(fmt.Sprintf("%v", end))
		if err != nil {
			logger.Log().Error(fmt.Sprintf("Normilize share error: %s", err.Error()))
		}
		require.NoError(t, err)

		require.Equal(t, int64(4), normShare.CoinID)

		span.End()

		// формируем пакет шар на отправку
		sharesBatch = append(sharesBatch, normShare)
	}

	// Записываем данные через API в ClickHouse, расположенный во внешнем микросервисе
	start := time.Now().UnixMicro()
	err = usecase.AddSharesBatch(sharesBatch)
	end := time.Now().UnixMicro()
	require.NoError(t, err)

	fmt.Println(fmt.Sprintf("%.6f", float64(end-start)/1000000))

	// Тестируем записанные данные путем их получения и сравнения с отправленными

}

var testData string = `[{"uuid":"23c4567b-f8e4-473f-bd06-a0ff8b295e82","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe.test_local","shareDate":1734885835318,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.003806","nonce":"9c44020300030000020003000104030400010401546c0600","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"d805d576-51e2-4df7-8915-7c16980070df","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe1.test_local","shareDate":1734885835600,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.012485","nonce":"9c440203000300000200030001040304000104014b1f0cc0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"e015b41b-38e1-40bc-8236-d8d17c640693","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe2.test_local","shareDate":1734885842376,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.003747","nonce":"9c44010102010303030101030102030201010002c93d0dc0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"5fd36415-39b1-4f4d-87b1-8fae79de6cea","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe3.test_local","shareDate":1734885856165,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.010218","nonce":"9c44030301030300030100030104030202010004a2a700c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"4c97ea64-9fb3-4988-852c-4821e746aca6","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe4.test_local","shareDate":1734885871857,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.005254","nonce":"9c44040401020402000101000201020201000201e0f71d80","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"39a098f5-4d7a-4ada-a0cf-5e031740ffbf","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe5.test_local","shareDate":1734885872138,"cHrate":0,"aHrate":0,"difficulty":"0.002649","sharedif":"0.005910","nonce":"9c44040401020402000101000201020201000201507b2200","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"69946949-fe02-409b-99f9-4bb4c3022324","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe6.test_local","shareDate":1734885897818,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.013472","nonce":"9c44030101030203020000010104010402040202ee830780","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"d4e2de4a-5cfd-4f91-a718-da5385881d0e","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe7.test_local1","shareDate":1734885900176,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.005153","nonce":"9c44020200020401020102010000000003040003401b0d00","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"11144fe2-7eb9-4131-b1d9-eecfb43161e4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe8.test_local","shareDate":1734885917960,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.012632","nonce":"9c44040203020203040304040400020400020300633f3140","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"378cd699-d2ad-4ae6-ae59-6823ec3e5653","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe9.test_local","shareDate":1734885918996,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.032333","nonce":"9c44030003040000000102040402030100040401d24815c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"9e7b8188-01a6-4989-8805-e4337e31195a","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe10.test_local","shareDate":1734885924927,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.049402","nonce":"9c4404020204020004030400030403040103030284b30a80","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"9c21ae88-d0e9-4c05-bfff-b59d439e1cfc","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe11.test_local","shareDate":1734885935749,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.050125","nonce":"9c44010100030200000304000301040304000300842c17c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"8e4f5784-109c-4e01-82ae-46ac54aa39ac","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe12.test_local","shareDate":1734885953230,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.368530","nonce":"9c440403040204010200040201010102020103011ba22c00","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"bf86fdf6-a9c1-4754-b9b8-f6a135e69a6f","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe13.test_local","shareDate":1734885974858,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.032649","nonce":"9c4401030203010003030304030402010004000228c50840","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"00592a0c-5220-4eab-ac7e-21fc568666e5","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe14.test_local","shareDate":1734885979961,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.005900","nonce":"9c44030004020104030103000304000100000002bb7d2440","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"f47a5661-720e-48bd-89e0-7fcff4e248a6","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe15.test_local","shareDate":1734885985798,"cHrate":0,"aHrate":0,"difficulty":"0.003974","sharedif":"0.005872","nonce":"9c4404030304040103030103010002010403040497860e00","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"37e513fa-827b-4123-ba2a-cab9bf251e7e","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe16.test_local","shareDate":1734886010586,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.017017","nonce":"9c4404040300030003020103010002000102020282620d00","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"ecf7ad0f-493c-41e0-975e-6d24d6ea12db","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe17.test_local2","shareDate":1734886056877,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.042427","nonce":"9c4400020301030002040300000302000403040440c83340","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"729513b0-d345-418a-ad16-8b1e16e08117","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe18.test_local","shareDate":1734886057673,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.024282","nonce":"9c4402030402020202030102000004000101040210130e40","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"d7ed9e99-3b06-487e-9315-54f65697b149","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe19.test_local","shareDate":1734886064172,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.019714","nonce":"9c44030202030103040402010000010302040204f1551780","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"334be8bf-d63c-4fb5-ac04-94e019160ffa","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe20.test_local","shareDate":1734886085966,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.013730","nonce":"9c44000203000001030003020003030102010100b5e905c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"b42cfb09-964b-48b6-a664-19ff90dad52b","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe21.test_local","shareDate":1734886112793,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.007799","nonce":"9c4400010100020000000402000101010203040144af0a80","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"03ae6f83-e444-4c23-8008-92f39789e6d5","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe22.test_local","shareDate":1734886113413,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.018658","nonce":"9c4400010100020000000402000101010203040192a819c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"c7f96ea3-043b-48d0-bd27-2fea5954cdf4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe23.test_local","shareDate":1734886113866,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.011432","nonce":"9c440001010002000000040200010101020304010b1b2400","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"57deba0c-9a8e-4bdb-85de-d061199cd31a","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe24.test_local","shareDate":1734886120028,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.024677","nonce":"9c440101000400040200000102000100030402012eaf1540","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"abfb28c8-4b5d-4d64-974e-8404c6290303","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe25.test_local","shareDate":1734886125750,"cHrate":0,"aHrate":0,"difficulty":"0.005961","sharedif":"0.018604","nonce":"9c440102040402000201030303020102000302012f9211c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}, 
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"},
{"uuid":"ba5f3eff-13bb-48a4-bf33-5bcb3c084ce4","blockType":"share_found","serverId":"EU-HSHP-ALPH-1","coinSymbol":"ALPH","workerfull":"15DPDpMdvB3iKzS3mVykxPqSyvE3SdArSUeE98vwyoyKe26.test_local","shareDate":1734886172493,"cHrate":0,"aHrate":0,"difficulty":"0.008941","sharedif":"5.146770","nonce":"9c44010001030201010202030400040402040304915711c0","minerIp":"127.0.0.1","isSolo":false,"rewardMethod":"PPLNS","cost":"0.000000"}]`
