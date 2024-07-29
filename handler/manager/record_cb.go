package manager

import (
	"bytes"
	"encoding/json"
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

func notify(url string, sessionID string, event interface{}) {
	content, err := json.Marshal(event)
	if err != nil {
		defaultLogger.Warnf("generate json while notifying %v: %s", event, err)
		return
	}
	defaultLogger.Infof("jobd %s notify %s: %s", sessionID, url, string(content))
	buf := bytes.NewBuffer(content)
	for i := 0; i < retryNotifyCount; i++ {
		resp, err := http.Post(url, "application/json", buf)
		if err == nil && resp.StatusCode == http.StatusOK {
			resp.Body.Close()
			return
		} else if err != nil {
			defaultLogger.Warnf("notify %s: %s", string(content), err)
		} else {
			defaultLogger.Warnf("notify %s got %d", string(content), resp.StatusCode)
		}
		time.Sleep(retryNotifyInterval)
	}
	defaultLogger.Warnf("failed to notify %s after retry", string(content))
}
