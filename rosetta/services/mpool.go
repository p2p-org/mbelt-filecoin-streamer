package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
)

const mpoolNotImplementedMessage = "mpool endpoints are not implemented yet!"

type MempoolAPI struct {}

func NewMempoolAPI() server.MempoolAPIServicer {
	return &MempoolAPI{}
}

func (s *MempoolAPI) Mempool(
	ctx context.Context, req *types.NetworkRequest,
) (*types.MempoolResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, mpoolNotImplementedMessage)
}

func (s *MempoolAPI) MempoolTransaction(
	ctx context.Context, req *types.MempoolTransactionRequest,
) (*types.MempoolTransactionResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, mpoolNotImplementedMessage)
}