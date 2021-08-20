package pg

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
	"github.com/p2p-org/mbelt-filecoin-streamer/datastore/filter"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/blocks"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/messages"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/state"
	"github.com/p2p-org/mbelt-filecoin-streamer/services/tipsets"
)

const GetMaxHeight = `SELECT coalesce(max(height), 0) FROM filecoin.tipsets`

const GetMaxHeightTipSet = `
	SELECT
		height, parents, parent_weight, parent_state, blocks, min_timestamp, state
	FROM filecoin.tipsets
	ORDER BY height DESC
	LIMIT 1
`

const GetTipSetByHeight = `
	SELECT
		height, parents, parent_weight, parent_state, blocks, min_timestamp, state
	FROM filecoin.tipsets
	WHERE height = $1
`

const GetTipSetsWithLimitAndOffset = `
	SELECT
		height, blocks
	FROM filecoin.tipsets
	ORDER BY height ASC
	LIMIT $1
	OFFSET $2
`

const GetTipSetByBlocks = `
	SELECT
		height, parents, parent_weight, parent_state, blocks, min_timestamp, state
	FROM filecoin.tipsets
	WHERE blocks = $1
`

const GetTipSetByHeightAndBlocks = `
	SELECT
		height, parents, parent_weight, parent_state, blocks, min_timestamp, state
	FROM filecoin.tipsets
	WHERE height = $1 AND blocks = $2
`

const GetMinHeightTipSetBlocks = `SELECT height, blocks FROM filecoin.tipsets ORDER BY height ASC LIMIT 1`

const GetTipSetBlocksAndStateByHeight = `SELECT blocks, state FROM filecoin.tipsets WHERE height = $1`

const GetBlocksCountByHeight = `SELECT count(*) FROM filecoin.blocks WHERE height = $1`

const GetMessagesCountByHeight = `SELECT count(*) FROM filecoin.messages WHERE height = $1`

const GetMinBlockTimestamp = `SELECT min_timestamp FROM filecoin.tipsets ORDER BY height ASC limit 1`

const GetBlocksByHeight = `
	SELECT
		cid, height, win_count, miner, validated, parent_base_fee, block_time
	FROM filecoin.blocks
	WHERE height = $1
`

const GetBlocksWithMessagesCidsByHeight = `
	SELECT
		blk.cid, blk.height, win_count, blk.miner, validated, blk.parent_base_fee, blk.block_time,
		array(
			SELECT msg.cid FROM filecoin.messages msg WHERE height = $1 AND blk.cid = ANY(block_cids)
		) AS msg_cids
	FROM filecoin.blocks blk
	WHERE height = $1
`

const GetMessagesCidsByHeight = `SELECT cid FROM filecoin.messages WHERE height = $1`

const GetMessagesCidsByHeightAndBlockCid = `SELECT cid FROM filecoin.messages WHERE height = $1 AND $2 = ANY(block_cids)`

const GetMessagesByHeight = `
	SELECT
		cid, height, block_cids, method, method_name, "from", from_id, from_type, "to", to_id, to_type, "value",
		gas_limit, gas_premium, gas_fee_cap, gas_used, base_fee, exit_code, block_time 
	FROM filecoin.messages
	WHERE height = $1
`

const GetMessagesByCid = `
	SELECT
		cid, height, block_cids, method, method_name, "from", from_id, from_type, "to", to_id, to_type, "value",
		gas_limit, gas_premium, gas_fee_cap, gas_used, base_fee, exit_code, block_time 
	FROM filecoin.messages
	WHERE cid = $1
`

const SearchMessages = `
	SELECT
		cid, msg.height, block_cids, method, method_name, "from", from_id, from_type, "to", to_id, to_type, "value",
		gas_limit, gas_premium, gas_fee_cap, gas_used, base_fee, exit_code, block_time, ts.blocks 
	FROM filecoin.messages msg
	LEFT JOIN filecoin.tipsets ts ON msg.height = ts.height 
	WHERE %s
`

