package krpc

type Client struct {
	conn *Conn

	KRPC KRPC
}

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

func Dial(addr string) (*Client, error) {
	conn, err := Connect(addr)
	if err != nil {
		return nil, err
	}

	client := Client{
		conn: conn,

		KRPC: KRPC{
			conn: conn,
		},
	}

	return &client, nil
}
