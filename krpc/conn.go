package krpc

import (
	"log"
	"net"

	"github.com/pkg/errors"

	"github.com/ilikebits/jeb/krpc/pb"
)

type Conn struct {
	id   []byte
	conn net.Conn
}

func (c *Conn) ID() []byte {
	return c.id
}

func (c *Conn) Close() error {
	return c.conn.Close()
}

func Connect(addr string) (*Conn, error) {
	// Open connection.
	conn, err := net.Dial("tcp", addr)
	if err != nil {
		return nil, err
	}

	// Construct connection.
	c := Conn{
		conn: conn,
	}

	// Make connection request.
	req := pb.ConnectionRequest{
		Type:             pb.ConnectionRequest_RPC,
		ClientName:       "jeb",
		ClientIdentifier: []byte{},
	}
	_, err = c.Send(&req)
	if err != nil {
		return nil, err
	}

	// Read connection response.
	log.Println("reading connection response")
	res := pb.ConnectionResponse{}
	err = c.Read(&res)
	if err != nil {
		return nil, err
	}
	log.Println("done reading connection response")

	// Parse connection response.
	if res.GetStatus() != pb.ConnectionResponse_OK {
		log.Println("bad connection response")
		return nil, errors.Errorf("bad connection response: %#s", res.GetMessage())
	}
	c.id = res.GetClientIdentifier()

	return &c, nil
}