const SearchMessagesCount = `
	SELECT 
		count(*) 
	FROM filecoin.messages
	WHERE %s
`

const GetLatestActorStateByAddress = `
	SELECT
		actor_code, actor_head, nonce, balance, height, ts_key, addr 
	FROM filecoin.actor_states
	WHERE addr = $1
	ORDER BY height DESC
	LIMIT 1
`

const GetActorStateByAddressAndHeight = `
	SELECT
		actor_code, actor_head, nonce, balance, height, ts_key, addr 
	FROM filecoin.actor_states
	WHERE addr = $1 AND height = $2
`

const GetActorStateByAddressAndTsKey = `
	SELECT
		actor_code, actor_head, nonce, balance, height, ts_key, addr 
	FROM filecoin.actor_states
	WHERE addr = $1 AND ts_key = $2
`



type PgDatastore struct {
	conn *sql.DB
}

func Init(config *config.Config) (*PgDatastore, error) {
	if config == nil {
		return nil, errors.New("can't init postgres datastore with nil config")
	}

	db, err := sql.Open("postgres", config.PgUrl)
	ds := &PgDatastore{db}
	return ds, err
}

func (ds *PgDatastore) GetMaxHeight() (height int, err error) {
	r := ds.conn.QueryRow(GetMaxHeight)
	err = r.Scan(&height)
	return
}

func (ds *PgDatastore) GetTipSetBlocksAndStateByHeight(height int) (blocks []string, state int, err error) {
	r := ds.conn.QueryRow(GetTipSetBlocksAndStateByHeight, height)
	err = r.Scan(pq.Array(&blocks), &state)
	return
}

func (ds *PgDatastore) GetMaxHeightTipSet() (ts tipsets.TipSetFromDb, err error) {
	r := ds.conn.QueryRow(GetMaxHeightTipSet)
	err = r.Scan(&ts.Height, pq.Array(&ts.Parents), &ts.ParentWeight, &ts.ParentState, pq.Array(&ts.Blocks), &ts.MinTs, &ts.State)
	return
}

func (ds *PgDatastore) GetTipSetByHeight(height int64) (ts tipsets.TipSetFromDb, err error) {
	r := ds.conn.QueryRow(GetTipSetByHeight, height)
	err = r.Scan(&ts.Height, pq.Array(&ts.Parents), &ts.ParentWeight, &ts.ParentState, pq.Array(&ts.Blocks), &ts.MinTs, &ts.State)
	return
}

func (ds *PgDatastore) GetTipSetsWithLimitAndOffset(limit, offset int64) (heights []int64, keys [][]string, err error) {
	rows, err := ds.conn.Query(GetTipSetsWithLimitAndOffset, limit, offset)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return
	}

	for rows.Next() {
		var height int64
		var tsKey []string

		err = rows.Scan(&height, pq.Array(&tsKey))
		if err != nil {
			return
		}

		heights = append(heights, height)
		keys = append(keys, tsKey)
	}

	return
}

func (ds *PgDatastore) GetTipSetByBlocks(blocks string) (ts tipsets.TipSetFromDb, err error) {
	r := ds.conn.QueryRow(GetTipSetByBlocks, blocks)
	err = r.Scan(&ts.Height, pq.Array(&ts.Parents), &ts.ParentWeight, &ts.ParentState, pq.Array(&ts.Blocks), &ts.MinTs, &ts.State)
	return
}

func (ds *PgDatastore) GetTipSetByHeightAndBlocks(height int64, blocks string) (ts tipsets.TipSetFromDb, err error) {
	r := ds.conn.QueryRow(GetTipSetByHeightAndBlocks, height, blocks)
	err = r.Scan(&ts.Height, pq.Array(&ts.Parents), &ts.ParentWeight, &ts.ParentState, pq.Array(&ts.Blocks), &ts.MinTs, &ts.State)
	return
}

