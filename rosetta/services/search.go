package services

import (
	"context"
	"fmt"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/filter"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
	"strings"
)

type SearchAPI struct {
	ds *pg.PgDatastore
}

func NewSearchAPI(ds *pg.PgDatastore) server.SearchAPIServicer {
	return &SearchAPI{ds: ds}
}

func (s *SearchAPI) SearchTransactions(
	ctx context.Context, request *types.SearchTransactionsRequest,
) (*types.SearchTransactionsResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier); err != nil {
		return nil, err
	}

	if request.Currency != nil && request.Currency.Symbol != common.FILCurrency.Symbol {
		return nil, common.NewErrorWithMessage(common.TransactionNotFoundError, "FIL is the only currency in filecoin")
	}

	var filters []filter.Filter
	if request.TransactionIdentifier != nil && len(request.TransactionIdentifier.Hash) > 0 {
		filters = append(filters, filter.NewKV("cid", request.TransactionIdentifier.Hash, filter.Eq))
	}

	if request.AccountIdentifier != nil && len(request.AccountIdentifier.Address) > 0 {
		accFilter := filter.NewOr(
			filter.NewKV("from", request.TransactionIdentifier.Hash, filter.Eq),
			filter.NewKV("from_id", request.TransactionIdentifier.Hash, filter.Eq),
			filter.NewKV("to", request.TransactionIdentifier.Hash, filter.Eq),
			filter.NewKV("to_id", request.TransactionIdentifier.Hash, filter.Eq),
		)
		filters = append(filters, accFilter)
	}

	if request.Address != nil && len(*request.Address) > 0 {
		accFilter := filter.NewOr(
			filter.NewKV("from", *request.Address, filter.Eq),
			filter.NewKV("from_id", *request.Address, filter.Eq),
			filter.NewKV("to", *request.Address, filter.Eq),
			filter.NewKV("to_id", *request.Address, filter.Eq),
		)
		filters = append(filters, accFilter)
	}

	if request.Status != nil && len(*request.Status) > 0 {
		if *request.Status == common.OperationStatusOutOfGas.Status {
			filters = append(filters, filter.NewKV("exit_code", 7, filter.Eq))
		} else if *request.Status == common.OperationStatusFailure.Status {
			filters = append(filters, filter.NewKV("exit_code", 0, filter.Neq))
		} else if *request.Status == common.OperationStatusSuccess.Status {
			filters = append(filters, filter.NewKV("exit_code", 0, filter.Eq))
		} else {
			return nil, common.NewErrorWithMessage(common.SanityCheckError, "unknown status provided")
		}
	}

	if request.Type != nil && len(*request.Type) > 0 {
		filters = append(filters, filter.NewKV("method_name", *request.Type, filter.Eq))
	}

	if request.Success != nil {
		if *request.Success && request.Status != nil && *request.Status != common.OperationStatusSuccess.Status {
			return nil, common.NewErrorWithMessage(common.SanityCheckError, "success is true but status is not Success")
		} else if !*request.Success && request.Status != nil && *request.Status == common.OperationStatusSuccess.Status {
			return nil, common.NewErrorWithMessage(common.SanityCheckError, "success is false but status is Success")
		}

		if *request.Success {
			filters = append(filters, filter.NewKV("exit_code", 0, filter.Eq))
		} else {
			filters = append(filters, filter.NewKV("exit_code", 0, filter.Neq))
		}
	}

	if request.Operator == nil && len(filters) > 1 {
		return nil, common.NewErrorWithMessage(common.SanityCheckError, "more then one query parameter provided but operator not specified")
	}

	var fltr filter.Filter
	if *request.Operator == types.AND {
		fltr = filter.NewAnd(filters...)
	} else if *request.Operator == types.OR {
		fltr = filter.NewOr(filters...)
	}

	if request.MaxBlock != nil {
		fltr = filter.NewAnd(fltr, filter.NewKV("msg.height", *request.MaxBlock, filter.Le))
	}

	var limit int64 = 100
	var offset int64
	if request.Limit != nil {
		limit = *request.Limit
	}
	if request.Offset != nil {
		offset = *request.Offset
	}

	msgs, tsks, err := s.ds.SearchMessages(fltr, limit, offset)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.TransactionNotFoundError, err)
	}

	trxs := common.RosettaTransactionsFromMessagesFromDb(msgs)
	blockTrxs := make([]*types.BlockTransaction, 0, len(trxs))
	for i, trx := range trxs {
		blockTrxs = append(blockTrxs, &types.BlockTransaction{
			BlockIdentifier: &types.BlockIdentifier{
				Index: msgs[i].Height,
				Hash:  fmt.Sprintf("{%s}", strings.Join(tsks[i], ",")),
			},
			Transaction:     trx,
		})
	}

	cnt, err := s.ds.SearchMessagesCount(fltr)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.TransactionNotFoundError, err)
	}

	var nextOffset *int64
	delta := cnt - (offset + limit)
	if delta > 0 {
		nextOffset = &delta
	}

	return &types.SearchTransactionsResponse{
		Transactions: blockTrxs,
		TotalCount:   cnt,
		NextOffset:   nextOffset,
	}, nil
}

