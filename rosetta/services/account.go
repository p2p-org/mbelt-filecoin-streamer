package services

import (
	"context"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
)

type AccountAPI struct {
	ds *pg.PgDatastore
}

func NewAccountAPI(ds *pg.PgDatastore) server.AccountAPIServicer {
	return &AccountAPI{ds: ds}
}

func (s *AccountAPI) AccountBalance(
	ctx context.Context, request *types.AccountBalanceRequest,
) (*types.AccountBalanceResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier); err != nil {
		return nil, err
	}

	if request.Currencies != nil {
		for _, curr := range request.Currencies {
			if curr.Symbol != common.FILCurrency.Symbol {
				return nil, common.NewErrorWithMessage(common.SanityCheckError,
					"There is no other currencies then FIL in Filecoin.")
			}
		}
	}

	if request.AccountIdentifier == nil || len(request.AccountIdentifier.Address) == 0 {
		return nil, common.NewErrorWithMessage(common.SanityCheckError,
			"Account address was not provided")
	}

	var acc *state.ActorFromDb
	var err error
	if request.BlockIdentifier != nil && request.BlockIdentifier.Index != nil {
		acc, err = s.ds.GetActorStateByAddressAndHeight(request.AccountIdentifier.Address, *request.BlockIdentifier.Index)
		if err != nil {
			return nil, common.NewRosettaErrorFromError(common.AccountNotFoundError, err)
		}

		if request.BlockIdentifier.Hash != nil && len(*request.BlockIdentifier.Hash) > 0 && acc.TsKey != *request.BlockIdentifier.Hash {
			return nil, common.NewErrorWithMessage(common.AccountNotFoundError,
				"Found account for requested height but it's tipset key is different from provided in request (block identifier hash)")
		}
	} else if request.BlockIdentifier.Hash != nil && len(*request.BlockIdentifier.Hash) > 0 {
		acc, err = s.ds.GetActorStateByAddressAndTsKey(request.AccountIdentifier.Address, *request.BlockIdentifier.Hash)
		if err != nil {
			return nil, common.NewRosettaErrorFromError(common.AccountNotFoundError, err)
		}
	} else {
		acc, err = s.ds.GetLatestActorStateByAddress(request.AccountIdentifier.Address)
		if err != nil {
			return nil, common.NewRosettaErrorFromError(common.AccountNotFoundError, err)
		}
	}

	meta := map[string]interface{}{
		"addr": acc.Addr,
		"actor_code": acc.ActorCode,
		"actor_head": acc.ActorHead,
		"nonce": acc.Nonce,
	}

	return &types.AccountBalanceResponse{
		BlockIdentifier: &types.BlockIdentifier{Index: acc.Height, Hash:  acc.TsKey},
		Balances:        []*types.Amount{{Value: acc.Balance.String(), Currency: common.FILCurrency}},
		Metadata:        meta,
	}, nil
}

func (s *AccountAPI) AccountCoins(
	ctx context.Context, request *types.AccountCoinsRequest,
) (*types.AccountCoinsResponse, *types.Error) {
	acc, err := s.AccountBalance(ctx, &types.AccountBalanceRequest{
		NetworkIdentifier: request.NetworkIdentifier,
		AccountIdentifier: request.AccountIdentifier,
		BlockIdentifier:   nil,
		Currencies:        request.Currencies,
	})

	if err != nil {
		return nil, err
	}

	return &types.AccountCoinsResponse{
		BlockIdentifier: acc.BlockIdentifier,
		Coins:           []*types.Coin{{CoinIdentifier: &types.CoinIdentifier{Identifier: common.FILCurrency.Symbol}}},
		Metadata:        acc.Metadata,
	}, nil
}
