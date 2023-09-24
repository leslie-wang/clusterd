package util

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"runtime/debug"
)

var (
	ErrInvalidSourceURL = errors.New("invalid source URL")
)

func WriteError(w http.ResponseWriter, err error) {
	fmt.Println(string(debug.Stack()))
	switch err {
	case ErrInvalidSourceURL:
		fallthrough
	default:
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func WriteBody(w http.ResponseWriter, body interface{}) {
	content, err := json.Marshal(body)
	if err != nil {
		WriteError(w, err)
		return
	}
	w.Write(content)
}
