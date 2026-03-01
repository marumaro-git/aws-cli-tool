package docdb

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/marumaro-git/aws-cli-tool/internal/config"
	"github.com/marumaro-git/aws-cli-tool/internal/domain/model"
	"github.com/morikuni/failure"
)

const (
	DatabaseName   = "eventstore"
	CollectionName = "events"
)

type DocDBClient struct {
	client *mongo.Client
}

func NewDocDBClient(ctx context.Context) (*DocDBClient, error) {
	clientOptions := options.Client().ApplyURI(config.MongoDBURI)

	connectCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(connectCtx, clientOptions)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to connect to DocumentDB"))
	}

	if err := client.Ping(connectCtx, nil); err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to ping DocumentDB"))
	}

	return &DocDBClient{client: client}, nil
}

func (d *DocDBClient) InsertEvent(ctx context.Context, event model.EventDocument) (string, error) {
	collection := d.client.Database(DatabaseName).Collection(CollectionName)

	event.ID = generateTimestampBasedID(event.EventType, "node_01")
	event.EventTime = time.Now()

	_, err := collection.InsertOne(ctx, event)
	if err != nil {
		return "", failure.Wrap(err, failure.Message("failed to insert event"))
	}

	return event.ID, nil
}

func (d *DocDBClient) GetEventsSorted(ctx context.Context) ([]model.EventDocument, error) {
	collection := d.client.Database(DatabaseName).Collection(CollectionName)

	cursor, err := collection.Find(
		ctx,
		bson.M{},
		options.Find().SetSort(bson.D{bson.E{Key: "event_time", Value: 1}}),
	)
	if err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to find events"))
	}
	defer cursor.Close(ctx)

	var events []model.EventDocument
	if err := cursor.All(ctx, &events); err != nil {
		return nil, failure.Wrap(err, failure.Message("failed to decode events"))
	}

	return events, nil
}

func (d *DocDBClient) Close(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

func generateTimestampBasedID(eventType string, nodeID string) string {
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("%d-%s-%s", timestamp, eventType, nodeID)
}
