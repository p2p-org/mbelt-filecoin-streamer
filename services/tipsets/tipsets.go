package tipsets

import (
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/utils"
	"log"
	"math/big"
)

const (
	StateNormal uint8 = iota
	StateNull
	StateInProgress
)

type TipSetsService struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

type TipSetWithState struct {
	*types.TipSet
	State uint8
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

func (s *TipSetsService) GetByKey(key types.TipSetKey) *types.TipSet {
	return s.api.GetByKey(key)
}

func (s *TipSetsService) PushTipSetsToRevert(tipset *TipSetWithState) {
	s.push(datastore.TopicTipsetsToRevert, tipset)
}

func (s *TipSetsService) Push(tipset *TipSetWithState) {
	s.push(datastore.TopicTipSets, tipset)
}

func (s *TipSetsService) PushNormalState(tipset *types.TipSet) {
	s.push(datastore.TopicTipSets, &TipSetWithState{tipset, StateNormal})
}

func (s *TipSetsService) push(topic string, tipset *TipSetWithState) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("[TipSetsService][Recover]", "Throw panic", r)
		}
	}()

	m := map[string]interface{}{
		tipset.Height().String(): serializeTipSet(tipset),
	}

	s.ds.Push(topic, m)
}

func serializeTipSet(tipset *TipSetWithState) map[string]interface{} {
	parentWeight, parentState, minTimestamp, blocks, parents := new(big.Int), "", uint64(0), "{}", "{}"

	if tipset.State != StateNull {
		parentWeight = tipset.ParentWeight().Int
		parentState = tipset.ParentState().String()
		minTimestamp = tipset.MinTimestamp()
		blocks = utils.CidsToVarcharArray(tipset.Cids())
		parents = utils.CidsToVarcharArray(tipset.Parents().Cids())
	}

	result := map[string]interface{}{
		"height":        tipset.Height(),
		"parent_weight": parentWeight,
		"parent_state":  parentState,
		"min_timestamp": minTimestamp,
		"blocks":        blocks,
		"parents":       parents,
		"state":         tipset.State,
	}

	return result
}