func (ds *PgDatastore) GetMinHeightTipSetBlocks() (height int64, blocks []string, err error) {
	r := ds.conn.QueryRow(GetMinHeightTipSetBlocks)
	err = r.Scan(pq.Array(&blocks), &height)
	return
}

func (ds *PgDatastore) GetBlocksCountByHeight(height int) (count int, err error) {
	r := ds.conn.QueryRow(GetBlocksCountByHeight, height)
	err = r.Scan(&count)
	return
}

func (ds *PgDatastore) GetMessagesCountByHeight(height int) (count int, err error) {
	r := ds.conn.QueryRow(GetMessagesCountByHeight, height)
	err = r.Scan(&count)
	return
}

func (ds *PgDatastore) GetMinBlockTimestamp() (timestamp int64, err error) {
	r := ds.conn.QueryRow(GetMinBlockTimestamp)
	err = r.Scan(&timestamp)
	return
}

func (ds *PgDatastore) GetBlocksByHeight(height int64) ([]*blocks.BlockFromDb, error) {
	rows, err := ds.conn.Query(GetBlocksByHeight, height)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, err
	}

	res := make([]*blocks.BlockFromDb, 0)
	for rows.Next() {
		blk := &blocks.BlockFromDb{}

		err = rows.Scan(&blk.Cid, &blk.Height, &blk.WinCount, &blk.Miner, &blk.Validated, &blk.ParentBaseFee, &blk.BlockTime)

		if err != nil {
			return nil, err
		}

		res = append(res, blk)
	}

	return res, nil
}

func (ds *PgDatastore) GetBlocksWithMessagesCidsByHeight(height int64) ([]*blocks.BlockFromDbWithMessagesCids, error) {
	rows, err := ds.conn.Query(GetBlocksWithMessagesCidsByHeight, height)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, err
	}

	res := make([]*blocks.BlockFromDbWithMessagesCids, 0)
	for rows.Next() {
		blk := &blocks.BlockFromDbWithMessagesCids{}

		err = rows.Scan(&blk.Cid, &blk.Height, &blk.WinCount, &blk.Miner, &blk.Validated, &blk.ParentBaseFee,
			&blk.BlockTime, pq.Array(&blk.MessagesCids))

		if err != nil {
			return nil, err
		}

		res = append(res, blk)
	}

	return res, nil
}

func (ds *PgDatastore) GetMessagesCidsByHeight(height int64) (res []string, err error) {
	rows, err := ds.conn.Query(GetMessagesCidsByHeight, height)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, err
	}

	for rows.Next() {
		var cid string

		err = rows.Scan(&cid)
		if err != nil {
			return nil, err
		}

		res = append(res, cid)
	}

	return
}

func (ds *PgDatastore) GetMessagesCidsByHeightAndBlockCid(height int64, blkCid string) (res []string, err error) {
	rows, err := ds.conn.Query(GetMessagesCidsByHeightAndBlockCid, height, blkCid)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, err
	}

	for rows.Next() {
		var cid string

		err = rows.Scan(&cid)
		if err != nil {
			return nil, err
		}

		res = append(res, cid)
	}

	return
}

func (ds *PgDatastore) GetMessagesByHeight(height int64) (msgs []*messages.MessageFromDb, err error) {
	rows, err := ds.conn.Query(GetMessagesByHeight, height)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, err
	}

	for rows.Next() {
		msg := &messages.MessageFromDb{}

		err = rows.Scan(&msg.Cid, &msg.Height, pq.Array(&msg.BlockCids), &msg.Method, &msg.MethodName, &msg.From,
			&msg.FromId, &msg.FromType, &msg.To, &msg.ToId, &msg.ToType, &msg.Value, &msg.GasLimit, &msg.GasPremium,
			&msg.GasFeeCap, &msg.GasUsed, &msg.BaseFee, &msg.ExitCode, &msg.BlockTime)
		if err != nil {
			return nil, err
		}

		msgs = append(msgs, msg)
	}

	return
}

