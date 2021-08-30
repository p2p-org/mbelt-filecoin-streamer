package tipsets

import (
	"context"
	"github.com/afiskon/promtail-client/promtail"
	"github.com/filecoin-project/go-state-types/abi"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/utils"
	"math/big"
	"strconv"
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
	logger promtail.Client
}

type TipSetWithState struct {
	*types.TipSet
	State uint8
}

type TipSetFromDb struct {
	Height int64
	Parents []string
	ParentWeight int64
	ParentState string
	Blocks []string
	MinTs int64
	State uint8
}

func Init(conf *config.Config, ds *datastore.KafkaDatastore, api *client.APIClient, l promtail.Client) (*TipSetsService, error) {
	return &TipSetsService{
		config: conf,
		ds:     ds,
		api:    api,
		logger: l,
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

func (s *TipSetsService) PushTipSetsToRevert(height int, ctx context.Context) {
	s.ds.Push(datastore.TopicTipsetsToRevert, map[string]interface{}{strconv.Itoa(height): height}, ctx)
}

func (s *TipSetsService) Push(tipset *TipSetWithState, ctx context.Context) {
	s.push(datastore.TopicTipSets, tipset, ctx)
}

func (s *TipSetsService) PushNormalState(tipset *types.TipSet, ctx context.Context) {
	s.push(datastore.TopicTipSets, &TipSetWithState{tipset, StateNormal}, ctx)
}

func (s *TipSetsService) push(topic string, tipset *TipSetWithState, ctx context.Context) {
	defer func() {
		if r := recover(); r != nil {
			s.logger.Errorf("[TipSetsService][Recover] Panic thrown: %s", r)
		}
	}()

	m := map[string]interface{}{
		tipset.Height().String(): serializeTipSet(tipset),
	}

	s.ds.Push(topic, m, ctx)
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
