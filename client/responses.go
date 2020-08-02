package client

import (
	"github.com/filecoin-project/lotus/api"
	"github.com/filecoin-project/lotus/chain/types"
)

const (
	ChainHead              = "Filecoin.ChainHead"
	ChainGetGenesis        = "Filecoin.ChainGetGenesis"
	ChainGetBlock          = "Filecoin.ChainGetBlock"
	ChainGetTipSetByHeight = "Filecoin.ChainGetTipSetByHeight"
	ChainGetBlockMessages  = "Filecoin.ChainGetBlockMessages"
	ChainGetMessage        = "Filecoin.ChainGetMessage"
)

var ()

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
