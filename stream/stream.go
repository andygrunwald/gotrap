package stream

import (
	"errors"
	"github.com/andygrunwald/gotrap/config"
)

const (
	StreamAmqp = iota
)

var Streams = make(map[int]Stream, 1)

type Stream interface {
	Initialize(*config.Configuration)
	Start() error
}

func GetStream(streamType int) (Stream, error) {
	if val, ok := Streams[streamType]; ok {
		return val, nil
	}

	return nil, errors.New("Stream not found")
}
