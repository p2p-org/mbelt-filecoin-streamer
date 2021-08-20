package blocks

import (
	"context"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
	"math/big"
)

type BlocksService struct {
	config  *config.Config
	kafkaDs *datastore.KafkaDatastore
	api     *client.APIClient
}

type BlockFromDb struct {
	Cid           string
	Height        int64
	WinCount      int64
	Miner         string
	Validated     bool
	ParentBaseFee big.Int
	BlockTime     int64
}

type BlockFromDbWithMessagesCids struct {
	Cid           string
	Height        int64
	WinCount      int64
	Miner         string
	Validated     bool
	ParentBaseFee big.Int
	BlockTime     int64
	MessagesCids  []string
}

func Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, apiClient *client.APIClient) (*BlocksService, error) {
	return &BlocksService{
		config:  config,
		kafkaDs: kafkaDs,
		api:     apiClient,
	}, nil
}

func (s *BlocksService) GetHeadUpdates(ctx context.Context, resChan *chan []*api.HeadChange) {
	go s.api.GetHeadUpdates(ctx, resChan)
}

func (s *BlocksService) Push(blocks []*types.BlockHeader, ctx context.Context) {
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
		if block != nil {
			m[block.Cid().String()] = serializeHeader(block)
		}
	}

	s.kafkaDs.Push(datastore.TopicBlocks, m, ctx)
}

func serializeHeader(header *types.BlockHeader) map[string]interface{} {
	var blockSig map[string]interface{}
	if header.BlockSig != nil {
		blockSig = map[string]interface{}{
			"type": header.BlockSig.Type,
			"data": header.BlockSig.Data,
		}
	}

	var blsAggregate map[string]interface{}
	if header.BLSAggregate != nil {
		blockSig = map[string]interface{}{
			"type": header.BLSAggregate.Type,
			"data": header.BLSAggregate.Data,
		}
	}

	result := map[string]interface{}{
		"cid":             header.Cid().String(),
		"height":          header.Height,
		"win_count":       header.ElectionProof.WinCount,
		"miner":           header.Miner.String(),
		"messages_cid":    header.Messages.String(),
		"validated":       header.IsValidated(),
		"blocksig":        blockSig,
		"bls_aggregate":   blsAggregate,
		"block":           header,
		"parent_base_fee": header.ParentBaseFee,
		// "block_time": time.Unix(int64(header.Timestamp), 0).Format(time.RFC3339),
		"block_time":      header.Timestamp,
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
