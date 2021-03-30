package pg

import (
	"database/sql"
	"errors"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
)

const GetMaxHeight = `SELECT coalesce(max(height), 0) FROM filecoin.tipsets`

const GetTipSetBlocksAndStateByHeight = `SELECT blocks, state FROM filecoin.tipsets WHERE height = $1`

const GetBlocksCountByHeight = `SELECT count(*) FROM filecoin.blocks WHERE height = $1`

const GetMessagesCountByHeight = `SELECT count(*) FROM filecoin.messages WHERE height = $1`

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

