package manager

import (
	"net/http"

	"github.com/gorilla/mux"
	"github.com/leslie-wang/clusterd/common/util"
	"github.com/leslie-wang/clusterd/types"
)

func (h *Handler) register(w http.ResponseWriter, r *http.Request) {
	runner := mux.Vars(r)[types.ID]

	job, err := h.findAndUpdateJobTx(runner)
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteBody(w, job)
	return
}

func (h *Handler) listRunners(w http.ResponseWriter, r *http.Request) {
	runners, err := h.listActiveRunnersDB()
	if err != nil {
		util.WriteError(w, err)
		return
	}
	util.WriteBody(w, runners)
}
