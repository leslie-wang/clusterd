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
	ErrNotExist         = errors.New("required item not exist")
)

func WriteError(w http.ResponseWriter, err error) {
	fmt.Println(string(debug.Stack()))
	switch err {
	case ErrInvalidSourceURL:
		fallthrough
	case ErrNotExist:
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
