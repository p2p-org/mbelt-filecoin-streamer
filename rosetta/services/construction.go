package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
)

const constructionAPINotImplementedMessage = "Construction API endpoints are not implemented yet!"

type ConstructionAPI struct {}

func NewConstructionAPI() server.ConstructionAPIServicer {
	return &ConstructionAPI{}
}

func (s *ConstructionAPI) ConstructionCombine(
	ctx context.Context, req *types.ConstructionCombineRequest,
)(*types.ConstructionCombineResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionDerive(
	ctx context.Context, request *types.ConstructionDeriveRequest,
) (*types.ConstructionDeriveResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionHash(
	ctx context.Context, request *types.ConstructionHashRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionMetadata(
	ctx context.Context, request *types.ConstructionMetadataRequest,
) (*types.ConstructionMetadataResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionParse(
	ctx context.Context, request *types.ConstructionParseRequest,
) (*types.ConstructionParseResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionPayloads(
	ctx context.Context, request *types.ConstructionPayloadsRequest,
) (*types.ConstructionPayloadsResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionPreprocess(
	ctx context.Context, request *types.ConstructionPreprocessRequest,
) (*types.ConstructionPreprocessResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}

func (s *ConstructionAPI) ConstructionSubmit(
	ctx context.Context, request *types.ConstructionSubmitRequest,
) (*types.TransactionIdentifierResponse, *types.Error) {
	return nil, common.NewErrorWithMessage(common.NotImplementedError, constructionAPINotImplementedMessage)
}
