package krpc

import (
	"errors"

	"github.com/golang/protobuf/proto"

	"github.com/ilikebits/jeb/krpc/pb"
)

type KRPC struct {
	conn *Conn
}

func (k *KRPC) GetStatus() (pb.Status, error) {
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "KRPC",
			Procedure: "GetStatus",
		}},
	}
	_, err := k.conn.Send(&req)
	if err != nil {
		return pb.Status{}, err
	}

	res := pb.Response{}
	err = k.conn.Read(&res)
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
