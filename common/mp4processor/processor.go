package mp4processor

import (
	"os"

	"github.com/Eyevinn/mp4ff/mp4"
)

func RewriteDuration(ifilename string, ofile *os.File, audioDuration, videoDuration uint64) error {
	ifile, err := mp4.ReadMP4File(ifilename)
	if err != nil {
		return err
	}

	duration := audioDuration
	if videoDuration > audioDuration {
		duration = videoDuration
	}
	ifile.Moov.Mvhd.Duration = duration

	for i := 0; i < len(ifile.Moov.Traks); i++ {
		if ifile.Moov.Traks[i].Tkhd == nil {
			continue
		}
		if ifile.Moov.Traks[i].Tkhd.Width != 0 {
			ifile.Moov.Traks[i].Tkhd.Duration = videoDuration
			if ifile.Moov.Traks[i].Edts == nil {
				continue
			} else if len(ifile.Moov.Traks[i].Edts.Elst) == 0 {
				continue
			} else if len(ifile.Moov.Traks[i].Edts.Elst[0].Entries) == 0 {
				continue
			}
			numEntries := len(ifile.Moov.Traks[i].Edts.Elst[0].Entries)
			ifile.Moov.Traks[i].Edts.Elst[0].Entries[numEntries-1].SegmentDuration = videoDuration
		} else {
			ifile.Moov.Traks[i].Tkhd.Duration = audioDuration
			if ifile.Moov.Traks[i].Edts == nil {
				continue
			} else if len(ifile.Moov.Traks[i].Edts.Elst) == 0 {
				continue
			} else if len(ifile.Moov.Traks[i].Edts.Elst[0].Entries) == 0 {
				continue
			}
			ifile.Moov.Traks[i].Edts.Elst[0].Entries[0].SegmentDuration = audioDuration
		}
	}

	return ifile.Encode(ofile)
}
