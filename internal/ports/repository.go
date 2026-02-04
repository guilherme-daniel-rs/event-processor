package ports

import (
	"context"

	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
)

type EventRepository interface {
	Save(ctx context.Context, event models.EventRecord) error
}
