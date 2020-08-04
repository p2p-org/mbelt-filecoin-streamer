package blocks

import (
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
		m[block.Cid().String()] = block
	}

	s.ds.Push(datastore.TopicBlocks, m)
}
