package usecase

import (
	"context"
	"fmt"
	"time"

	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
)

type DynamoDBExpressiveRepository interface {
	PutItem(ctx context.Context) (string, error)
	GetItemByID(ctx context.Context, id string) (*model.Item, error)
	BatchWriteItems(ctx context.Context) error
	GetList(ctx context.Context, city string) ([]model.Item, error)
}

type DynamoDBExpressiveUseCase struct {
	repo DynamoDBExpressiveRepository
}

func NewDynamoDBExpressiveUseCase(repo DynamoDBExpressiveRepository) *DynamoDBExpressiveUseCase {
	return &DynamoDBExpressiveUseCase{
		repo: repo,
	}
}

func (u *DynamoDBExpressiveUseCase) CheckTTLProcess(ctx context.Context) {
	fmt.Println("Checking TTL process...")
	fmt.Println("Putting item...")
	id, err := u.repo.PutItem(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Item put with ID:", id)

	fmt.Println("Getting item...")
	item, err := u.repo.GetItemByID(ctx, id)
	if err != nil {
		panic(err)
	}
	fmt.Println("Got item:", item)

	fmt.Println("Waiting for 5 seconds to allow TTL to expire...")
	time.Sleep(5 * time.Second)

	fmt.Println("Trying to get item after TTL expiration...")
	item, err = u.repo.GetItemByID(ctx, id)
	if err != nil {
		fmt.Println("Expected error after TTL expiration:", err)
	}

	if item == nil {
		fmt.Println("Expected item to be nil after TTL expiration")
	} else {
		fmt.Println("Got item after TTL expiration:", item)
	}

	fmt.Println("Check complete.")
}

func (u *DynamoDBExpressiveUseCase) BatchWriteItems(ctx context.Context) {
	fmt.Println("Batch writing items...")
	err := u.repo.BatchWriteItems(ctx)
	if err != nil {
		panic(err)
	}
	fmt.Println("Batch write complete.")

	fmt.Println("Getting list of items...")
	items, err := u.repo.GetList(ctx, "Batch City expressive")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Got %d items:\n", len(items))
}
