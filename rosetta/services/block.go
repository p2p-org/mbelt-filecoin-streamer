package services

import (
	"context"
	"fmt"
	"github.com/VictoriaMetrics/fastcache"
	"github.com/coinbase/rosetta-sdk-go/server"
	"github.com/coinbase/rosetta-sdk-go/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"github.com/p2p-org/mbelt-filecoin-streamer/rosetta/common"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
	"strings"
)

type BlockAPI struct {
	ds    *pg.PgDatastore
	cache *fastcache.Cache
}

func NewBlockAPI(/*cacheFile string, cacheSize int,*/ ds *pg.PgDatastore) server.BlockAPIServicer {
	return &BlockAPI{
		ds:    ds,
		//cache: fastcache.LoadFromFileOrNew(cacheFile, cacheSize),
	}
}

// TODO: Add cache support
func (s *BlockAPI) Block(
	ctx context.Context, req *types.BlockRequest,
) (res *types.BlockResponse, rosettaError *types.Error) {
	if err := assertValidNetworkIdentifier(req.NetworkIdentifier); err != nil {
		return nil, err
	}

	blkId := req.BlockIdentifier

	if blkId.Index == nil && (blkId.Hash == nil || len(*blkId.Hash) > 0) {
		return nil, common.NewErrorWithMessage(common.SanityCheckError,
			"neither block index nor block hash specified")
	}

	var ts tipsets.TipSetFromDb
	var err error
	if blkId.Hash != nil && len(*blkId.Hash) > 0 {
		tsk := *blkId.Hash
		if blkId.Index != nil {
			ts, err = s.ds.GetTipSetByHeightAndBlocks(*blkId.Index, tsk)
		} else {
			ts, err = s.ds.GetTipSetByBlocks(tsk)
		}
	} else { // if we get here it means that blkId.Hash not specified but blkId.Index is
		ts, err = s.ds.GetTipSetByHeight(*blkId.Index)
	}

	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.BlockNotFoundError, err)
	}

	parentTsk := fmt.Sprintf("{%s}", strings.Join(ts.Parents, ","))
	parentTs, err := s.ds.GetTipSetByBlocks(parentTsk)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.BlockNotFoundError, err)
	}

	blockId := &types.BlockIdentifier{
		Index: ts.Height,
		Hash:  fmt.Sprintf("{%s}", strings.Join(ts.Blocks, ",")),
	}
	parentBlockId := &types.BlockIdentifier{
		Index: parentTs.Height,
		Hash:  parentTsk,
	}

	if ts.State != tipsets.StateNormal {
		return &types.BlockResponse{
			Block: &types.Block{
				BlockIdentifier:       blockId,
				ParentBlockIdentifier: parentBlockId,
				Timestamp:             ts.MinTs,
				Metadata:              map[string]interface{}{"state": ts.State},
			}}, nil
	}

	blks, err := s.ds.GetBlocksWithMessagesCidsByHeight(ts.Height)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.BlockNotFoundError, err)
	}

	msgs, err := s.ds.GetMessagesByHeight(ts.Height)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.TransactionNotFoundError, err)
	}

	trxs := common.RosettaTransactionsFromMessagesFromDb(msgs)
	
	blocksInMeta := make([]map[string]interface{}, len(blks))
	for _, blk := range blks {
		m, err := types.MarshalMap(blk)
		if err != nil {
			return nil, common.NewRosettaErrorFromError(common.MarshallingError, err)
		}

		blocksInMeta = append(blocksInMeta, m)
	}

	tsMeta, err := types.MarshalMap(ts)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.MarshallingError, err)
	}

	tsMeta["blocks"] = blocksInMeta

	blk := &types.Block{
		BlockIdentifier:       blockId,
		ParentBlockIdentifier: parentBlockId,
		Timestamp:             ts.MinTs,
		Transactions:          trxs,
		Metadata:              tsMeta,
	}

	return &types.BlockResponse{Block: blk}, nil
}

func (s *BlockAPI) BlockTransaction(
	ctx context.Context, request *types.BlockTransactionRequest,
) (*types.BlockTransactionResponse, *types.Error) {
	if err := assertValidNetworkIdentifier(request.NetworkIdentifier); err != nil {
		return nil, err
	}

	if request.TransactionIdentifier == nil {
		return nil, common.NewErrorWithMessage(common.SanityCheckError, "Transaction identifier is nil")
	}

	msg, err := s.ds.GetMessageByCid(request.TransactionIdentifier.Hash)
	if err != nil {
		return nil, common.NewRosettaErrorFromError(common.TransactionNotFoundError, err)
	}

	if msg.Height != request.BlockIdentifier.Index {
		return nil, common.NewErrorWithMessage(common.TransactionNotFoundError,
			"Found transaction with provided hash but it belongs to different tipset according to provided block index")
	}

	trx := common.RosettaTransactionFromMessageFromDb(msg)

	return &types.BlockTransactionResponse{Transaction: trx}, nil
}
