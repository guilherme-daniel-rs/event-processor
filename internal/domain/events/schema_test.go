package events_test

import (
	"testing"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/events"
	"github.com/stretchr/testify/assert"
)

func TestSchemaRegistry(t *testing.T) {
	registry := events.NewSchemaRegistry()

	t.Run("should unmarshal valid user.created event", func(t *testing.T) {
		data := []byte(`{
			"user_id": "123",
			"email": "test@example.com",
			"name": "Test User",
			"role": "admin",
			"verified": true
		}`)

		schema, err := registry.Unmarshal("user.created", "v1", data)
		assert.NoError(t, err)
		assert.NotNil(t, schema)

		event, ok := schema.(*events.UserCreatedV1)
		assert.True(t, ok)
		assert.Equal(t, "123", event.UserID)
	})

	t.Run("should unmarshal valid order.placed event", func(t *testing.T) {
		data := []byte(`{
			"order_id": "ord-1",
			"user_id": "u-1",
			"total": 100.50,
			"items_count": 2,
			"status": "confirmed"
		}`)

		schema, err := registry.Unmarshal("order.placed", "v1", data)
		assert.NoError(t, err)
		assert.NotNil(t, schema)

		event, ok := schema.(*events.OrderPlacedV1)
		assert.True(t, ok)
		assert.Equal(t, "ord-1", event.OrderID)
	})

	t.Run("should unmarshal valid payment.processed event", func(t *testing.T) {
		data := []byte(`{
			"payment_id": "pay-1",
			"order_id": "ord-1",
			"amount": 100.50,
			"payment_method": "credit_card",
			"status": "success"
		}`)

		schema, err := registry.Unmarshal("payment.processed", "v1", data)
		assert.NoError(t, err)
		assert.NotNil(t, schema)

		event, ok := schema.(*events.PaymentProcessedV1)
		assert.True(t, ok)
		assert.Equal(t, "pay-1", event.PaymentID)
	})

	t.Run("should fail for unknown event type", func(t *testing.T) {
		_, err := registry.Unmarshal("unknown.event", "v1", []byte("{}"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown event type")
	})

	t.Run("should fail for unknown version", func(t *testing.T) {
		_, err := registry.Unmarshal("user.created", "v99", []byte("{}"))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "unknown schema version")
	})

	t.Run("should fail validation on unmarshal", func(t *testing.T) {
		data := []byte(`{"user_id": "123"}`)
		_, err := registry.Unmarshal("user.created", "v1", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema validation failed")
	})

	t.Run("should fail on invalid json", func(t *testing.T) {
		data := []byte(`{invalid-json}`)
		_, err := registry.Unmarshal("user.created", "v1", data)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal")
	})
}

func TestUserCreatedV1_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   events.UserCreatedV1
		wantErr bool
	}{
		{
			name: "valid event",
			event: events.UserCreatedV1{
				UserID: gofakeit.UUID(),
				Email:  gofakeit.Email(),
				Name:   gofakeit.Name(),
				Role:   gofakeit.JobTitle(),
			},
			wantErr: false,
		},
		{
			name:    "missing user_id",
			event:   events.UserCreatedV1{Email: gofakeit.Email(), Name: gofakeit.Name(), Role: gofakeit.JobTitle()},
			wantErr: true,
		},
		{
			name:    "missing email",
			event:   events.UserCreatedV1{UserID: gofakeit.UUID(), Name: gofakeit.Name(), Role: gofakeit.JobTitle()},
			wantErr: true,
		},
		{
			name:    "missing name",
			event:   events.UserCreatedV1{UserID: gofakeit.UUID(), Email: gofakeit.Email(), Role: gofakeit.JobTitle()},
			wantErr: true,
		},
		{
			name:    "missing role",
			event:   events.UserCreatedV1{UserID: gofakeit.UUID(), Email: gofakeit.Email(), Name: gofakeit.Name()},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestOrderPlacedV1_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   events.OrderPlacedV1
		wantErr bool
	}{
		{
			name: "valid event",
			event: events.OrderPlacedV1{
				OrderID:    gofakeit.UUID(),
				UserID:     gofakeit.UUID(),
				Total:      gofakeit.Price(10, 1000),
				ItemsCount: gofakeit.Number(1, 10),
				Status:     "PENDING",
			},
			wantErr: false,
		},
		{
			name:    "missing order_id",
			event:   events.OrderPlacedV1{UserID: gofakeit.UUID(), Total: 100, ItemsCount: 1, Status: "PENDING"},
			wantErr: true,
		},
		{
			name:    "invalid total",
			event:   events.OrderPlacedV1{OrderID: gofakeit.UUID(), UserID: gofakeit.UUID(), Total: 0, ItemsCount: 1, Status: "PENDING"},
			wantErr: true,
		},
		{
			name:    "invalid items count",
			event:   events.OrderPlacedV1{OrderID: gofakeit.UUID(), UserID: gofakeit.UUID(), Total: 100, ItemsCount: 0, Status: "PENDING"},
			wantErr: true,
		},
		{
			name:    "missing user_id",
			event:   events.OrderPlacedV1{OrderID: gofakeit.UUID(), Total: 100, ItemsCount: 1, Status: "PENDING"},
			wantErr: true,
		},
		{
			name:    "missing status",
			event:   events.OrderPlacedV1{OrderID: gofakeit.UUID(), UserID: gofakeit.UUID(), Total: 100, ItemsCount: 1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}

func TestPaymentProcessedV1_Validate(t *testing.T) {
	tests := []struct {
		name    string
		event   events.PaymentProcessedV1
		wantErr bool
	}{
		{
			name: "valid event",
			event: events.PaymentProcessedV1{
				PaymentID:     gofakeit.UUID(),
				OrderID:       gofakeit.UUID(),
				Amount:        gofakeit.Price(10, 500),
				PaymentMethod: "CREDIT_CARD",
				Status:        "SUCCESS",
			},
			wantErr: false,
		},
		{
			name:    "missing payment_id",
			event:   events.PaymentProcessedV1{OrderID: gofakeit.UUID(), Amount: 100, PaymentMethod: "CREDIT_CARD", Status: "SUCCESS"},
			wantErr: true,
		},
		{
			name:    "invalid amount",
			event:   events.PaymentProcessedV1{PaymentID: gofakeit.UUID(), OrderID: gofakeit.UUID(), Amount: 0, PaymentMethod: "CREDIT_CARD", Status: "SUCCESS"},
			wantErr: true,
		},
		{
			name:    "missing order_id",
			event:   events.PaymentProcessedV1{PaymentID: gofakeit.UUID(), Amount: 100, PaymentMethod: "CREDIT_CARD", Status: "SUCCESS"},
			wantErr: true,
		},
		{
			name:    "missing payment_method",
			event:   events.PaymentProcessedV1{PaymentID: gofakeit.UUID(), OrderID: gofakeit.UUID(), Amount: 100, Status: "SUCCESS"},
			wantErr: true,
		},
		{
			name:    "missing status",
			event:   events.PaymentProcessedV1{PaymentID: gofakeit.UUID(), OrderID: gofakeit.UUID(), Amount: 100, PaymentMethod: "CREDIT_CARD"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.event.Validate()
			if tt.wantErr {
				assert.Error(t, err)
				return
			}
			assert.NoError(t, err)
		})
	}
}
