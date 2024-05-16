package manager

import (
	"database/sql"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/leslie-wang/clusterd/common/db"
	"github.com/leslie-wang/clusterd/common/db/job"
	"github.com/leslie-wang/clusterd/common/db/record"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/types"
)

// Config is configuration for the handler
type Config struct {
	Driver           string
	DBAddress        string
	DBUser           string
	DBPass           string
	DBName           string
	ParamQuery       bool
	ScheduleInterval time.Duration
	NotifyURL        string
	BaseURL          string
	MediaDir         string
}

// Handler is structure for recorder API
type Handler struct {
	cfg  Config
	r    *mux.Router
	lock *sync.Mutex

	db       *sql.DB
	recordDB *record.DB
	jobDB    *job.DB

	runners map[string]time.Time // <runner_name, last checkin time>
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) (*Handler, error) {
	h := &Handler{
		cfg:     c,
		lock:    &sync.Mutex{},
		runners: map[string]time.Time{},
	}
	return h, h.init()
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Do stuff here
		if !strings.Contains(r.RequestURI, types.URLJobRunner) {
			log.Printf("%s - %s", r.Method, r.RequestURI)
		}
		// Call the next handler, which can be another middleware in the chain, or the final handler.
		next.ServeHTTP(w, r)
	})
}

func (h *Handler) CreateRouter() *mux.Router {
	if h.r == nil {
		h.r = mux.NewRouter()

		// recording
		h.r.HandleFunc(types.URLRecord, h.record).Methods(http.MethodPost)

		// job related
		h.r.HandleFunc(types.URLJob, h.listJobs).Methods(http.MethodGet)
		h.r.HandleFunc(types.MkIDURLByBase(types.URLJobRunner), h.acquireJob).Methods(http.MethodPost)
		h.r.HandleFunc(types.MkIDURLByBase(types.URLJob), h.reportJob).Methods(http.MethodPost)
		h.r.HandleFunc(types.MkIDURLByBase(types.URLJob), h.getJob).Methods(http.MethodGet)

		// playback
		h.r.HandleFunc(types.MkIDURLByBase(types.URLPlay)+"/{filename}", h.playback).Methods(http.MethodGet)

		h.r.Use(loggingMiddleware)
	}
	return h.r
}

// init will initialize the handler with corresponding handle function
func (h *Handler) init() (err error) {
	// prepare DB
	h.db, err = db.OpenDB(types.Config{
		Driver: h.cfg.Driver,
		DBUser: h.cfg.DBUser,
		DBPass: h.cfg.DBPass,
		DBName: h.cfg.DBName,
		Addr:   h.cfg.DBAddress,
	})
	if err != nil {
		return
	}
	h.jobDB = job.NewDB(h.db)
	err = h.jobDB.Prepare()
	if err != nil {
		return
	}

	h.recordDB = record.NewDB(h.db)
	return h.recordDB.Prepare()
}

func (h *Handler) newTx() (*sql.Tx, error) {
	return h.db.Begin()
}

func hasPrefixInQueryKeys(q url.Values, prefix string) bool {
	for k := range q {
		if strings.HasPrefix(k, prefix) {
			return true
		}
	}
	return false
}
