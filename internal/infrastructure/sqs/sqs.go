package sqs

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	"github.com/aws/aws-sdk-go-v2/service/sqs/types"
	"github.com/google/uuid"
	"github.com/morikuni/failure"

	"github.com/marumaro-git/aws-cli-tool/internal/config"
	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/customerror"
)

type ReceiveMessage struct {
	Name    string `json:"name"`
	Age     int32  `json:"age"`
	City    string `json:"city"`
	Country string `json:"country"`
}

type SendMessage struct {
	Name string `json:"name"`
	Age  int32  `json:"age"`
	City string `json:"city"`
}

type SQSClient struct {
	client *sqs.Client
}

var (
	baseQueueURL = fmt.Sprintf("%s/000000000000", config.LocalStackEndpoint)
)

func NewSQSClient(ctx context.Context) *SQSClient {
	cfg := config.GetLocalStackConfig(ctx)
	client := sqs.NewFromConfig(cfg, func(o *sqs.Options) {
		o.BaseEndpoint = aws.String(config.LocalStackEndpoint)
	})
	return &SQSClient{
		client: client,
	}
}

func (s *SQSClient) ReceiveMessages(ctx context.Context) ([]model.Message, error) {

	receiveCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	input := &sqs.ReceiveMessageInput{
		QueueUrl:            aws.String(fmt.Sprintf("%s/receive-queue", baseQueueURL)),
		MaxNumberOfMessages: 10,
		WaitTimeSeconds:     1,
	}

	receiveMessages, err := s.client.ReceiveMessage(receiveCtx, input)
	if err != nil {
		return nil, failure.Wrap(err)
	}

	var messages []model.Message
	for _, msg := range receiveMessages.Messages {
		fmt.Println("Received raw message:", *msg.Body)

		var message model.Message
		if err := json.Unmarshal([]byte(*msg.Body), &message.User); err != nil {
			return nil, err
		}
		message.Metadata.ReceiptHandle = *msg.ReceiptHandle
		messages = append(messages, message)
	}

	return messages, nil
}

func (s *SQSClient) SendMessages(ctx context.Context, messages []model.Message) (*int, error) {

	if len(messages) > 10 {
		return nil, fmt.Errorf("SQS SendMessageBatch supports a maximum of 10 messages per batch")
	}

	sendCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	var entries []types.SendMessageBatchRequestEntry
	uuid := uuid.New().String()

	for i, message := range messages {
		jsonData, err := json.Marshal(message)
		if err != nil {
			return nil, failure.Wrap(err, failure.Message("failed to marshal message to JSON"))
		}

		entries = append(entries, types.SendMessageBatchRequestEntry{
			Id:          aws.String(fmt.Sprintf("%s-%d", uuid, i)),
			MessageBody: aws.String(string(jsonData)),
		})
	}

	input := &sqs.SendMessageBatchInput{
		QueueUrl: aws.String(fmt.Sprintf("%s/send-queue", baseQueueURL)),
		Entries:  entries,
	}

	sendMessages, err := s.client.SendMessageBatch(sendCtx, input)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to send messages"))
	}

	if len(sendMessages.Failed) > 0 {
		for _, failed := range sendMessages.Failed {
			fmt.Printf("Failed to send message %s: %v\n", *failed.Id, failed.Message)
		}
		return nil, failure.New(customerror.FailedSendMessage, failure.Message("some messages failed to send"))
	}

	return aws.Int(len(sendMessages.Successful)), nil
}

func (s *SQSClient) DeleteMessages(ctx context.Context, messages []model.Message) (*int, error) {

	var entries []types.DeleteMessageBatchRequestEntry
	uuid := uuid.New().String()

	for i, message := range messages {
		entries = append(entries, types.DeleteMessageBatchRequestEntry{
			Id:            aws.String(fmt.Sprintf("%s-%d", uuid, i)),
			ReceiptHandle: aws.String(message.Metadata.ReceiptHandle),
		})
	}

	input := &sqs.DeleteMessageBatchInput{
		QueueUrl: aws.String(fmt.Sprintf("%s/receive-queue", baseQueueURL)),
		Entries:  entries,
	}

	deleteMessages, err := s.client.DeleteMessageBatch(ctx, input)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to delete messages"))
	}

	if len(deleteMessages.Failed) > 0 {
		for _, failed := range deleteMessages.Failed {
			fmt.Printf("Failed to delete message %s: %v\n", *failed.Id, failed.Message)
		}
		return nil, failure.New(customerror.FailedSendMessage, failure.Message("some messages failed to delete"))
	}

	fmt.Printf("Successfully deleted %d messages\n", len(deleteMessages.Successful))

	return aws.Int(len(deleteMessages.Successful)), nil
}
