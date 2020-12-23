package pg

import (
	"context"
)

//height
//parents
//parent_weight
//parent_state
//blocks
//min_timestamp
//state

const SelectTipset = "select * from tipsets "

func (db *PgDatastore) GetCurrentTipset(ctx context.Context) (*Tipset, error) {
	var tipset Tipset
	err := db.db.Get(&tipset, SelectTipset+` order by height desc limit 1`)
	//err := row.Scan(&tipset.Cid, &tipset.Height, &tipset.BlockTime, &tipset.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetGenesisTipset(ctx context.Context) (*Tipset, error) {
	var tipset Tipset
	err := db.db.GetContext(ctx, &tipset, SelectTipset+` where height = 0 limit 1`)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetTipsetByCID(ctx context.Context, cid string) (*Tipset, error) {
	var tipset Tipset
	err := db.db.GetContext(ctx, &tipset, SelectTipset+" where cid = $1 limit 1", cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetTipsetByHeight(ctx context.Context, height int64) (*Tipset, error) {
	var tipset Tipset
	err := db.db.GetContext(ctx, &tipset, SelectTipset+" where height = $1 limit 1", height)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetParentTipsetByCID(ctx context.Context, cid string) (*Tipset, error) {
	var tipset Tipset
	err := db.db.GetContext(ctx, &tipset, SelectTipset+` 
	where cid = (select blocks.parents -> 'cids' ->> 0
	from blocks where cid = $1);`, cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetParentTipsetByHeight(ctx context.Context, height int64) (*Tipset, error) {
	var tipset Tipset
	err := db.db.GetContext(ctx, &tipset, SelectTipset+` 
	where cid = (select blocks.parents -> 'cids' ->> 0
	from blocks where height = $1);`, height)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return &tipset, err
}

func (db *PgDatastore) GetParentsTipsets(ctx context.Context, cid string) ([]Tipset, error) {
	var tipset []Tipset

	q := `
select * from filecoin.blocks
where cid = ANY (
--     select  array_agg(filecoin.blocks.parents -> 'cids')
    select  jsonb_array_elements_text(filecoin.blocks.parents -> 'cids')
    from  filecoin.blocks
    where cid = $1
);`
	err := db.db.SelectContext(ctx, &tipset, q, cid)
	//err := row.Scan(&block.Cid, &block.Height, &block.BlockTime, &block.Parents)
	return tipset, err
}
