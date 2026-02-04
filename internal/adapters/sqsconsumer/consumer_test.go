package sqsconsumer

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockSQSClient struct {
	mock.Mock
}

func (m *MockSQSClient) ReceiveMessage(ctx context.Context, params *sqs.ReceiveMessageInput, optFns ...func(*sqs.Options)) (*sqs.ReceiveMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ReceiveMessageOutput), args.Error(1)
}

func (m *MockSQSClient) DeleteMessage(ctx context.Context, params *sqs.DeleteMessageInput, optFns ...func(*sqs.Options)) (*sqs.DeleteMessageOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.DeleteMessageOutput), args.Error(1)
}

func (m *MockSQSClient) ChangeMessageVisibility(ctx context.Context, params *sqs.ChangeMessageVisibilityInput, optFns ...func(*sqs.Options)) (*sqs.ChangeMessageVisibilityOutput, error) {
	args := m.Called(ctx, params, optFns)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*sqs.ChangeMessageVisibilityOutput), args.Error(1)
}

func TestConsumer_Receive(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		mockClient := new(MockSQSClient)
		consumer := NewSqsConsumer(mockClient, Options{
			QueueURL:    "test-queue",
			MaxMessages: 10,
			WaitTimeSec: 20,
		})

		mockClient.On("ReceiveMessage", mock.Anything, mock.MatchedBy(func(input *sqs.ReceiveMessageInput) bool {
			return *input.QueueUrl == "test-queue" && input.MaxNumberOfMessages == 10
		}), mock.Anything).Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					MessageId:     aws.String("msg-1"),
					Body:          aws.String("body"),
					ReceiptHandle: aws.String("handle"),
				},
			},
		}, nil)

		msgs, err := consumer.Receive(context.Background())
		assert.NoError(t, err)
		assert.Len(t, msgs, 1)
		assert.Equal(t, "msg-1", msgs[0].ID)
	})

	t.Run("error", func(t *testing.T) {
		mockClient := new(MockSQSClient)
		consumer := NewSqsConsumer(mockClient, Options{})

		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(nil, errors.New("sqs error"))

		msgs, err := consumer.Receive(context.Background())
		assert.Error(t, err)
		assert.Nil(t, msgs)
	})

	t.Run("max_retries_reached_dlq_trigger", func(t *testing.T) {
		mockClient := new(MockSQSClient)
		consumer := NewSqsConsumer(mockClient, Options{QueueURL: "test-queue", MaxRetries: 3})

		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(&sqs.ReceiveMessageOutput{
			Messages: []types.Message{
				{
					MessageId:     aws.String("msg-1"),
					ReceiptHandle: aws.String("handle-1"),
					Attributes: map[string]string{
						"ApproximateReceiveCount": "3",
					},
				},
			},
		}, nil).Once()

		mockClient.On("ChangeMessageVisibility", mock.Anything, mock.MatchedBy(func(input *sqs.ChangeMessageVisibilityInput) bool {
			return *input.ReceiptHandle == "handle-1" && input.VisibilityTimeout == 0
		}), mock.Anything).Return(&sqs.ChangeMessageVisibilityOutput{}, nil).Once()

		mockClient.On("ReceiveMessage", mock.Anything, mock.Anything, mock.Anything).Return(&sqs.ReceiveMessageOutput{}, nil).Maybe()

		ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
		defer cancel()

		_ = consumer.Read(ctx, func(ctx context.Context, msg ports.Message) error {
			return errors.New("persistent failure")
		})

		mockClient.AssertExpectations(t)
	})
}

func TestConsumer_Ack(t *testing.T) {
	mockClient := new(MockSQSClient)
	consumer := NewSqsConsumer(mockClient, Options{QueueURL: "test-queue"})
	msg := ports.Message{AckToken: "handle-1"}

	mockClient.On("DeleteMessage", mock.Anything, mock.MatchedBy(func(input *sqs.DeleteMessageInput) bool {
		return *input.QueueUrl == "test-queue" && *input.ReceiptHandle == "handle-1"
	}), mock.Anything).Return(&sqs.DeleteMessageOutput{}, nil)

	err := consumer.Ack(context.Background(), msg)
	assert.NoError(t, err)
}

func TestConsumer_Nack(t *testing.T) {
	mockClient := new(MockSQSClient)
	consumer := NewSqsConsumer(mockClient, Options{QueueURL: "test-queue"})
	msg := ports.Message{AckToken: "handle-1"}

	mockClient.On("ChangeMessageVisibility", mock.Anything, mock.MatchedBy(func(input *sqs.ChangeMessageVisibilityInput) bool {
		return *input.QueueUrl == "test-queue" && *input.ReceiptHandle == "handle-1" && input.VisibilityTimeout == 30
	}), mock.Anything).Return(&sqs.ChangeMessageVisibilityOutput{}, nil)

	err := consumer.Nack(context.Background(), msg, nackOptions{DelayBeforeRetrySeconds: 30})
	assert.NoError(t, err)
}
