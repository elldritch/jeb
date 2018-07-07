package krpc

import (
	"encoding/binary"
	"errors"
	"log"

	"github.com/golang/protobuf/proto"
)

func (c *Conn) Send(msg proto.Message) (int, error) {
	log.Printf("Send: %#v", msg)

	data, err := proto.Marshal(msg)
	if err != nil {
		return 0, err
	}
	m, err := c.conn.Write(proto.EncodeVarint(uint64(len(data))))
	if err != nil {
		return m, err
	}
	n, err := c.conn.Write(data)
	if err != nil {
		return m + n, err
	}
	return m + n, nil
}

func (c *Conn) Read(msg proto.Message) error {
	log.Printf("Read: %#v", msg)

	// Read varint-encoded message size.
	msglen, err := binary.ReadUvarint(c)
	if err != nil {
		return err
	}
	log.Printf("msglen: %#v", msglen)

	// Read message.
	buf := make([]byte, msglen)
	_, err = c.conn.Read(buf)
	if err != nil {
		return err
	}

	// Decode message contents.
	log.Printf("proto.Unmarshal(%#v, %#v)", buf, msg)
	err = proto.Unmarshal(buf, msg)
	if err != nil {
		return err
	}
	log.Println("done unmarshalling buffer")

	return nil
}

func (c *Conn) ReadByte() (byte, error) {
	log.Println("ReadByte()")

	b := make([]byte, 1)
	n, err := c.conn.Read(b)
	if err != nil {
		return 0, err
	}
	if n != 1 {
		return 0, errors.New("ReadByte: read wrong length")
	}
	log.Printf("b: %#v", b)
	return b[0], nil
}
