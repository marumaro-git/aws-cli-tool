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
	DatabaseName       = "eventstore"
	CollectionName     = "events"
	UserDatabaseName   = "userstore"
	UserCollectionName = "users"
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

func (d *DocDBClient) UpdateUserSingle(ctx context.Context, userID string, event model.EventDocument) (*model.UserState, bool, error) {
	collection := d.client.Database(UserDatabaseName).Collection(UserCollectionName)

	var currentUser model.UserState
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&currentUser)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, false, failure.Wrap(err, failure.Message("failed to find user"))
	}

	// イベント時刻が現在の状態より古い場合はスキップ（Last Write Wins）
	if !currentUser.LastUpdated.IsZero() && !event.EventTime.After(currentUser.LastUpdated) {
		return &currentUser, false, nil
	}

	update := bson.M{
		"$set": bson.M{
			"name":         event.Data["name"],
			"email":        event.Data["email"],
			"last_updated": event.EventTime,
		},
		"$inc": bson.M{"version": 1},
	}

	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, bson.M{"_id": userID}, update, opts)
	if err != nil {
		return nil, false, failure.Wrap(err, failure.Message("failed to update user"))
	}

	var updatedUser model.UserState
	if err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&updatedUser); err != nil {
		return nil, false, failure.Wrap(err, failure.Message("failed to fetch updated user"))
	}

	return &updatedUser, true, nil
}

func (d *DocDBClient) UpdateUserBatch(ctx context.Context, userID string, events []model.EventDocument) (*model.UserState, int, error) {
	collection := d.client.Database(UserDatabaseName).Collection(UserCollectionName)

	var currentUser model.UserState
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&currentUser)
	if err != nil && err != mongo.ErrNoDocuments {
		return nil, 0, failure.Wrap(err, failure.Message("failed to find user"))
	}

	session, err := d.client.StartSession()
	if err != nil {
		return nil, 0, failure.Wrap(err, failure.Message("failed to start session"))
	}
	defer session.EndSession(ctx)

	var appliedCount int
	_, err = session.WithTransaction(ctx, func(sessCtx mongo.SessionContext) (interface{}, error) {
		appliedCount = 0
		latestTime := currentUser.LastUpdated

		for _, event := range events {
			// 古いイベントはスキップ
			if !latestTime.IsZero() && !event.EventTime.After(latestTime) {
				continue
			}

			update := bson.M{
				"$set": bson.M{
					"name":         event.Data["name"],
					"email":        event.Data["email"],
					"last_updated": event.EventTime,
				},
				"$inc": bson.M{"version": 1},
			}

			opts := options.Update().SetUpsert(true)
			_, err := collection.UpdateOne(sessCtx, bson.M{"_id": userID}, update, opts)
			if err != nil {
				return nil, err
			}

			latestTime = event.EventTime
			appliedCount++
		}
		return nil, nil
	})
	if err != nil {
		return nil, 0, failure.Wrap(err, failure.Message("failed to execute batch update transaction"))
	}

	var updatedUser model.UserState
	if err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&updatedUser); err != nil {
		return nil, 0, failure.Wrap(err, failure.Message("failed to fetch updated user"))
	}

	return &updatedUser, appliedCount, nil
}

func (d *DocDBClient) GetUser(ctx context.Context, userID string) (*model.UserState, error) {
	collection := d.client.Database(UserDatabaseName).Collection(UserCollectionName)

	var user model.UserState
	err := collection.FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, failure.Wrap(err, failure.Message("failed to find user"))
	}
	return &user, nil
}

func (d *DocDBClient) DeleteUser(ctx context.Context, userID string) error {
	collection := d.client.Database(UserDatabaseName).Collection(UserCollectionName)
	_, err := collection.DeleteOne(ctx, bson.M{"_id": userID})
	if err != nil {
		return failure.Wrap(err, failure.Message("failed to delete user"))
	}
	return nil
}

func (d *DocDBClient) Close(ctx context.Context) error {
	return d.client.Disconnect(ctx)
}

func generateTimestampBasedID(eventType string, nodeID string) string {
	timestamp := time.Now().UnixMilli()
	return fmt.Sprintf("%d-%s-%s", timestamp, eventType, nodeID)
}
