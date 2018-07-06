package krpc

import "errors"

type StreamClient struct{}

func DialStream(addr string) (*StreamClient, error) {
	return nil, errors.New("not implemented")
}
