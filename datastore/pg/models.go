package pg

import (
	"strconv"
	"time"

	"github.com/filecoin-project/lotus/chain/types"
	"github.com/lib/pq"
)

type Block struct {
	Height    int64            `db:"height"`
	Cid       string           `db:"cid"`
	Parents   []pq.StringArray `db:"parents"`
	BlockTime time.Time        `db:"block_time"`
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
