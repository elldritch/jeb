package krpc

import (
	"errors"
	"log"
	"math"

	"github.com/golang/protobuf/proto"

	"github.com/ilikebits/jeb/krpc/pb"
)

func (c *Client) Vessel() (*Vessel, error) {
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "SpaceCenter",
			Procedure: "get_ActiveVessel",
		}},
	}
	_, err := c.conn.Send(&req)
	if err != nil {
		return nil, err
	}

	res := pb.Response{}
	err = c.conn.Read(&res)
	if err != nil {
		return nil, err
	}
	if e := res.GetError(); e != nil {
		return nil, errors.New(e.GetDescription())
	}
	idBytes := res.GetResults()[0].GetValue()
	log.Printf("parsing vessel ID: %#v", idBytes)
	id, err := proto.NewBuffer(idBytes).DecodeVarint()
	if err != nil {
		return nil, err
	}

	log.Println("got vessel")
	return &Vessel{
		id: id,
		c:  c,
	}, nil
}

type Vessel struct {
	c  *Client
	id uint64
}

func (v *Vessel) Flight() (*Flight, error) {
	vProto := proto.EncodeVarint(v.id)
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "SpaceCenter",
			Procedure: "Vessel_Flight",
			Arguments: []*pb.Argument{
				&pb.Argument{
					Position: 0,
					Value:    vProto,
				},
			},
		}},
	}
	_, err := v.c.conn.Send(&req)
	if err != nil {
		return nil, err
	}

	res := pb.Response{}
	err = v.c.conn.Read(&res)
	if err != nil {
		return nil, err
	}
	if e := res.GetError(); e != nil {
		return nil, errors.New(e.GetDescription())
	}
	idBytes := res.GetResults()[0].GetValue()
	log.Printf("parsing flight: %#v", idBytes)
	id, err := proto.NewBuffer(idBytes).DecodeVarint()
	if err != nil {
		return nil, err
	}

	log.Println("got vessel")
	return &Flight{
		c:  v.c,
		id: id,
	}, nil
}

type Flight struct {
	c  *Client
	id uint64
}

func (f *Flight) SurfaceAltitude() (float64, error) {
	fProto := proto.EncodeVarint(f.id)
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "SpaceCenter",
			Procedure: "Flight_get_SurfaceAltitude",
			Arguments: []*pb.Argument{
				&pb.Argument{
					Position: 0,
					Value:    fProto,
				},
			},
		}},
	}
	_, err := f.c.conn.Send(&req)
	if err != nil {
		return math.NaN(), err
	}

	res := pb.Response{}
	err = f.c.conn.Read(&res)
	if err != nil {
		return math.NaN(), err
	}
	if e := res.GetError(); e != nil {
		return math.NaN(), errors.New(e.GetDescription())
	}
	altitudeBytes := res.GetResults()[0].GetValue()
	log.Printf("parsing altitude: %#v", altitudeBytes)
	altitude, err := proto.NewBuffer(altitudeBytes).DecodeFixed64()
	if err != nil {
		return math.NaN(), err
	}

	log.Println("got vessel")
	return math.Float64frombits(altitude), nil
}
