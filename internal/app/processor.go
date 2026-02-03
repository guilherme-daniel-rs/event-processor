package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Processor struct {
}

func NewProcessor() *Processor {
	return &Processor{}
}

type MessageHeader struct {
	EventID       string `json:"event_id"`
	EventType     string `json:"event_type"`
	TenantID      string `json:"tenant_id"`
	ClientID      string `json:"client_id"`
	SchemaVersion string `json:"schema_version"`
	OccurredAt    string `json:"occurred_at"`
	Body          []byte `json:"body"`
}

func (m MessageHeader) IsValid() bool {
	return m.EventID != "" &&
		m.EventType != "" &&
		m.TenantID != "" &&
		m.ClientID != "" &&
		m.SchemaVersion != "" &&
		m.OccurredAt != "" &&
		len(m.Body) > 0
}

func (p *Processor) Process(ctx context.Context, msg ports.Message) error {
	fmt.Println("Processing message")

	payload := MessageHeader{}
	err := json.Unmarshal(msg.Body, &payload)
	if err != nil {
		return err
	}

	if !payload.IsValid() {
		return fmt.Errorf("invalid message header")
	}

	return nil
}
