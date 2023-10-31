package mysql

import (
	"database/sql"
	"github.com/leslie-wang/clusterd/types"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
)

// OpenDB open mysql db
func OpenDB(cfg types.Config) (*sql.DB, error) {
	mcfg := mysql.NewConfig()
	mcfg.User = cfg.DBUser
	mcfg.Passwd = cfg.DBPass
	mcfg.Addr = cfg.Addr
	mcfg.DBName = cfg.DBName
	mcfg.ParseTime = true

	db, err := sql.Open("mysql", mcfg.FormatDSN())
	if err != nil {
		return nil, err
	}

	// TODO: tune this info, or put it into config file
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)
	return db, nil
}
