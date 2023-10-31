package db

import (
	"database/sql"
	"fmt"
	"github.com/leslie-wang/clusterd/common/db/mysql"
	"github.com/leslie-wang/clusterd/common/db/sqlite"
	"github.com/leslie-wang/clusterd/types"
)

const (
	MySQL  = "mysql"
	Sqlite = "sqlite"
)

// OpenDB opens the db
func OpenDB(cfg types.Config) (*sql.DB, error) {
	switch cfg.Driver {
	case MySQL:
		return mysql.OpenDB(cfg)
	case Sqlite:
		return sqlite.OpenDB(cfg)
	}

	return nil, fmt.Errorf("not support DB driver")
}
