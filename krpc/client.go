package krpc

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
