package krpc

import (
	"errors"

	"github.com/ilikebits/jeb/krpc/pb"
)

type Client struct {
	conn *Conn
}

func Dial(addr string) (*Client, error) {
	conn, err := Connect(addr)
	if err != nil {
		return nil, err
	}

	client := Client{
		conn: conn,
	}

	return &client, nil
}

func (c *Client) Status() (pb.Status, error) {
	return pb.Status{}, errors.New("not implemented")
}
