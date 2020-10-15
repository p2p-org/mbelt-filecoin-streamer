package tipsets

import (
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
)

type TipSetsService struct {
	config *config.Config
	ds     *datastore.Datastore
	api    *client.APIClient
}

func Init(config *config.Config, ds *datastore.Datastore, apiClient *client.APIClient) (*TipSetsService, error) {
	return &TipSetsService{
		config: config,
		ds:     ds,
		api:    apiClient,
	}, nil
}

func (s *TipSetsService) GetHead() *types.TipSet {
	return s.api.GetHead()
}

func (s *TipSetsService) GetGenesis() *types.TipSet {
	return s.api.GetGenesis()
}

func (s *TipSetsService) GetByHeight(height abi.ChainEpoch) (*types.TipSet, bool) {
	return s.api.GetByHeight(height)
}

func (s *TipSetsService) Push(tipset *types.TipSet) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Throw panic", r)
		}
	}()

	m := map[string]interface{}{
		tipset.Height().String(): serializeTipSet(tipset),
	}

	s.ds.Push(datastore.TipSetBlocks, m)
}

func serializeTipSet(tipset *types.TipSet) map[string]interface{} {
	result := map[string]interface{}{
		"key":           tipset.Key().String(),
		"height":        tipset.Height(),
		"parents":       tipset.Parents().String(),
		"parent_weight": tipset.ParentWeight(),
		"parent_state":  tipset.ParentState().String(),
	}

	blocksCids := make([]string, 0)

	for _, block := range tipset.Blocks() {
		blocksCids = append(blocksCids, block.Cid().String())
	}
	log.Println("cids", tipset.Cids())
	log.Println("Block cids", blocksCids)
	result["blocks"] = blocksCids

	return result
}
