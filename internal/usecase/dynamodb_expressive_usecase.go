package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/logger"
)

type DynamoDBExpressiveRepository interface {
	PutItem(ctx context.Context) (string, error)
	GetItemByID(ctx context.Context, id string) (*model.Item, error)
	BatchWriteItems(ctx context.Context) error
	GetList(ctx context.Context, city string) ([]model.Item, error)
}

type DynamoDBExpressiveUseCase struct {
	repo   DynamoDBExpressiveRepository
	logger logger.Logger
}

func NewDynamoDBExpressiveUseCase(repo DynamoDBExpressiveRepository, logger logger.Logger) *DynamoDBExpressiveUseCase {
	return &DynamoDBExpressiveUseCase{
		repo:   repo,
		logger: logger,
	}
}

func (u *DynamoDBExpressiveUseCase) CheckTTLProcess(ctx context.Context) {

	id, err := u.repo.PutItem(ctx)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprint("Item put with ID:", id))

	item, err := u.repo.GetItemByID(ctx, id)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprint("Got item:", item))

	u.logger.Info(ctx, "Waiting for 5 seconds to allow TTL to expire...")
	time.Sleep(5 * time.Second)

	u.logger.Info(ctx, "Trying to get item after TTL expiration...")
	item, err = u.repo.GetItemByID(ctx, id)
	if err != nil {
		u.logger.Info(ctx, fmt.Sprint("Expected error after TTL expiration:", err))
	}

	if item == nil {
		u.logger.Info(ctx, "Expected item to be nil after TTL expiration")
	} else {
		u.logger.Info(ctx, fmt.Sprint("Got item after TTL expiration:", item))
	}

}

func (u *DynamoDBExpressiveUseCase) BatchWriteItems(ctx context.Context) {
	u.logger.Info(ctx, "Batch writing items...")
	err := u.repo.BatchWriteItems(ctx)
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, "Batch write complete.")

	u.logger.Info(ctx, "Getting list of items...")
	items, err := u.repo.GetList(ctx, "Batch City expressive")
	if err != nil {
		panic(err)
	}
	u.logger.Info(ctx, fmt.Sprintf("Got %d items:", len(items)))
}
