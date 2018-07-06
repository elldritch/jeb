package krpc

import (
	"errors"

	"github.com/golang/protobuf/proto"

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
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "KRPC",
			Procedure: "GetStatus",
		}},
	}
	_, err := c.conn.Send(&req)
	if err != nil {
		return pb.Status{}, err
	}

	res := pb.Response{}
	err = c.conn.Read(&res)
	if err != nil {
		return pb.Status{}, err
	}
	if e := res.GetError(); e != nil {
		return pb.Status{}, errors.New(e.GetDescription())
	}
	statBytes := res.GetResults()[0].GetValue()
	status := pb.Status{}
	err = proto.Unmarshal(statBytes, &status)
	if err != nil {
		return pb.Status{}, err
	}

	return status, nil
}
