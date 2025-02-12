package grpc

import (
	"context"

	"google.golang.org/grpc"

	"github.com/dnsoftware/mpm-shares-processor/adapter/grpc/proto"
)

type GRPCCoinStorage struct {
	conn   *grpc.ClientConn
	client proto.MinersServiceClient
}

func NewCoinStorage(conn *grpc.ClientConn) (*GRPCCoinStorage, error) {

	client := proto.NewMinersServiceClient(conn)

	return &GRPCCoinStorage{
		client: client,
		conn:   conn,
	}, nil
}

func (g *GRPCCoinStorage) GetCoinIDByName(ctx context.Context, coin string) (int64, error) {

	resp, err := g.client.GetCoinIDByName(ctx, &proto.GetCoinIDByNameRequest{
		Coin: coin,
	})

	if err != nil {
		return 0, err
	}

	return resp.Id, err
}