func (ds *PgDatastore) GetMessageByCid(cid string) (msg *messages.MessageFromDb, err error) {
	msg = &messages.MessageFromDb{}
	r := ds.conn.QueryRow(GetMessagesByCid, cid)

	err = r.Scan(&msg.Cid, &msg.Height, pq.Array(&msg.BlockCids), &msg.Method, &msg.MethodName, &msg.From,
		&msg.FromId, &msg.FromType, &msg.To, &msg.ToId, &msg.ToType, &msg.Value, &msg.GasLimit, &msg.GasPremium,
		&msg.GasFeeCap, &msg.GasUsed, &msg.BaseFee, &msg.ExitCode, &msg.BlockTime)

	return
}

func (ds *PgDatastore) SearchMessages(f filter.Filter, limit, offset int64) (msgs []*messages.MessageFromDb, tsks [][]string, err error) {
	query, args, err := filter.RenderQuery(SearchMessages, f)
	if err != nil {
		return nil, nil, err
	}

	query = query + " LIMIT ? OFFSET ?"
	args = append(args, limit)
	args = append(args, offset)

	rows, err := ds.conn.Query(query, args...)

	defer closeRows(rows)

	if err != nil || rows == nil {
		return nil, nil, err
	}

	for rows.Next() {
		msg := &messages.MessageFromDb{}
		var tsk []string

		err = rows.Scan(&msg.Cid, &msg.Height, pq.Array(&msg.BlockCids), &msg.Method, &msg.MethodName, &msg.From,
			&msg.FromId, &msg.FromType, &msg.To, &msg.ToId, &msg.ToType, &msg.Value, &msg.GasLimit, &msg.GasPremium,
			&msg.GasFeeCap, &msg.GasUsed, &msg.BaseFee, &msg.ExitCode, &msg.BlockTime, pq.Array(&tsk))
		if err != nil {
			return nil, nil, err
		}

		msgs = append(msgs, msg)
		tsks = append(tsks, tsk)
	}

	return
}

func (ds *PgDatastore) SearchMessagesCount(f filter.Filter) (count int64, err error) {
	query, args, err := filter.RenderQuery(SearchMessages, f)
	if err != nil {
		return
	}

	r := ds.conn.QueryRow(query, args...)

	err = r.Scan(&count)

	return
}

func (ds *PgDatastore) GetLatestActorStateByAddress(addr string) (act *state.ActorFromDb, err error) {
	act = &state.ActorFromDb{}
	r := ds.conn.QueryRow(GetLatestActorStateByAddress, addr)

	err = r.Scan(&act.ActorCode, &act.ActorHead, &act.Nonce, &act.Balance, &act.Height, &act.TsKey, &act.Addr)

	return
}

func (ds *PgDatastore) GetActorStateByAddressAndHeight(addr string, height int64) (act *state.ActorFromDb, err error) {
	act = &state.ActorFromDb{}
	r := ds.conn.QueryRow(GetActorStateByAddressAndHeight, addr, height)

	err = r.Scan(&act.ActorCode, &act.ActorHead, &act.Nonce, &act.Balance, &act.Height, &act.TsKey, &act.Addr)

	return
}

func (ds *PgDatastore) GetActorStateByAddressAndTsKey(addr string, tsKey string) (act *state.ActorFromDb, err error) {
	act = &state.ActorFromDb{}
	r := ds.conn.QueryRow(GetActorStateByAddressAndTsKey, addr, tsKey)

	err = r.Scan(&act.ActorCode, &act.ActorHead, &act.Nonce, &act.Balance, &act.Height, &act.TsKey, &act.Addr)

	return
}

func closeRows(rows *sql.Rows) {
	if rows != nil {
		rows.Close()
	}
}

