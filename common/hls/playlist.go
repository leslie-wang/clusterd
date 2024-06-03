package hls

import (
	"os"

	"github.com/bluenviron/gohlslib/pkg/playlist"
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
