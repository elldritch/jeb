package krpc

import (
	"errors"
	"log"

	"github.com/golang/protobuf/proto"
)

func (c *Conn) Send(msg proto.Message) (int, error) {
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
	// Read initial buffer.
	buf := make([]byte, 256)
	buflen, err := c.conn.Read(buf)
	if err != nil {
		return err
	}
	log.Println("read buffer length:", buflen)

	// Decode message size.
	msglen, n := proto.DecodeVarint(buf)
	buf = buf[n:]
	log.Println("read message length:", msglen)

	// Read remaining message if needed.
	//
	// TODO: what do we do if we read _too many_ messages? What if we read past
	// the first message?
	//
	// I think the answer here is to always parse messages and store parsed
	// messages in a buffer. Since responses are always guaranteed to be in the
	// order of requests, we can always pop off the latest response.
	//
	// > Requests are processed in order of receipt. The next request from a
	// > client will not be processed until the previous one completes execution
	// > and it’s response has been received by the client. When there are
	// > multiple client connections, requests are processed in round-robin order.
	// > - https://krpc.github.io/krpc/communication-protocols/messages.html#invoking-remote-procedures
	remaining := (int(msglen) + n) - buflen
	log.Println("remaining messages bytes:", remaining)
	if remaining > 0 {
		rembuf := make([]byte, remaining)
		rembuflen, err := c.conn.Read(rembuf)
		if err != nil {
			return err
		}
		if rembuflen != remaining {
			return errors.New("read fewer bytes than expecting")
		}
		buf = append(buf, rembuf...)
	}

	// Decode message contents.
	log.Println("unmarshalling buffer")
	err = proto.Unmarshal(buf[:msglen], msg)
	if err != nil {
		return err
	}
	log.Println("done unmarshalling buffer")

	return nil
}
