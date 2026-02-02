package app

import (
	"context"
	"fmt"

	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Processor struct {
}

func NewProcessor() *Processor {
	return &Processor{}
}

func (p *Processor) Process(ctx context.Context, msg ports.Message) error {
	fmt.Println("Processing message", msg)

	return nil
}
