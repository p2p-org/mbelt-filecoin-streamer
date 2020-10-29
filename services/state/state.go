package state

import (
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/actors/abi"
	"github.com/ipfs/go-cid"
	"github.com/p2p-org/mbelt-filecoin-streamer/client"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore"
	"log"
)

type StateService struct {
	config *config.Config
	ds     *datastore.KafkaDatastore
	api    *client.APIClient
}

type ActorInfo struct {
	Act types.Actor

	StateRoot cid.Cid
	Height    abi.ChainEpoch // so that we can walk the actor changes in chronological order.

	TsKey       types.TipSetKey
	ParentTsKey types.TipSetKey

	Addr  address.Address
	State string
}

func Init(config *config.Config, kafkaDs *datastore.KafkaDatastore, apiClient *client.APIClient) (*StateService, error) {
	return &StateService{
		config: config,
		ds:     kafkaDs,
		api:    apiClient,
	}, nil
}

func (s *StateService) GetChangedActors(start, end cid.Cid) map[string]types.Actor {
	return s.api.GetChangedActors(start, end)
}

func (s *StateService) ChainHasObj(cid cid.Cid) (bool, error) {
	return s.api.ChainHasObj(cid)
}

func (s *StateService) ReadState(actor address.Address, tsk types.TipSetKey) *client.ActorState {
	return s.api.ReadState(actor, tsk)
}

func (s *StateService) PushActors(actors []*ActorInfo) {
	// Empty actor produces panic
	defer func() {
		if r := recover(); r != nil {
			log.Println("[StateService][Recover]", "Throw panic", r)
		}
	}()

	if actors == nil {
		return
	}

	m := map[string]interface{}{}

	for _, actor := range actors {
		m[actor.Addr.String()+"_"+actor.TsKey.String()] = serializeActor(actor)
	}

	s.ds.Push(datastore.TopicActorStates, m)
}

func serializeActor(actor *ActorInfo) map[string]interface{} {

	result := map[string]interface{}{
		"actor_state_key":  actor.Addr.String()+"_"+actor.TsKey.String(),
		"actor_code":       actor.Act.Code.String(),
		"actor_head":       actor.Act.Head.String(),
		"nonce":            actor.Act.Nonce,
		"balance":          actor.Act.Balance,
		"is_account_actor": actor.Act.IsAccountActor(),
		"state_root":       actor.StateRoot.String(),
		"height":           actor.Height,
		"ts_key":           actor.TsKey.String(),
		"parent_ts_key":    actor.ParentTsKey.String(),
		"addr":             actor.Addr.String(),
		"state":            actor.State,
	}

	log.Println(result)

	return result
}
