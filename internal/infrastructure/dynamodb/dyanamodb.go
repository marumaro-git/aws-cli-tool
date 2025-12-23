package dynamodb

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/marumaro-git/aws-cli-tool/internal/config"
	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
)

type DynamoDBClient struct {
	client *dynamodb.Client
}

type Item struct {
	ID        string `dynamodbav:"ID"`
	Name      string `dynamodbav:"Name"`
	City      string `dynamodbav:"City"`
	CreatedAt string `dynamodbav:"CreatedAt"`
	ExpiresAt int64  `dynamodbav:"ExpiresAt"`
}

const TableName = "SampleTable"

func NewDynamoDBClient() *DynamoDBClient {
	cfg := config.GetLocalStackConfig()
	client := dynamodb.NewFromConfig(cfg, func(o *dynamodb.Options) {
		o.BaseEndpoint = aws.String(config.LocalStackEndpoint)
	})
	return &DynamoDBClient{
		client: client,
	}
}

func (d *DynamoDBClient) PutItem(ctx context.Context) (string, error) {
	jst, err := time.LoadLocation("Asia/Tokyo")
	if err != nil {
		return "", err
	}
	createdAt := time.Now().In(jst).Format(time.RFC3339)
	item := Item{
		ID:        uuid.New().String(),
		Name:      "Sample Name",
		City:      "Sample City",
		CreatedAt: createdAt,
		ExpiresAt: time.Now().Add(5 * time.Second).Unix(),
	}

	itemMap, err := attributevalue.MarshalMap(item)
	if err != nil {
		return "", err
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      itemMap,
	}

	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return "", err
	}

	return item.ID, nil

}

func (d *DynamoDBClient) GetItemByID(ctx context.Context, id string) (*model.Item, error) {

	input := &dynamodb.GetItemInput{
		TableName: aws.String(TableName),
		Key: map[string]types.AttributeValue{
			"ID": &types.AttributeValueMemberS{Value: id},
		},
	}

	result, err := d.client.GetItem(ctx, input)
	if err != nil {
		return nil, err
	}

	if result.Item == nil {
		return &model.Item{}, nil
	}

	var item Item
	err = attributevalue.UnmarshalMap(result.Item, &item)
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
