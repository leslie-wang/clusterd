package runner

import (
	"context"
	"log"
	"os/exec"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/client/manager"
	"github.com/leslie-wang/clusterd/types"
	"github.com/pkg/errors"
)

// Config is configuration for the handler
type Config struct {
	MgrHost string
	MgrPort uint
	Name    string
}

// Handler is structure for recorder API
type Handler struct {
	c Config
	r *mux.Router

	cli *manager.Client
}

// NewHandler create new instance of Handler struct
func NewHandler(c Config) *Handler {
	h := &Handler{c: c}
	h.cli = manager.NewClient(c.MgrHost, c.MgrPort)
	return h
}

func (h *Handler) CreateRouter() *mux.Router {
	if h.r == nil {
		h.r = mux.NewRouter()
		//h.r.HandleFunc(types.RecordStartURL, h.start).Methods(http.MethodPost)
	}
	return h.r
}

func (h *Handler) Run(ctx context.Context) error {
	for {
		job, err := h.cli.RegisterRunner(h.c.Name)
		if err != nil {
			log.Printf("Request job: %s", err)
		} else if job != nil {
			err = h.runJob(ctx, job)
			if err != nil {
				log.Printf("Handle job %+v: %v", job, err)
			}
		}
		after := time.After(10 * time.Second)
		select {
		case <-after:
		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (h *Handler) runJob(ctx context.Context, j *types.Job) error {
	if len(j.Commands) == 0 {
		return errors.Errorf("Empty commands in job %d", j.ID)
	}
	name := j.Commands[0]
	var arg []string
	if len(j.Commands) > 1 {
		arg = j.Commands[1:]
	}
	content, err := exec.Command(name, arg...).CombinedOutput()
	if err != nil {
		return errors.Wrapf(err, "%s", string(content))
	}
	return nil
}
