package mysql

import (
	"database/sql"
	"time"

	"github.com/leslie-wang/clusterd/types"

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

/*
func  testDB(cfg types.Config) {
	cfg := mysql.NewConfig()
	cfg.User = cfg.DBUser
	cfg.Passwd = cfg.DBPass
	cfg.Addr = cfg.DBAddress
	cfg.DBName = types.ClusterDBName
	cfg.ParseTime = true
	fmt.Println(cfg.FormatDSN())
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		panic(err)
	}
	defer db.Close()

	// See "Important settings" section.
	db.SetConnMaxLifetime(3 * time.Minute)
	db.SetMaxOpenConns(10)
	db.SetMaxIdleConns(10)

	// Prepare statement for inserting data
	stmtIns, err := db.Prepare("INSERT INTO assets (ref_id, url, start_time) VALUES( 1, ?,? )") // ? = placeholder
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Prepare statement for reading data
	stmtOut, err := db.Prepare("SELECT id, url, start_time FROM assets")
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtOut.Close()

	// Insert square numbers for 0-24 in the database
	t := time.Now()
	for i := 0; i < 25; i++ {
		_, err = stmtIns.Exec("http://"+strconv.Itoa(i), t) // Insert tuples (i, i^2)
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		t = t.Add(time.Hour)
	}

	var (
		id  int // we "scan" the result in here
		url string
	)

	// Query the square-number of 13
	rows, err := stmtOut.QueryContext(context.Background())
	if err != nil {
		panic(err.Error()) // proper error handling instead of panic in your app
	}
	for rows.Next() {
		err = rows.Scan(&id, &url, &t) // WHERE number = 13
		if err != nil {
			panic(err.Error()) // proper error handling instead of panic in your app
		}
		fmt.Printf("%d: %s, %v\n", id, url, t)
	}
	fmt.Println("done")
}
*/
