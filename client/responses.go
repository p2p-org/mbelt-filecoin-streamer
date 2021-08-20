package client

import (
	"encoding/json"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
	"github.com/filecoin-project/specs-actors/v3/actors/builtin/miner"
	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p-core/peer"
)

const (
	ChainHead              = "Filecoin.ChainHead"
	ChainGetGenesis        = "Filecoin.ChainGetGenesis"
	ChainGetBlock          = "Filecoin.ChainGetBlock"
	ChainGetTipSetByHeight = "Filecoin.ChainGetTipSetByHeight"
	ChainGetTipSet         = "Filecoin.ChainGetTipSet"
	ChainGetBlockMessages  = "Filecoin.ChainGetBlockMessages"
	ChainGetMessage        = "Filecoin.ChainGetMessage"
	ChainGetParentMessages = "Filecoin.ChainGetParentMessages"
	ChainGetParentReceipts = "Filecoin.ChainGetParentReceipts"
	ChainNotify            = "Filecoin.ChainNotify"
	ChainHasObj            = "Filecoin.ChainHasObj"

	StateGetActor          = "Filecoin.StateGetActor"
	StateChangedActors     = "Filecoin.StateChangedActors"
	StateReadState         = "Filecoin.StateReadState"
	StateListMiners        = "Filecoin.StateListMiners"
	StateMinerInfo         = "Filecoin.StateMinerInfo"
	StateMinerPower        = "Filecoin.StateMinerPower"
	StateMinerSectors      = "Filecoin.StateMinerSectors"
	StateLookupID          = "Filecoin.StateLookupID"
	StateAccountKey        = "Filecoin.StateAccountKey"
	StateNetworkName       = "Filecoin.StateNetworkName"
	StateNetworkVersion    = "Filecoin.StateNetworkVersion"

	NetPeers = "Filecoin.NetPeers"
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

type MessageAndCid struct {
	Cid cid.Cid
	Message *types.Message
}

type MessageAndCidResponse struct {
	APIResponse
	Result []*MessageAndCid `json:"result"` // payload
}

type MessageReceiptResponse struct {
	APIResponse
	Result []*types.MessageReceipt `json:"result"` // payload
}

type Actor struct {
	APIResponse
	Result *types.Actor `json:"result"` // payload
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

type AddressResponse struct {
	APIResponse
	Result address.Address
}

type AddressListResponse struct {
	APIResponse
	Result []address.Address `json:"result"` // payload
}

type MinerInfoResponse struct {
	APIResponse
	Result *miner.MinerInfo `json:"result"` // payload
}

type MinerPowerResponse struct {
	APIResponse
	Result *api.MinerPower `json:"result"` // payload
}

type MinerSectorsResponse struct {
	APIResponse
	Result []*miner.SectorOnChainInfo `json:"result"` // payload
}

type StringResponse struct {
	APIResponse
	Result string `json:"result"` // payload
}

type IntResponse struct {
	APIResponse
	Result int `json:"result"` // payload
}

type PeersResponse struct {
	APIResponse
	Result []*peer.AddrInfo `json:"result"` // payload
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

type BlockWs struct {
	Jsonrpc string `json:"jsonrpc"`      // "2.0"
	ID      *int64 `json:"id,omitempty"` // most likely it's nil and according to comment filecoin's go-jsonrpc library it means notification
	Method  string `json:"method"`       // "xrpc.ch.val" (channel with value), "xrpc.ch.close", "xrpc.cancel"

	Params BlockWsParams `json:"params"`
	Meta   map[string]string `json:"meta,omitempty"` // most likely empty
}

type BlockWsParams struct {
	ChanId int // should be 1 in case of head updates
	Block  *types.BlockHeader
}

type TipSetWs struct {
	Jsonrpc string `json:"jsonrpc"`      // "2.0"
	ID      *int64 `json:"id,omitempty"` // most likely it's nil and according to comment filecoin's go-jsonrpc library it means notification
	Method  string `json:"method"`       // "xrpc.ch.val" (channel with value), "xrpc.ch.close", "xrpc.cancel"

	Params TipSetWsParams `json:"params"`
	Meta   map[string]string `json:"meta,omitempty"` // most likely empty
}

type TipSetWsParams struct {
	ChanId int // should be 1 in case of head updates
	TipSet  *types.TipSet
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

func (r *BlockWsParams) UnmarshalJSON(p []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}

	blk := &types.BlockHeader{}

	if err := json.Unmarshal(tmp[0], &r.ChanId); err != nil {
		return err
	}

	if err := json.Unmarshal(tmp[1], blk); err != nil {
		return err
	}

	r.Block = blk

	return nil
}

func (r *TipSetWsParams) UnmarshalJSON(p []byte) error {
	var tmp []json.RawMessage
	if err := json.Unmarshal(p, &tmp); err != nil {
		return err
	}

	ts := &types.TipSet{}

	if err := json.Unmarshal(tmp[0], &r.ChanId); err != nil {
		return err
	}

	if err := json.Unmarshal(tmp[1], ts); err != nil {
		return err
	}

	r.TipSet = ts

	return nil
}
