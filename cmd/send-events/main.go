package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/brianvoe/gofakeit/v6"
	"github.com/guilherme-daniel-rs/event-processor/internal/config"
)

func init() {
	config.Load()
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

func main() {
	count := flag.Int("count", 1, "Number of messages to send")
	eventType := flag.String("type", "user.created", "Event type")
	schemaVersion := flag.String("version", "v1", "Schema version")
	flag.Parse()

	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(config.Get().AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			config.Get().AWS.AccessKeyID,
			config.Get().AWS.SecretAccessKey,
			"",
		)),
	)
	if err != nil {
		log.Fatalf("failed to load aws config: %v", err)
	}

	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(config.Get().AWS.Endpoint)
	})

	queueURL := config.Get().SQS.QueueURL

	fmt.Printf("Sending %d message to queue: %s, eventyType: %s, version: %s\n", *count, queueURL, *eventType, *schemaVersion)

	for i := 0; i < *count; i++ {
		message := createMessage(*eventType, *schemaVersion)

		messageBody, _ := json.Marshal(message)

		_, err = sqsClient.SendMessage(ctx, &sqs.SendMessageInput{
			QueueUrl:    aws.String(queueURL),
			MessageBody: aws.String(string(messageBody)),
		})
		if err != nil {
			log.Printf("Failed to send message %d: %v", i, err)
			continue
		}

		fmt.Printf("Message %d sent successfully - Event ID: %s\n", i+1, message.EventID)
	}

	fmt.Printf("\nSuccessfully sent %d messages!\n", *count)
}

func createMessage(eventType, schemaVersion string) MessageHeader {
	eventID := fmt.Sprintf("evt-%d-%d", time.Now().Unix(), gofakeit.Number(1, 9999))
	occurredAt := time.Now().Format(time.RFC3339)

	tenantID := fmt.Sprintf("tenant-%d", gofakeit.Number(1, 10))
	clientID := fmt.Sprintf("client-%d", gofakeit.Number(1, 100))

	var fakeBody any
	fakeBody = getFakeData(eventType)

	bodyJSON, _ := json.Marshal(fakeBody)

	return MessageHeader{
		EventID:       eventID,
		EventType:     eventType,
		TenantID:      tenantID,
		ClientID:      clientID,
		SchemaVersion: schemaVersion,
		OccurredAt:    occurredAt,
		Body:          bodyJSON,
	}
}

func getFakeData(eventType string) map[string]any {
	switch eventType {
	case "payment.processed":
		return map[string]any{
			"payment_id":     fmt.Sprintf("pay-%d", gofakeit.Number(1000, 9999)),
			"order_id":       fmt.Sprintf("order-%d", gofakeit.Number(1000, 9999)),
			"amount":         99.99 + float64(gofakeit.Number(0, 9999)),
			"payment_method": "credit_card",
			"status":         "success",
		}
	default:
		return map[string]any{
			"message": fmt.Sprintf("Event %d", gofakeit.Number(1000, 9999)),
		}
	}
}
