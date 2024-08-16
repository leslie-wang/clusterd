package common

import "path/filepath"

func MkStoragePath(dir string, id string) string {
	return filepath.Join(dir, id)
}
