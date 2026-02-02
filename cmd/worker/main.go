package main

import (
	"context"
	"fmt"
	"log"
	"sync"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/guilherme-daniel-rs/event-processor/internal/adapters/sqsconsumer"
	"github.com/guilherme-daniel-rs/event-processor/internal/app"
	"github.com/guilherme-daniel-rs/event-processor/internal/config"
)

func init() {
	config.Load()
}

func main() {
	ctx := context.Background()

	cfg, err := awsconfig.LoadDefaultConfig(
		ctx,
		awsconfig.WithRegion(config.Get().AWS.Region),
		awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(config.Get().AWS.AccessKeyID, config.Get().AWS.SecretAccessKey, "")),
	)
	if err != nil {
		log.Fatalf("failed to load aws config: %v", err)
	}

	localstackEndpoint := config.Get().AWS.Endpoint

	sqsClient := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(localstackEndpoint)
	})

	sqsConsumer := sqsconsumer.NewSqsConsumer(sqsClient, sqsconsumer.Options{
		QueueURL:    config.Get().SQS.QueueURL,
		MaxMessages: config.Get().SQS.MaxMessages,
		WaitTimeSec: config.Get().SQS.WaitTimeSec,
	})

	processor := app.NewProcessor(sqsConsumer)

	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := processor.Run(ctx); err != nil {
			fmt.Println("Processor stopped with error:", err)
		}
	}()

	wg.Wait()
	fmt.Println("Worker service is running...")
}
