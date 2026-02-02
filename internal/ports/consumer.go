package ports

import "context"

type Message struct {
	ID           string
	Body         []byte
	Attributes   map[string]string
	AckToken     string
	ReceiveCount int
}

type NackOptions struct {
	DelayBeforeRetrySeconds int32
}

type Consumer interface {
	Receive(ctx context.Context) ([]Message, error)
	Ack(ctx context.Context, msg Message) error
	Nack(ctx context.Context, msg Message, opts NackOptions) error
}
