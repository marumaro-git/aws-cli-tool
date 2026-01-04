package dynamodb

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/google/uuid"
	"github.com/marumaro-git/aws-cli-tool/internal/config"
	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/marumaro-git/aws-cli-tool/internal/pkg/customerror"
	"github.com/morikuni/failure"
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

func NewDynamoDBClient(ctx context.Context) *DynamoDBClient {
	cfg := config.GetLocalStackConfig(ctx)
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
		return "", failure.Wrap(err)
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
		return "", failure.Wrap(err)
	}

	input := &dynamodb.PutItemInput{
		TableName: aws.String(TableName),
		Item:      itemMap,
		// NOTE: すでに存在するIDの場合は上書きせずエラーになる
		ConditionExpression: aws.String("attribute_not_exists(ID)"),
	}

	_, err = d.client.PutItem(ctx, input)
	if err != nil {
		return "", failure.Wrap(err)
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
		return nil, failure.Wrap(err)
	}

	if result.Item == nil {
		return nil, failure.New(customerror.ItemNotFound, failure.Message("item not found"))
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

func (d *DynamoDBClient) BatchWriteItems(ctx context.Context) error {

	items := []Item{
		{
			ID:        uuid.New().String(),
			Name:      "Sample Name 1",
			City:      "Batch City",
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
		},
		{
			ID:        uuid.New().String(),
			Name:      "Sample Name 2",
			City:      "Batch City",
			CreatedAt: time.Now().Format(time.RFC3339),
			ExpiresAt: time.Now().Add(60 * time.Minute).Unix(),
		},
	}

	var writeRequests []types.WriteRequest

	for _, item := range items {
		itemMap, err := attributevalue.MarshalMap(item)
		if err != nil {
			return failure.Wrap(err)
		}

		writeRequests = append(writeRequests, types.WriteRequest{
			PutRequest: &types.PutRequest{
				Item: itemMap,
			},
		})
	}

	input := &dynamodb.BatchWriteItemInput{
		RequestItems: map[string][]types.WriteRequest{
			TableName: writeRequests,
		},
	}

	output, err := d.client.BatchWriteItem(ctx, input)
	if err != nil {
		return failure.Wrap(err)
	}
	if len(output.UnprocessedItems) > 0 {
		return failure.Wrap(err, failure.Message(fmt.Sprintf("some items were not processed: %d", len(output.UnprocessedItems))))
	}
	return nil
}

func (d *DynamoDBClient) GetList(ctx context.Context, city string) ([]model.Item, error) {
	paginator := dynamodb.NewQueryPaginator(d.client, &dynamodb.QueryInput{
		TableName: aws.String(TableName),
		// 検索条件
		KeyConditionExpression: aws.String("City = :city"),
		ExpressionAttributeValues: map[string]types.AttributeValue{
			":city": &types.AttributeValueMemberS{Value: city},
		},
		IndexName: aws.String("City-CreatedAt-index"),
		Limit:     aws.Int32(4),
		// 最新順に取得
		ScanIndexForward: aws.Bool(false),
	})

	var items []model.Item

	for paginator.HasMorePages() {
		fmt.Println("Fetching next page...")
		output, err := paginator.NextPage(ctx)
		if err != nil {
			return nil, failure.Wrap(err)
		}

		for _, item := range output.Items {
			var modelItem model.Item
			err := attributevalue.UnmarshalMap(item, &modelItem)
			if err != nil {
				return nil, failure.Wrap(err)
			}
			items = append(items, modelItem)
		}
	}

	// DynamoDBからアイテムを取得するロジックを実装
	return items, nil
}
