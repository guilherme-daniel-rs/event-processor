package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/guilherme-daniel-rs/event-processor/internal/domain/events"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Processor struct {
	schemaRegistry *events.SchemaRegistry
}

func NewProcessor() *Processor {
	return &Processor{
		schemaRegistry: events.NewSchemaRegistry(),
	}
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
	var header MessageHeader
	if err := json.Unmarshal(msg.Body, &header); err != nil {
		return fmt.Errorf("failed to unmarshal message header: %w", err)
	}

	if !header.IsValid() {
		return fmt.Errorf("invalid message header")
	}

	_, err := p.schemaRegistry.Unmarshal(header.EventType, header.SchemaVersion, header.Body)
	if err != nil {
		return fmt.Errorf("failed to unmarshal event body: %w", err)
	}

	return nil
}
