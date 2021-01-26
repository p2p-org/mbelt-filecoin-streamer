package pg

import (
	"context"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

const (
	cid    = "testCID"
	height = 8
)

var block_time = time.Date(2020, 11, 30, 0, 0, 0, 0, time.Local)

// a successful case
func TestGetBlockByCID(t *testing.T) {

	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()
	sqlxDB := &sqlx.DB{
		DB:     db,
		Mapper: nil,
	}
	dbs := PgDatastore{
		//conn: nil,
		db: sqlxDB,
	}
	positiveRows := sqlmock.NewRows([]string{
		"cid",
		"height",
		"block_time",
	}).AddRow(cid, height, block_time)

	mock.ExpectQuery(`select cid, height, block_time from blocks where cid = \$1 limit 1`).WithArgs(cid).WillReturnRows(positiveRows)

	testCID := cid
	block, err := dbs.GetBlockByCID(context.Background(), testCID)
	if err != nil {
		t.Errorf("error was not expected while GetBlockByCID: %s", err)
	}

	if block.Cid != cid {
		t.Errorf("expected cid %s, got %s", cid, block.Cid)
	}
}

// a failing test case
//func TestShouldRollbackStatUpdatesOnFailure(t *testing.T) {
//	db, mock, err := sqlmock.New()
//	if err != nil {
//		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
//	}
//	defer db.Close()
//
//	mock.ExpectBegin()
//	mock.ExpectExec("UPDATE products").WillReturnResult(sqlmock.NewResult(1, 1))
//	mock.ExpectExec("INSERT INTO product_viewers").
//		WithArgs(2, 3).
//		WillReturnError(fmt.Errorf("some error"))
//	mock.ExpectRollback()
//
//	// now we execute our method
//	if err = recordStats(db, 2, 3); err == nil {
//		t.Errorf("was expecting an error, but there was none")
//	}
//
//	// we make sure that all expectations were met
//	if err := mock.ExpectationsWereMet(); err != nil {
//		t.Errorf("there were unfulfilled expectations: %s", err)
//	}
//}
