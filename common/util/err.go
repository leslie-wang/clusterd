package util

import (
	"errors"
	"io"
	"strings"
)

func MakeStatusError(body io.Reader) error {
	content, err := io.ReadAll(body)
	if err != nil {
		return err
	}
	return errors.New(strings.TrimSpace(string(content)))
}
