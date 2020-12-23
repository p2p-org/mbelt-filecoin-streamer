package pg

import (
	"database/sql"
	"errors"

	"github.com/jmoiron/sqlx"

	_ "github.com/lib/pq"
	"github.com/p2p-org/mbelt-filecoin-streamer/config"
)

const (
	QueryGetMaxHeight = `SELECT coalesce(max(height), 0) FROM filecoin.blocks`
)

type PgDatastore struct {
	conn *sql.DB
	db   *sqlx.DB
}

func Init(config *config.Config) (*PgDatastore, error) {
	if config == nil {
		return nil, errors.New("can't init postgres datastore with nil config")
	}

	db, err := sql.Open("postgres", config.PgUrl)
	if err != nil {
		return nil, err
	}
	sqlxConnet, err := sqlx.Connect("postgres", config.PgUrl)
	if err != nil {
		return nil, err
	}
	//db.SetMaxOpenConns(15)
	db.SetMaxIdleConns(30)

	ds := &PgDatastore{db, sqlxConnet}
	return ds, nil
}

func (ds *PgDatastore) GetMaxHeight() (height int, err error) {
	r := ds.conn.QueryRow(QueryGetMaxHeight)
	err = r.Scan(&height)
	return
}

func (ds *PgDatastore) GetMaxHeightOfTipsets() (height int, err error) {
	r := ds.conn.QueryRow(`SELECT coalesce(max(height), 0) FROM filecoin.tipsets`)
	err = r.Scan(&height)
	return
}
