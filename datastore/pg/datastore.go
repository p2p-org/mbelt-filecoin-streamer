package pg

import (
	"database/sql"
	"errors"
	_ "github.com/lib/pq"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
)

const QueryGetMaxHeight = `SELECT coalesce(max(height), 0) FROM filecoin.blocks`

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
	r := ds.conn.QueryRow(QueryGetMaxHeight)
	err = r.Scan(&height)
	return
}
