package manager

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"time"
)

/*
func (h *Handler) notifyPushEvent(url string, typ types.LiveCallbackEventType) {
	switch typ {
	case types.LiveCallbackEventTypePushStart:
	case types.LiveCallbackEventTypePushStop:
	case types.LiveCallbackEventTypeRecordFile:
	case types.LiveCallbackEventTypeException:
	case types.LiveCallbackEventTypeRecordStatus:
	}
}
*/

const (
	retryNotifyCount    = 12
	retryNotifyInterval = time.Minute
)

func notify(url string, event interface{}) {
	content, err := json.Marshal(event)
	if err != nil {
		log.Printf("WARN: generate json while notifying %v: %s", event, err)
		return
	}
	buf := bytes.NewBuffer(content)
	for i := 0; i < retryNotifyCount; i++ {
		resp, err := http.Post(url, "application/json", buf)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		} else if err != nil {
			log.Printf("WARN: notify %s: %s", string(content), err)
		} else {
			log.Printf("WARN: notify %s got %d", string(content), resp.StatusCode)
		}
		time.Sleep(retryNotifyInterval)
	}
	log.Printf("WARN: failed to notify %s after retry", string(content))
}
