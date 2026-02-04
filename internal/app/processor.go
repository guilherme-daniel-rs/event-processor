package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/events"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Processor struct {
	schemaRegistry *events.SchemaRegistry
	repository     ports.EventRepository
}

func NewProcessor(repository ports.EventRepository) *Processor {
	return &Processor{
		schemaRegistry: events.NewSchemaRegistry(),
		repository:     repository,
	}
}

type MessageHeader struct {
	EventID       string          `json:"event_id"`
	EventType     string          `json:"event_type"`
	TenantID      string          `json:"tenant_id"`
	ClientID      string          `json:"client_id"`
	SchemaVersion string          `json:"schema_version"`
	OccurredAt    string          `json:"occurred_at"`
	Body          json.RawMessage `json:"body"`
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
		return ports.NewNonRetriableError(fmt.Errorf("failed to unmarshal message header: %w", err))
	}

	record := models.EventRecord{
		ID:            uuid.New().String(),
		EventID:       header.EventID,
		TenantID:      header.TenantID,
		ClientID:      header.ClientID,
		SchemaVersion: header.SchemaVersion,
		OccurredAt:    header.OccurredAt,
		Status:        "processed",
		Body:          string(header.Body),
	}

	if !header.IsValid() {
		record.Status = "failed"
		if err := p.repository.Save(ctx, record); err != nil {
			return fmt.Errorf("failed to save event to repository: %w", err)
		}
		return ports.NewNonRetriableError(fmt.Errorf("invalid message header"))
	}

	_, err := p.schemaRegistry.Unmarshal(header.EventType, header.SchemaVersion, header.Body)
	if err != nil {
		record.Status = "failed"
		if err := p.repository.Save(ctx, record); err != nil {
			return fmt.Errorf("failed to save event to repository: %w", err)
		}
		return ports.NewNonRetriableError(fmt.Errorf("failed to unmarshal event body: %w", err))
	}

	if err := p.repository.Save(ctx, record); err != nil {
		return fmt.Errorf("failed to save event to repository: %w", err)
	}

	return nil
}
