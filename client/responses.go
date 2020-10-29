package client

import (
	"encoding/json"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
)

const (
	ChainHead              = "Filecoin.ChainHead"
	ChainGetGenesis        = "Filecoin.ChainGetGenesis"
	ChainGetBlock          = "Filecoin.ChainGetBlock"
	ChainGetTipSetByHeight = "Filecoin.ChainGetTipSetByHeight"
	ChainGetTipSet         = "Filecoin.ChainGetTipSet"
	ChainGetBlockMessages  = "Filecoin.ChainGetBlockMessages"
	ChainGetMessage        = "Filecoin.ChainGetMessage"
	ChainNotify            = "Filecoin.ChainNotify"
	ChainHasObj            = "Filecoin.ChainHasObj"

	StateChangedActors     = "Filecoin.StateChangedActors"
	StateReadState         = "Filecoin.StateReadState"
)

type TipSet struct {
	APIResponse
	Result *types.TipSet `json:"result"` // payload
}

type Block struct {
	APIResponse
	Result *types.BlockHeader `json:"result"` // payload
}

type BlockMessages struct {
	APIResponse
	Result *api.BlockMessages `json:"result"` // payload
}

type Message struct {
	APIResponse
	Result *types.Message `json:"result"` // payload
}

type Actors struct {
	APIResponse
	Result map[string]types.Actor `json:"result"` // payload
}

type HasObj struct {
	APIResponse
	Result bool `json:"result"` // payload
}

type ActorStateResponse struct {
	APIResponse
	Result *ActorState `json:"result"` // payload
}

type ActorState struct {
	Balance types.BigInt
	State   interface{}
}

type HeadUpdates struct {
	Jsonrpc string `json:"jsonrpc"`      // "2.0"
	ID      *int64 `json:"id,omitempty"` // most likely it's nil and according to comment filecoin's go-jsonrpc library it means notification
	Method  string `json:"method"`       // "xrpc.ch.val" (channel with value), "xrpc.ch.close", "xrpc.cancel"

	Params HeadUpdatesParams `json:"params"`
	Meta   map[string]string `json:"meta,omitempty"` // most likely empty
}

type HeadUpdatesParams struct {
	ChanId      int // should be 1 in case of head updates
	HeadChanges []*api.HeadChange
}

func (r *HeadUpdatesParams) UnmarshalJSON(p []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}

	changes := make([]*api.HeadChange, 0, len(tmp) - 1)

	if err := json.Unmarshal(tmp[0], &r.ChanId); err != nil {
		return err
	}

	if err := json.Unmarshal(tmp[1], &changes); err != nil {
		return err
	}

	r.HeadChanges = changes

	return nil
}
