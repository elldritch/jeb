package main

import (
	"flag"
	"log"

	"github.com/golang/protobuf/proto"

	"github.com/ilikebits/jeb/krpc"
	"github.com/ilikebits/jeb/krpc/pb"
)

func main() {
	// Parse flags.
	addr := flag.String("addr", "127.0.0.1:50000", "server TCP address")
	flag.Parse()

	// Open connection.
	log.Println("opening connection")
	conn, err := krpc.Connect(*addr)
	if err != nil {
		panic(err)
	}
	defer conn.Close()
	log.Println("opened connection:", conn.ID())

	// Call KRPC.GetStatus()
	req := pb.Request{
		Calls: []*pb.ProcedureCall{&pb.ProcedureCall{
			Service:   "KRPC",
			Procedure: "GetStatus",
		}},
	}
	_, err = conn.Send(&req)
	if err != nil {
		panic(err)
	}

	res := pb.Response{}
	err = conn.Read(&res)
	if err != nil {
		panic(err)
	}
	if e := res.GetError(); e != nil {
		panic(e.GetDescription())
	}
	statBytes := res.GetResults()[0].GetValue()
	status := pb.Status{}
	err = proto.Unmarshal(statBytes, &status)
	if err != nil {
		panic(err)
	}
	log.Printf("%#v", status)

	// Spin to keep the connection open.
	select {}
}
