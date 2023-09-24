package manager

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/go-sql-driver/mysql"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/types"
)

// Config is configuration for the handler
type Config struct {
	DBAddress string
	DBUser    string
	DBPass    string
}

// Handler is structure for recorder API
type Handler struct {
	cfg  Config
	r    *mux.Router
	lock *sync.Mutex

	db *sql.DB

	runners map[string]time.Time // <runner_name, last checkin time>
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) (*Handler, error) {
	h := &Handler{
		cfg:     c,
		lock:    &sync.Mutex{},
		runners: map[string]time.Time{},
	}
	h.init()
	return h, h.init()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		log.Println(r.RequestURI)
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) CreateRouter() *mux.Router {
	if h.r == nil {
		h.r = mux.NewRouter()
		h.r.HandleFunc(types.URLRunnerID, h.register).Methods(http.MethodPost)
		h.r.HandleFunc(types.URLRunner, h.listRunners).Methods(http.MethodGet)

		h.r.HandleFunc(types.URLJob, h.createJob).Methods(http.MethodPost)
		h.r.HandleFunc(types.URLJob, h.listJobs).Methods(http.MethodGet)

		h.r.HandleFunc(types.URLJobID, h.reportJob).Methods(http.MethodPost)

		h.r.Use(loggingMiddleware)
	}
	return h.r
}

// init will initialize the handler with corresponding handle function
func (h *Handler) init() error {
	// prepare DB
	return h.prepareDB()
}

func (h *Handler) testDB() {
	cfg := mysql.NewConfig()
	cfg.User = h.cfg.DBUser
	cfg.Passwd = h.cfg.DBPass
	cfg.Addr = h.cfg.DBAddress
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
