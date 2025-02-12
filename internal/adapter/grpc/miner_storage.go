package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-shares-processor/adapter/grpc/proto"
	"github.com/dnsoftware/mpm-shares-processor/entity"
)

type GRPCMinerStorage struct {
	client proto.MinersServiceClient
	conn   *grpc.ClientConn
}

func NewMinerStorage(conn *grpc.ClientConn) (*GRPCMinerStorage, error) {
	client := proto.NewMinersServiceClient(conn)

	return &GRPCMinerStorage{
		client: client,
		conn:   conn,
	}, nil
}

func (g *GRPCMinerStorage) CreateWallet(ctx context.Context, wallet entity.Wallet) (int64, error) {
	resp, err := g.client.CreateWallet(ctx, &proto.CreateWalletRequest{
		CoinId:       wallet.CoinID,
		Name:         wallet.Name,
		IsSolo:       wallet.IsSolo,
		RewardMethod: wallet.RewardMethod,
	})

	if err != nil {
		return 0, err
	}

	return resp.Id, err

}

func (g *GRPCMinerStorage) CreateWorker(ctx context.Context, worker entity.Worker) (int64, error) {
	resp, err := g.client.CreateWorker(ctx, &proto.CreateWorkerRequest{
		CoinId:       worker.CoinID,
		Workerfull:   worker.Workerfull,
		Wallet:       worker.Wallet,
		Worker:       worker.Worker,
		ServerId:     worker.ServerID,
		Ip:           worker.IP,
		IsSolo:       worker.IsSolo,
		RewardMethod: worker.RewardMethod,
	})

	if err != nil {
		return 0, err
	}

	return resp.Id, err
}

func (g *GRPCMinerStorage) GetWalletIDByName(ctx context.Context, wallet string, coinID int64, rewardMethod string) (int64, error) {
	resp, err := g.client.GetWalletIDByName(ctx, &proto.GetWalletIDByNameRequest{
		Wallet:       wallet,
		CoinId:       coinID,
		RewardMethod: rewardMethod,
	})

	if err != nil {
		return 0, err
	}

	return resp.Id, err
}

func (g *GRPCMinerStorage) GetWorkerIDByName(ctx context.Context, worker string, coinID int64, rewardMethod string) (int64, error) {
	resp, err := g.client.GetWorkerIDByName(ctx, &proto.GetWorkerIDByNameRequest{
		Workerfull:   worker,
		CoinId:       coinID,
		RewardMethod: rewardMethod,
	})

	if err != nil {
		return 0, err
	}

	return resp.Id, err
}
