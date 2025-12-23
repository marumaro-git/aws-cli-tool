package usecase

import (
	"context"
	"fmt"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
)

type SQSRepository interface {
	ReceiveMessages(ctx context.Context) ([]model.Message, error)
	SendMessages(ctx context.Context, messages []model.Message) (*int, error)
	DeleteMessages(ctx context.Context, messages []model.Message) (*int, error)
}

type MessageUseCase struct {
	repository SQSRepository
}

func NewMessageUseCase(repository SQSRepository) *MessageUseCase {
	return &MessageUseCase{
		repository: repository,
	}
}

func (u *MessageUseCase) TransferMessages(ctx context.Context) {
	fmt.Println("Transferring messages...")

	fmt.Println("Receiving messages...")
	messages, err := u.repository.ReceiveMessages(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Received %d messages.\n", len(messages))

	if len(messages) == 0 {
		fmt.Println("No messages to transfer.")
		return
	}

	fmt.Println("Sending messages...")
	sendCount, err := u.repository.SendMessages(ctx, messages)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Messages sent: %d\n", *sendCount)

	deleteCount, err := u.repository.DeleteMessages(ctx, messages)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Messages deleted: %d\n", *deleteCount)

	fmt.Println("Message transfer complete.")
}
