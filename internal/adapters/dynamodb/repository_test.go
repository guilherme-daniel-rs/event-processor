package dynamodb

import (
	"context"
	"errors"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/guilherme-daniel-rs/event-processor/internal/domain/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockDynamoDBClient struct {
	mock.Mock
}

func (m *MockDynamoDBClient) PutItem(ctx context.Context, params *dynamodb.PutItemInput, optFns ...func(*dynamodb.Options)) (*dynamodb.PutItemOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*dynamodb.PutItemOutput), args.Error(1)
}

func TestEventRepository_Save(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockDynamoDBClient)
		repo := NewEventRepository(mockClient)

		event := models.EventRecord{
			ID: "123",
		}

		mockClient.On("PutItem",
			mock.Anything,
			mock.MatchedBy(func(input *dynamodb.PutItemInput) bool {
				return *input.TableName == "events"
			}),
			mock.Anything,
		).Return(&dynamodb.PutItemOutput{}, nil)

		err := repo.Save(context.Background(), event)
		assert.NoError(t, err)
		mockClient.AssertExpectations(t)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := new(MockDynamoDBClient)
		repo := NewEventRepository(mockClient)

		event := models.EventRecord{
			ID: "123",
		}

		mockClient.On("PutItem", mock.Anything, mock.Anything, mock.Anything).
			Return(nil, errors.New("dynamodb error"))

		err := repo.Save(context.Background(), event)
		assert.Error(t, err)
		assert.Equal(t, "dynamodb error", err.Error())
		mockClient.AssertExpectations(t)
	})
}
