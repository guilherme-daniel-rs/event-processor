package sqsconsumer

import (
	"context"
	"fmt"
	"maps"
	"strconv"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/guilherme-daniel-rs/event-processor/internal/ports"
)

type Consumer struct {
	client      *sqs.Client
	queueURL    string
	maxMessages int32
	waitTimeSec int32
}

type Options struct {
	QueueURL    string
	MaxMessages int32
	WaitTimeSec int32
}

type nackOptions struct {
	DelayBeforeRetrySeconds int32
}

func NewSqsConsumer(client *sqs.Client, opts Options) *Consumer {
	return &Consumer{
		client:      client,
		queueURL:    opts.QueueURL,
		maxMessages: opts.MaxMessages,
		waitTimeSec: opts.WaitTimeSec,
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
	if opts.DelayBeforeRetrySeconds <= 0 {
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

				if err := process(ctx, m); err != nil {
					fmt.Println("Error processing message ID ", m.ID, ": ", err.Error())
					_ = c.Nack(ctx, m, nackOptions{
						DelayBeforeRetrySeconds: 30,
					})
				} else {
					_ = c.Ack(ctx, m)
				}
			}(msg)
		}
		wg.Wait()

	}
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
