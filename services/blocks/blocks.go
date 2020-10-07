package blocks

import (
	"context"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
)

type BlocksService struct {
	config *config.Config
	ds     *datastore.Datastore
	api    *client.APIClient
}

func Init(config *config.Config, ds *datastore.Datastore, apiClient *client.APIClient) (*BlocksService, error) {
	return &BlocksService{
		config: config,
		ds:     ds,
		api:    apiClient,
	}, nil
}

func (s *BlocksService) GetHeadUpdates(ctx context.Context, resChan *chan []*api.HeadChange) {
	go s.api.GetHeadUpdates(ctx, resChan)
}

func (s *BlocksService) GetHead() *types.TipSet {
	return s.api.GetHead()
}

func (s *BlocksService) GetGenesis() *types.TipSet {
	return s.api.GetGenesis()
}

func (s *BlocksService) GetByHeight(height abi.ChainEpoch) (*types.TipSet, bool) {
	return s.api.GetByHeight(height)
}

func (s *BlocksService) Push(blocks []*types.BlockHeader) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Cid throw panic")
		}
	}()

	if len(blocks) == 0 {
		return
	}

	m := map[string]interface{}{}

	for _, block := range blocks {
		m[block.Cid().String()] = serializeHeader(block)
	}

	s.ds.Push(datastore.TopicBlocks, m)
}

func serializeHeader(header *types.BlockHeader) map[string]interface{} {
	result := map[string]interface{}{
		"cid":          header.Cid().String(),
		"height":       header.Height,
		"win_count":    header.ElectionProof.WinCount,
		"miner":        header.Miner.String(),
		"messages_cid": header.Messages.String(),
		"validated":    header.IsValidated(),
		"blocksig": map[string]interface{}{
			"type": header.BlockSig.Type,
			"data": header.BlockSig.Data,
		},
		"bls_aggregate": map[string]interface{}{
			"type": header.BLSAggregate.Type,
			"data": header.BLSAggregate.Data,
		},
		"block": header,
		// "block_time": time.Unix(int64(header.Timestamp), 0).Format(time.RFC3339),
		"block_time": header.Timestamp,
	}

	// Parents data
	parentCids := make([]string, 0)

	for _, parentBlock := range header.Parents {
		parentCids = append(parentCids, parentBlock.String())
	}

	result["parents"] = map[string]interface{}{
		"cids":             parentCids,
		"state_root":       header.ParentStateRoot,
		"weight":           header.ParentWeight,
		"base_fee":         header.ParentBaseFee,
		"message_receipts": header.ParentMessageReceipts.String(),
	}

	return result
}
