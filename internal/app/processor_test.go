package app_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/brianvoe/gofakeit/v6"
	"github.com/guilherme-daniel-rs/event-processor/internal/app"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockEventRepository struct {
	mock.Mock
}

func (m *MockEventRepository) Save(ctx context.Context, event models.EventRecord) error {
	args := m.Called(ctx, event)
	return args.Error(0)
}

func generateValidBody() ([]byte, map[string]any) {
	bodyMap := map[string]any{
		"user_id":  gofakeit.UUID(),
		"email":    gofakeit.Email(),
		"name":     gofakeit.Name(),
		"role":     gofakeit.JobTitle(),
		"verified": gofakeit.Bool(),
	}
	bodyBytes, _ := json.Marshal(bodyMap)
	return bodyBytes, bodyMap
}

func createMessage(header app.MessageHeader) ports.Message {
	body, _ := json.Marshal(header)
	return ports.Message{
		ID:   gofakeit.UUID(),
		Body: body,
	}
}

func TestProcessor_Process(t *testing.T) {
	t.Run("success processing", func(t *testing.T) {
		repo := new(MockEventRepository)
		processor := app.NewProcessor(repo)

		validBodyBytes, _ := generateValidBody()
		eventID := gofakeit.UUID()

		header := app.MessageHeader{
			EventID:       eventID,
			EventType:     "user.created",
			SchemaVersion: "v1",
			TenantID:      gofakeit.UUID(),
			ClientID:      gofakeit.UUID(),
			OccurredAt:    time.Now().Format(time.RFC3339),
			Body:          json.RawMessage(validBodyBytes),
		}

		repo.On("Save", mock.Anything, mock.MatchedBy(func(e models.EventRecord) bool {
			var expected map[string]any
			var actual map[string]any
			_ = json.Unmarshal(validBodyBytes, &expected)
			_ = json.Unmarshal([]byte(e.Body), &actual)

			return e.EventID == eventID &&
				e.Status == "processed" &&
				assert.ObjectsAreEqual(expected, actual)
		})).Return(nil)

		err := processor.Process(context.Background(), createMessage(header))
		assert.NoError(t, err)
		repo.AssertExpectations(t)
	})

	t.Run("invalid header structure", func(t *testing.T) {
		repo := new(MockEventRepository)
		processor := app.NewProcessor(repo)

		msg := ports.Message{Body: []byte(`{invalid-json}`)}

		err := processor.Process(context.Background(), msg)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to unmarshal message header")
		repo.AssertNotCalled(t, "Save")
	})

	t.Run("incomplete header fields", func(t *testing.T) {
		repo := new(MockEventRepository)
		processor := app.NewProcessor(repo)

		validBodyBytes, _ := generateValidBody()

		header := app.MessageHeader{
			EventType:     "user.created",
			SchemaVersion: "v1",
			Body:          json.RawMessage(validBodyBytes),
		}

		repo.On("Save", mock.Anything, mock.MatchedBy(func(e models.EventRecord) bool {
			return e.Status == "failed"
		})).Return(nil)

		err := processor.Process(context.Background(), createMessage(header))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid message header")
		repo.AssertExpectations(t)
	})

	t.Run("schema validation failure", func(t *testing.T) {
		repo := new(MockEventRepository)
		processor := app.NewProcessor(repo)

		invalidBody, _ := json.Marshal(map[string]any{"user_id": gofakeit.UUID()})
		eventID := gofakeit.UUID()

		header := app.MessageHeader{
			EventID:       eventID,
			EventType:     "user.created",
			SchemaVersion: "v1",
			TenantID:      gofakeit.UUID(),
			ClientID:      gofakeit.UUID(),
			OccurredAt:    time.Now().Format(time.RFC3339),
			Body:          json.RawMessage(invalidBody),
		}

		repo.On("Save", mock.Anything, mock.MatchedBy(func(e models.EventRecord) bool {
			return e.EventID == eventID && e.Status == "failed"
		})).Return(nil)

		err := processor.Process(context.Background(), createMessage(header))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema validation failed")
		repo.AssertExpectations(t)
	})

	t.Run("repository save error", func(t *testing.T) {
		repo := new(MockEventRepository)
		processor := app.NewProcessor(repo)

		validBodyBytes, _ := generateValidBody()

		header := app.MessageHeader{
			EventID:       gofakeit.UUID(),
			EventType:     "user.created",
			SchemaVersion: "v1",
			TenantID:      gofakeit.UUID(),
			ClientID:      gofakeit.UUID(),
			OccurredAt:    time.Now().Format(time.RFC3339),
			Body:          json.RawMessage(validBodyBytes),
		}

		repo.On("Save", mock.Anything, mock.Anything).Return(errors.New("db error"))

		err := processor.Process(context.Background(), createMessage(header))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "failed to save event")
	})
}
