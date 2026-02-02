package ports

import "context"

type Message struct {
	ID           string
	Body         []byte
	Attributes   map[string]string
	AckToken     string
	ReceiveCount int
}

type Consumer interface {
	Read(ctx context.Context, process func(ctx context.Context, msg Message) error) error
}
