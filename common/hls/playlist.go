package hls

import (
	"os"
	"path/filepath"

	"github.com/bluenviron/gohlslib/pkg/playlist"
	"github.com/leslie-wang/clusterd/common/logger"
	"github.com/pkg/errors"
)

func ParseMediaPlaylist(filename string) (*playlist.Media, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, err
	}

	pl, err := playlist.Unmarshal(content)
	if err != nil {
		return nil, err
	}

	mediaPL, ok := pl.(*playlist.Media)
	if !ok {
		return nil, errors.New("invalid media playlist file")
	}
	return mediaPL, nil
}

func CalculateDuration(media *playlist.Media) uint64 {
	var d float64
	for _, seg := range media.Segments {
		d += seg.Duration.Seconds()
	}
	return uint64(d * 1000)
}

func CalculateFileSize(dir string, media *playlist.Media, logger *logger.Logger) (size uint64) {
	initFile := filepath.Join(dir, "init.mp4")
	stat, err := os.Stat(initFile)
	if err != nil {
		logger.Warnf("stat %s: %s", initFile, err)
	} else {
		size += uint64(stat.Size())
	}
	for _, seg := range media.Segments {
		fname := filepath.Join(dir, seg.URI)
		stat, err = os.Stat(fname)
		if err != nil {
			logger.Warnf("stat %s: %s", fname, err)
			continue
		}
		size += uint64(stat.Size())
	}
	return
}
