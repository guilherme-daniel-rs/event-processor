package events

import (
	"encoding/json"
	"fmt"
)

type Schema interface {
	Validate() error
}

type SchemaRegistry struct {
	schemas map[string]map[string]func() Schema
}

func NewSchemaRegistry() *SchemaRegistry {
	registry := &SchemaRegistry{
		schemas: make(map[string]map[string]func() Schema),
	}

	registry.Register("payment.processed", "v1", func() Schema { return &PaymentProcessedV1{} })
	registry.Register("user.created", "v1", func() Schema { return &UserCreatedV1{} })
	registry.Register("order.placed", "v1", func() Schema { return &OrderPlacedV1{} })

	return registry
}

func (r *SchemaRegistry) Register(eventType, version string, constructor func() Schema) {
	if r.schemas[eventType] == nil {
		r.schemas[eventType] = make(map[string]func() Schema)
	}
	r.schemas[eventType][version] = constructor
}

func (r *SchemaRegistry) Unmarshal(eventType, version string, data []byte) (Schema, error) {
	versions, ok := r.schemas[eventType]
	if !ok {
		return nil, fmt.Errorf("unknown event type: %s", eventType)
	}

	constructor, ok := versions[version]
	if !ok {
		return nil, fmt.Errorf("unknown schema version %s for event type %s", version, eventType)
	}

	schema := constructor()
	if err := json.Unmarshal(data, schema); err != nil {
		return nil, fmt.Errorf("failed to unmarshal event body: %w", err)
	}

	if err := schema.Validate(); err != nil {
		return nil, fmt.Errorf("schema validation failed: %w", err)
	}

	return schema, nil
}
