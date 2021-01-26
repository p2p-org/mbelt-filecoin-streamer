package pg

import "context"

const SelectMessage = `select cid,block_cid,value,from,to,block_time 
from filecoin.messages`

func (db *PgDatastore) GetMessageByCid(ctx context.Context, cid string) (*Message, error) {
	var msg Message
	err := db.db.Get(&msg, SelectBlocks+` where cid=$1 limit 1`, cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &msg, err
}

func (db *PgDatastore) GetMessageByBlockCid(ctx context.Context, blockCid string) (*Message, error) {
	var msg Message
	err := db.db.Get(&msg, SelectBlocks+` where block_cid=$1 limit 1`, blockCid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &msg, err
}
