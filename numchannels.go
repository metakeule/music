package music

import "github.com/mkb218/gosndfile/sndfile"

func numChannels(file string) (int, error) {
	var info sndfile.Info
	_, err := sndfile.Open(file, sndfile.Read, &info)
	if err != nil {
		return -1, err
	}
	return int(info.Channels), nil
}
