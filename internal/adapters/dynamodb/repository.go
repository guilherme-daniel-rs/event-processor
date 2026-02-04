package dynamodb

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
)

type EventRepository struct {
	client *dynamodb.Client
}

func NewEventRepository(client *dynamodb.Client) *EventRepository {
	return &EventRepository{
		client: client,
	}
}

func (r *EventRepository) Save(ctx context.Context, event models.EventRecord) error {
	item, err := attributevalue.MarshalMap(event)
	if err != nil {
		return err
	}

	_, err = r.client.PutItem(ctx, &dynamodb.PutItemInput{
		TableName: event.TableName(),
		Item:      item,
	})

	return err
}
