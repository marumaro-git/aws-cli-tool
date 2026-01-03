package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/uuid"
	"github.com/guregu/dynamo/v2"
	"github.com/marumaro-git/aws-cli-tool/internal/config"
	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
)

type ExpressiveDynamoDBClient struct {
	client *dynamo.DB
}

type ExpressiveTableClient struct {
	table dynamo.Table
}

func NewExpressiveDynamoDBClient(ctx context.Context) *ExpressiveDynamoDBClient {
	cfg := config.GetLocalStackConfig(ctx)
	client := dynamo.New(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(config.LocalStackEndpoint)
	})

	return &ExpressiveDynamoDBClient{
		client: client,
	}
}

func (d *ExpressiveDynamoDBClient) NewExpressiveTableClient() *ExpressiveTableClient {
	table := d.client.Table(TableName)
	return &ExpressiveTableClient{
		table: table,
	}
}

func (t *ExpressiveTableClient) PutItem(ctx context.Context) (string, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}
	createdAt := time.Now().In(jst).Format(time.RFC3339)
	item := Item{
		ID:        uuid.New().String(),
		Name:      "Sample Name expressive",
		City:      "Sample City expressive",
		CreatedAt: createdAt,
		ExpiresAt: time.Now().Add(5 * time.Second).Unix(),
	}
	err = t.table.Put(item).If("attribute_not_exists(ID)").Run(ctx)
	if err != nil {
		return "", err
	}
	return item.ID, nil
}

func (t *ExpressiveTableClient) GetItemByID(ctx context.Context, id string) (*model.Item, error) {
	var item Item
	err := t.table.Get("ID", id).One(ctx, &item)
	if err != nil {
		return nil, err
	}

	modelItem := model.Item{
		ID:        item.ID,
		Name:      item.Name,
		City:      item.City,
		CreatedAt: item.CreatedAt,
	}
	return &modelItem, nil
}

func (t *ExpressiveTableClient) BatchWriteItems(ctx context.Context) error {
	items := []Item{
		{
			ID:        uuid.New().String(),
			Name:      "Sample Name expressive 1",
			City:      "Batch City expressive",
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
		},
		{
			ID:        uuid.New().String(),
			Name:      "Sample Name expressive 2",
			City:      "Batch City expressive",
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
		},
	}

	itemsInterface := make([]interface{}, len(items))
	for i, item := range items {
		itemsInterface[i] = item
	}
	
	cnt, err := t.table.Batch().Write().Put(itemsInterface...).Run(ctx)
	if err != nil {
		return err
	}

	if cnt != len(items) {
		return fmt.Errorf("not all items were written: expected %d, got %d", len(items), cnt)
	}

	return nil
}

func (t *ExpressiveTableClient) GetList(ctx context.Context, city string) ([]model.Item, error) {
	var items []Item

	err := t.table.Get("City", city).Index("City-CreatedAt-index").All(ctx, &items)
	if err != nil {
		return nil, err
	}

	var modelItems []model.Item

	for _, item := range items {
		modelItem := model.Item{
			ID:        item.ID,
			Name:      item.Name,
			City:      item.City,
			CreatedAt: item.CreatedAt,
		}
		modelItems = append(modelItems, modelItem)
	}

	return modelItems, nil
}
