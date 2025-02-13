package share

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"

	"github.com/dnsoftware/mpm-shares-processor/internal/dto"
	"github.com/dnsoftware/mpm-shares-processor/internal/entity"
)

// NormalizeShare нормализация шары
// получаем коды монеты. майнера, воркера из кеша или из базы, чтобы сформировать структуру шары для вставки в базу данных
func (u *ShareUseCase) NormalizeShare(ctxTask context.Context, shareFound dto.ShareFound) (entity.Share, error) {

	tracer := otel.Tracer("StartNormalize")
	ctxTask, taskSpan := tracer.Start(ctxTask, "ProcessTask")

	//*** получение кода монеты в базе
	ctxCoin, spanCoin := tracer.Start(ctxTask, "GetCoin")

	coinID, err := u.coinCache.GetCoinIDByName(shareFound.CoinSymbol)
	if err != nil {
		return entity.Share{}, err
	}
	if coinID == 0 { // в кэше нет
		_, spanCoinRemote := tracer.Start(ctxCoin, "GetCoinRemote")

		coinID, err = u.coinStorage.GetCoinIDByName(ctxCoin, shareFound.CoinSymbol)
		if err != nil {
			spanCoinRemote.RecordError(err)
			spanCoinRemote.SetStatus(codes.Error, err.Error())
			return entity.Share{}, err
		}
		if coinID == 0 { // в базе нет (а должна быть, так как в миграциях заполнили все монеты в таблице)
			spanCoinRemote.RecordError(err)
			spanCoinRemote.SetStatus(codes.Error, "coinID must be greater then 0")
			return entity.Share{}, fmt.Errorf("coinID must be greater then 0")
		}

		coinID, err = u.coinCache.CreateCoin(shareFound.CoinSymbol, coinID) // кешируем
		if err != nil {
			spanCoinRemote.RecordError(err)
			spanCoinRemote.SetStatus(codes.Error, err.Error())
			return entity.Share{}, err
		}
		if coinID == 0 { // в кеше должна быть уже
			spanCoinRemote.RecordError(err)
			spanCoinRemote.SetStatus(codes.Error, "coinID in cache must be greater then 0")
			return entity.Share{}, fmt.Errorf("coinID in cache must be greater then 0")
		}

		spanCoinRemote.End()
	}
	spanCoin.End()

	ctxWallet, spanWallet := tracer.Start(ctxTask, "GetWallet")
	//*** получение кода майнера(кошелька) в базе
	walletName := WalletFromWorkerfull(shareFound.Workerfull)

	// Запрашиваем данные майнера из кеша
	walletID, err := u.minerCache.GetWalletIDByName(walletName, coinID, shareFound.RewardMethod)
	if err != nil {
		spanWallet.RecordError(err)
		spanWallet.SetStatus(codes.Error, err.Error())
		return entity.Share{}, err
	}
	if walletID == 0 { // нет в кеше
		_, spanWalletRemote := tracer.Start(ctxWallet, "GetWalletRemote")

		walletID, err = u.minerStorage.GetWalletIDByName(ctxWallet, walletName, coinID, shareFound.RewardMethod)
		if err != nil {
			spanWalletRemote.RecordError(err)
			spanWalletRemote.SetStatus(codes.Error, err.Error())
			return entity.Share{}, err
		}
		spanWalletRemote.End()

		walletEntity := entity.Wallet{
			ID:           walletID,
			CoinID:       coinID,
			Name:         walletName,
			IsSolo:       shareFound.IsSolo,
			RewardMethod: shareFound.RewardMethod,
		}

		if walletID == 0 { // нет в базе
			_, spanAddWalletRemote := tracer.Start(ctxWallet, "AddWalletRemote")

			walletID, err = u.minerStorage.CreateWallet(ctxWallet, walletEntity)
			if err != nil {
				spanAddWalletRemote.RecordError(err)
				spanAddWalletRemote.SetStatus(codes.Error, err.Error())
				return entity.Share{}, err
			}
			walletEntity.ID = walletID

			spanAddWalletRemote.End()
		}

		walletID, err = u.minerCache.CreateWallet(walletEntity)
	}

	spanWallet.End()

	ctxWorker, spanWorker := tracer.Start(ctxTask, "GetWorker")
	//*** получение кода воркера в базе
	workerName := WorkerFromWorkerfull(shareFound.Workerfull)

	// Запрашиваем данные воркера из кеша
	workerID, err := u.minerCache.GetWorkerIDByName(shareFound.Workerfull, coinID, shareFound.RewardMethod)
	if err != nil {
		spanWorker.RecordError(err)
		spanWorker.SetStatus(codes.Error, err.Error())
		return entity.Share{}, err
	}
	if workerID == 0 { // нет в кеше
		_, spanWorkerRemote := tracer.Start(ctxWorker, "GetWorkerRemote")

		workerID, err = u.minerStorage.GetWorkerIDByName(ctxWorker, shareFound.Workerfull, coinID, shareFound.RewardMethod)
		if err != nil {
			spanWorkerRemote.RecordError(err)
			spanWorkerRemote.SetStatus(codes.Error, err.Error())
			return entity.Share{}, err
		}
		spanWorkerRemote.End()

		workerEntity := entity.Worker{
			ID:           workerID,
			CoinID:       coinID,
			Workerfull:   shareFound.Workerfull,
			Wallet:       walletName,
			Worker:       workerName,
			ServerID:     shareFound.ServerID,
			IP:           shareFound.MinerIp,
			IsSolo:       shareFound.IsSolo,
			RewardMethod: shareFound.RewardMethod,
		}

		if workerID == 0 { // нет в базе
			_, spanAddWorkerRemote := tracer.Start(ctxWorker, "AddWorkerRemote")

			workerID, err = u.minerStorage.CreateWorker(ctxWorker, workerEntity)
			if err != nil {
				spanAddWorkerRemote.RecordError(err)
				spanAddWorkerRemote.SetStatus(codes.Error, err.Error())
				return entity.Share{}, err
			}
			workerEntity.ID = workerID

			spanAddWorkerRemote.End()
		}

		workerID, err = u.minerCache.CreateWorker(workerEntity)
	}

	// формируем entity.Share
	share := shareFound.ToShare()
	share.CoinID = coinID
	share.WalletID = walletID
	share.WorkerID = workerID

	spanWorker.End()
	taskSpan.End()

	return share, nil
}
