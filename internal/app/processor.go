package app

import (
	"context"
	"fmt"

	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Processor struct {
	consumer ports.Consumer
}

func NewProcessor(consumer ports.Consumer) *Processor {
	return &Processor{
		consumer: consumer,
	}
}

func (p *Processor) Run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			messages, err := p.consumer.Receive(ctx)
			if err != nil {
				fmt.Println("receive failed:", err)
				continue
			}

			for _, msg := range messages {
				fmt.Println("Processing message ID:", msg.ID)
			}
		}
	}
}
