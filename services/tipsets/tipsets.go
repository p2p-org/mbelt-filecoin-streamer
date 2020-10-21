package tipsets

import (
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/utils"
	"log"
)

type TipSetsService struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

func Init(config *config.Config, ds *datastore.KafkaDatastore, apiClient *client.APIClient) (*TipSetsService, error) {
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

func (s *TipSetsService) PushTipSetsToRevert(blocks *types.TipSet) {
	s.push(datastore.TopicTipsetsToRevert, blocks)
}

func (s *TipSetsService) Push(tipset *types.TipSet) {
	s.push(datastore.TopicTipSets, tipset)
}

func (s *TipSetsService) push(topic string, tipset *types.TipSet) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[MessagesService][Recover]", "Throw panic", r)
		}
	}()

	m := map[string]interface{}{
		tipset.Height().String(): serializeTipSet(tipset),
	}

	s.ds.Push(topic, m)
}

func serializeTipSet(tipset *types.TipSet) map[string]interface{} {
	result := map[string]interface{}{
		"height":        tipset.Height(),
		"parent_weight": tipset.ParentWeight(),
		"parent_state":  tipset.ParentState().String(),
		"min_timestamp": tipset.MinTimestamp(),
	}

	blocksCids := make([]string, 0)

	for _, cid := range tipset.Cids() {
		blocksCids = append(blocksCids, cid.String())
	}
	result["blocks"] = utils.ToVarcharArray(blocksCids)

	parentsCids := make([]string, 0)

	if len(tipset.Blocks()) > 0 {
		for _, cid := range tipset.Blocks()[0].Parents {
			parentsCids = append(parentsCids, cid.String())
		}
	}

	result["parents"] = utils.ToVarcharArray(parentsCids)

	return result
}
