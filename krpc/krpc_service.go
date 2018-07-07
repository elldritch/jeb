package krpc

import (
	"errors"

	"github.com/golang/protobuf/proto"

	"github.com/ilikebits/jeb/krpc/pb"
)

type Point struct {
	X, Y float64
}

type Vector struct {
	X, Y, Z float64
}

type Quaternion struct {
	A, B, C, D float64
}

type BoundingBox struct {
	Min, Max Vector
}

func (c *Client) GetStatus() (pb.Status, error) {
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
