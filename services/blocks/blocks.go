package blocks

import (
	"context"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/pg"
	"log"
)

type BlocksService struct {
	config  *config.Config
	kafkaDs *datastore.KafkaDatastore
	pgDs    *pg.PgDatastore
	api     *client.APIClient
}

func Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, pgDs *pg.PgDatastore, apiClient *client.APIClient) (*BlocksService, error) {
	return &BlocksService{
		config:  config,
		kafkaDs: kafkaDs,
		pgDs:    pgDs,
		api:     apiClient,
	}, nil
}

func (s *BlocksService) GetHeadUpdates(ctx context.Context, resChan *chan []*api.HeadChange) {
	go s.api.GetHeadUpdates(ctx, resChan)
}

func (s *BlocksService) GetMaxHeightFromDB() (int, error) {
	return s.pgDs.GetMaxHeight()
}

func (s *BlocksService) PushBlocks(blocks []*types.BlockHeader) {
	s.push(datastore.TopicBlocks, blocks)
}

func (s *BlocksService) PushBlocksToRevert(blocks []*types.BlockHeader) {
	s.push(datastore.TopicBlocksToRevert, blocks)
}

func (s *BlocksService) push(topic string, blocks []*types.BlockHeader) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[BlocksService][Recover]", "Throw panic", r)
		}
	}()

	if blocks == nil {
		return
	}

	m := map[string]interface{}{}

	for _, block := range blocks {
		m[block.Cid().String()] = serializeHeader(block)
	}

	s.kafkaDs.Push(topic, m)
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
