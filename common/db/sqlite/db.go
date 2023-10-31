package sqlite

import (
	"database/sql"
	"github.com/leslie-wang/clusterd/types"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// OpenDB open sqlite db
func OpenDB(cfg types.Config) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", cfg.DBName)
	if err != nil {
		return nil, err
	}

	// TODO: tune this info, or put it into config file
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}
