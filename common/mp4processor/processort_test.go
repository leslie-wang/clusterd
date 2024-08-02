package mp4processor

import (
	"encoding/binary"
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMp4Rewrite(t *testing.T) {
	input := "../../tests/test-media/fmp4/init.mp4"
	output := "./init_output.mp4"
	ofile, err := os.Create(output)
	assert.Nil(t, err)

	assert.Nil(t, RewriteDuration(input, ofile, 12075, 11979))

	mp4, err := Open(output)
	assert.Nil(t, err)
	assert.Equal(t, uint32(12075), mp4.Moov.Mvhd.Duration)

	// hard code check atom value
	content, err := os.ReadFile(output)
	assert.Nil(t, err)

	// check moov duration
	assert.Equal(t, uint32(12075), binary.BigEndian.Uint32(content[56:60]))
	content[58] = 0
	content[59] = 0

	// check video duration
	// tkhd - video
	assert.Equal(t, uint32(11979), binary.BigEndian.Uint32(content[176:180]))
	content[178] = 0
	content[179] = 0

	// elst - video
	assert.Equal(t, uint32(11979), binary.BigEndian.Uint32(content[276:280]))
	content[278] = 0
	content[279] = 0

	// check audio duration
	// tkhd - audio
	assert.Equal(t, uint32(12075), binary.BigEndian.Uint32(content[694:698]))
	content[696] = 0
	content[697] = 0

	// elst - audio
	fmt.Println(content[276:280])

	assert.Equal(t, uint32(12075), binary.BigEndian.Uint32(content[782:786]))
	content[784] = 0
	content[785] = 0

	// read original input, set reserved value to 0, then compare
	contentOrigin, err := os.ReadFile(input)
	assert.Nil(t, err)

	contentOrigin[1230] = 0
	contentOrigin[1231] = 0
	contentOrigin[1232] = 0
	contentOrigin[1233] = 0

	assert.EqualValues(t, contentOrigin, content)
}
