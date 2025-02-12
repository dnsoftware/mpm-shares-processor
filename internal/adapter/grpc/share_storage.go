package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-shares-processor/adapter/grpc/proto"
	"github.com/dnsoftware/mpm-shares-processor/entity"
)

type GRPCShareStorage struct {
	client proto.SharesServiceClient
	conn   *grpc.ClientConn
}

func NewShareStorage(conn *grpc.ClientConn) (*GRPCShareStorage, error) {
	client := proto.NewSharesServiceClient(conn)

	return &GRPCShareStorage{
		client: client,
		conn:   conn,
	}, nil
}

func (s *GRPCShareStorage) AddSharesBatch(ctx context.Context, shares []entity.Share) error {

	batch := make([]*proto.Share, 0, len(shares))

	for _, sh := range shares {
		newShare := &proto.Share{
			Uuid:         sh.UUID,
			ServerId:     sh.ServerID,
			CoinId:       sh.CoinID,
			WorkerId:     sh.WorkerID,
			WalletId:     sh.WalletID,
			ShareDate:    sh.ShareDate,
			Difficulty:   sh.Difficulty,
			ShareDif:     sh.Sharedif,
			Nonce:        sh.Nonce,
			IsSolo:       sh.IsSolo,
			RewardMethod: sh.RewardMethod,
			Cost:         sh.Cost,
		}

		batch = append(batch, newShare)
	}

	resp, err := s.client.AddSharesBatch(ctx, &proto.AddSharesBatchRequest{Shares: batch})
	if err != nil {
		return err
	}
	_ = resp

	return nil
}
