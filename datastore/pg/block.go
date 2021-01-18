package pg

import (
	"context"
)

const tbl_blocks = "blocks"
const (
	SelectBlocks            = "select cid, height, block_time from filecoin.blocks "
	SelectBlocksWithParents = `select cid, height, block_time,
		ARRAY(SELECT json_array_elements_text((parents -> 'cids')::json))
		as parents_cids  from filecoin.blocks`
)

func (db *PgDatastore) GetCurrentBlock(ctx context.Context) (*Block, error) {
	var block Block
	err := db.db.Get(&block, SelectBlocks+` order by height desc limit 1`)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetGenesisBlock(ctx context.Context) (*Block, error) {
	var block Block
	err := db.db.GetContext(ctx, &block, SelectBlocks+` where height = 0 limit 1`)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetBlockByCID(ctx context.Context, cid string) (*Block, error) {
	var block Block
	err := db.db.GetContext(ctx, &block, SelectBlocks+" where cid = $1 limit 1", cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetBlockByHeight(ctx context.Context, height int64) (*Block, error) {
	var block Block
	err := db.db.GetContext(ctx, &block, SelectBlocksWithParents+" where height = $1 limit 1", height)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetParentBlockByCID(ctx context.Context, cid string) (*Block, error) {
	var block Block
	err := db.db.GetContext(ctx, &block, SelectBlocks+` 
	where cid = (select blocks.parents -> 'cids' ->> 0
	from blocks where cid = $1);`, cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetParentBlockByHeight(ctx context.Context, height int64) (*Block, error) {
	var block Block
	err := db.db.GetContext(ctx, &block, SelectBlocksWithParents+` 
	where cid = (select blocks.parents -> 'cids' ->> 0
	from blocks where height = $1  LIMIT 1);`, height)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &block, err
}

func (db *PgDatastore) GetParentsBlocks(ctx context.Context, cid string) ([]Block, error) {
	var block []Block

	q := `
select * from filecoin.blocks
where cid = ANY (
--     select  array_agg(filecoin.blocks.parents -> 'cids')
    select  jsonb_array_elements_text(filecoin.blocks.parents -> 'cids')
    from  filecoin.blocks
    where cid = $1
);`
	err := db.db.SelectContext(ctx, &block, q, cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return block, err
}
