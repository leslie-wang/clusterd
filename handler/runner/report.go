package runner

import (
	"log"

	"github.com/leslie-wang/clusterd/types"
)

func (h *Handler) reportLoop() {
	for {
		r := <-h.reportChan
		err := h.cli.ReportJobStatus(&r)
		if err != nil {
			log.Printf("WARN: report %v: %s", r, err)
		}
	}
}

func (h *Handler) addReport(r types.JobStatus) {
	h.reportChan <- r
}
