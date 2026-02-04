package sqsconsumer

import (
	"context"
	"maps"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/guilherme-daniel-rs/event-processor/internal/logging"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Consumer struct {
	client      SQSClient
	queueURL    string
	maxMessages int32
	waitTimeSec int32
	maxRetries  int32
}

type Options struct {
	QueueURL    string
	MaxMessages int32
	WaitTimeSec int32
	MaxRetries  int32
}

type nackOptions struct {
	DelayBeforeRetrySeconds int32
}

func NewSqsConsumer(client SQSClient, opts Options) *Consumer {
	return &Consumer{
		client:      client,
		queueURL:    opts.QueueURL,
		maxMessages: opts.MaxMessages,
		waitTimeSec: opts.WaitTimeSec,
		maxRetries:  opts.MaxRetries,
	}
}

func (c *Consumer) Receive(ctx context.Context) ([]ports.Message, error) {
	out, err := c.client.ReceiveMessage(ctx, &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(c.queueURL),
		MaxNumberOfMessages: c.maxMessages,
		WaitTimeSeconds:     c.waitTimeSec,
		AttributeNames: []types.QueueAttributeName{
			types.QueueAttributeNameAll,
		},
	})
	if err != nil {
		return nil, err
	}

	msgs := make([]ports.Message, 0, len(out.Messages))
	for _, m := range out.Messages {
		msgs = append(msgs, toPortsMessage(m))
	}

	return msgs, nil
}

func (c *Consumer) Ack(ctx context.Context, msg ports.Message) error {
	_, err := c.client.DeleteMessage(ctx, &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(c.queueURL),
		ReceiptHandle: aws.String(msg.AckToken),
	})
	return err
}

func (c *Consumer) Nack(ctx context.Context, msg ports.Message, opts nackOptions) error {
	if opts.DelayBeforeRetrySeconds < 0 {
		return nil
	}
	_, err := c.client.ChangeMessageVisibility(ctx, &sqs.ChangeMessageVisibilityInput{
		QueueUrl:          aws.String(c.queueURL),
		ReceiptHandle:     aws.String(msg.AckToken),
		VisibilityTimeout: opts.DelayBeforeRetrySeconds,
	})
	return err
}

func (c *Consumer) Read(ctx context.Context, process func(ctx context.Context, msg ports.Message) error) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		messages, err := c.Receive(ctx)
		if err != nil {
			continue
		}

		var wg sync.WaitGroup
		for _, msg := range messages {
			if ctx.Err() != nil {
				return ctx.Err()
			}

			wg.Add(1)
			go func(m ports.Message) {
				defer wg.Done()

				tCtx := logging.WithTrace(ctx, m.ID)
				logging.Append(tCtx, "Started processing message (attempt %d)", m.ReceiveCount)

				err := process(tCtx, m)

				logging.Flush(tCtx, err)

				if err == nil {
					_ = c.Ack(ctx, m)
					return
				}

				if ports.IsNonRetriable(err) {
					_ = c.Ack(ctx, m)
					return
				}

				if c.maxRetries > 0 && int32(m.ReceiveCount) >= c.maxRetries {
					_ = c.Nack(ctx, m, nackOptions{
						DelayBeforeRetrySeconds: 0,
					})
					return
				}

				delay := calculateBackoffDelay(int32(m.ReceiveCount))

				_ = c.Nack(ctx, m, nackOptions{
					DelayBeforeRetrySeconds: delay,
				})
			}(msg)
		}
		wg.Wait()

	}
}

func calculateBackoffDelay(receiveCount int32) int32 {
	const maxAttempts = int32(10)

	delay := int32(30)

	attempt := receiveCount - 1
	if attempt < 0 {
		attempt = 0
	}
	if attempt > maxAttempts {
		attempt = maxAttempts
	}

	for i := int32(0); i < attempt; i++ {
		delay *= 2
	}

	return delay
}

func toPortsMessage(m types.Message) ports.Message {
	attrs := map[string]string{}
	maps.Copy(attrs, m.Attributes)

	receiveCount := 0
	if s, ok := attrs["ApproximateReceiveCount"]; ok {
		if n, err := strconv.Atoi(s); err == nil {
			receiveCount = n
		}
	}

	id := ""
	if m.MessageId != nil {
		id = *m.MessageId
	}

	body := []byte{}
	if m.Body != nil {
		body = []byte(*m.Body)
	}

	ackToken := ""
	if m.ReceiptHandle != nil {
		ackToken = *m.ReceiptHandle
	}

	return ports.Message{
		ID:           id,
		Body:         body,
		Attributes:   attrs,
		AckToken:     ackToken,
		ReceiveCount: receiveCount,
	}
}
