package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/events"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
	"github.com/guilherme-daniel-rs/event-processor/internal/logging"
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
	logging.Append(ctx, "Header unmarshaled successfully for event %s", header.EventID)

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
	logging.Append(ctx, "Record initialized for tenant %s", record.TenantID)

	if !header.IsValid() {
		logging.Append(ctx, "Header validation failed")
		return p.saveFailure(ctx, record, fmt.Errorf("invalid message header"))
	}
	logging.Append(ctx, "Header validated")

	_, err := p.schemaRegistry.Unmarshal(header.EventType, header.SchemaVersion, header.Body)
	if err != nil {
		logging.Append(ctx, "Body unmarshal failed: %v", err)
		return p.saveFailure(ctx, record, fmt.Errorf("failed to unmarshal event body: %w", err))
	}
	logging.Append(ctx, "Body unmarshaled and validated via registry")

	if err := p.repository.Save(ctx, record); err != nil {
		return fmt.Errorf("failed to save event to repository: %w", err)
	}
	logging.Append(ctx, "Event saved successfully to repository")

	return nil
}

func (p *Processor) saveFailure(ctx context.Context, record models.EventRecord, err error) error {
	record.Status = "failed"
	if saveErr := p.repository.Save(ctx, record); saveErr != nil {
		return fmt.Errorf("failed to save failed event record: %w (original error: %v)", saveErr, err)
	}
	return ports.NewNonRetriableError(err)
}
