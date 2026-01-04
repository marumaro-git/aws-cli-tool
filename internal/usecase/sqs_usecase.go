package usecase

import (
	"context"
	"fmt"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
)

type SQSRepository interface {
	ReceiveMessages(ctx context.Context) ([]model.Message, error)
	SendMessages(ctx context.Context, messages []model.Message) (*int, error)
	DeleteMessages(ctx context.Context, messages []model.Message) (*int, error)
}

type MessageUseCase struct {
	repository SQSRepository
	logger     logger.Logger
}

func NewMessageUseCase(repository SQSRepository, logger logger.Logger) *MessageUseCase {
	return &MessageUseCase{
		repository: repository,
		logger:     logger,
	}
}

func (u *MessageUseCase) TransferMessages(ctx context.Context) {
	messages, err := u.repository.ReceiveMessages(ctx)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprintf("Received %d messages.", len(messages)))

	if len(messages) == 0 {
		u.logger.Info(ctx, "No messages to transfer.")
		return
	}

	u.logger.Info(ctx, "Sending messages...")
	sendCount, err := u.repository.SendMessages(ctx, messages)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprintf("Messages sent: %d", *sendCount))

	deleteCount, err := u.repository.DeleteMessages(ctx, messages)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprintf("Messages deleted: %d", *deleteCount))

}
