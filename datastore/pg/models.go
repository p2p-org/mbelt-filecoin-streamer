package pg

import (
	"strconv"
	"time"

	"github.com/filecoin-project/go-state-types/abi"
	"github.com/ipfs/go-cid"

	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/lib/pq"
)

type Block struct {
	Height    int64          `db:"height"`
	Cid       string         `db:"cid"`
	Parents   pq.StringArray `db:"parents_cids"`
	BlockTime time.Time      `db:"block_time"`
}

type Parents struct {
	Cids      []string `json:"cids"`
	Weight    string   `json:"weight"`
	BaseFee   string   `json:"base_fee"`
	StateRoot struct {
		NAMING_FAILED string `json:"/"`
	} `json:"state_root"`
	MessageReceipts string `json:"message_receipts"`
}

type Tipset struct {
	Height       int64          `db:"height"`
	Blocks       pq.StringArray `db:"blocks"`
	Parents      pq.StringArray `db:"parents"`
	ParentWeight int64          `db:"parent_weight"`
	ParentState  string         `db:"parent_state"`
	State        int            `db:"state"`
	MinTimestamp time.Time      `db:"min_timestamp"`
}

type Message struct {
	Cid       string    `db:"cid"`
	BlockCid  string    `db:"block_cid"`
	Value     int64     `db:"value"`
	From      string    `db:"from"`
	To        string    `db:"to"`
	BlockTime time.Time `db:"block_time"`
}

func ParseTipset(t *types.TipSet) *Tipset {
	var res Tipset
	h, _ := strconv.Atoi(t.Height().String())
	res.Height = int64(h)
	for _, b := range t.Blocks() {
		res.Blocks = append(res.Blocks, b.Cid().String())
	}
	for _, p := range t.Parents().Cids() {
		res.Parents = append(res.Parents, p.String())
	}
	res.MinTimestamp = time.Unix(int64(t.MinTimestamp()), 0)
	res.ParentState = t.ParentState().String()
	res.ParentWeight = t.ParentWeight().Int64()

	return &res

}

func ParseBlocks(t *types.TipSet) []Block {
	var res []Block

	for _, b := range t.Blocks() {
		t := Block{}
		t.Cid = b.Cid().String()
		t.BlockTime = time.Unix(int64(b.Timestamp), 0)
		t.Height = int64(b.Height)
		for _, p := range b.Parents {
			t.Parents = append(t.Parents, p.String())
		}
		res = append(res, t)
	}

	return res

}

func ParseMessages(t []*types.Message, block *Block) []Message {
	var res []Message
	for _, m := range t {
		var tm Message
		tm.Cid = m.Cid().String()
		tm.BlockCid = m.Cid().String()
		tm.Value = m.GasLimit
		tm.From = m.From.String()
		tm.To = m.To.String()
		tm.BlockCid = block.Cid
		res = append(res, tm)
		tm.BlockTime = block.BlockTime
	}
	return res
}

func ParseMessageExtended(t []*types.Message, block *Block) []*messages.MessageExtended {
	var res []*messages.MessageExtended
	for _, m := range t {
		bcid, _ := cid.Cast([]byte(block.Cid))
		em := messages.MessageExtended{
			Cid:       m.Cid(),
			BlockCid:  bcid,
			Height:    abi.ChainEpoch(block.Height),
			Message:   m,
			Timestamp: uint64(block.BlockTime.Unix()),
		}
		res = append(res, &em)

	}
	return res
}
