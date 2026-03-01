package model

import "time"

type EventDocument struct {
	ID        string                 `json:"id" bson:"_id"`
	EventTime time.Time              `json:"event_time" bson:"event_time"`
	EventType string                 `json:"event_type" bson:"event_type"`
	Data      map[string]interface{} `json:"data" bson:"data"`
	Version   int                    `json:"version" bson:"version"`
}

type UserState struct {
	ID          string    `json:"id" bson:"_id"`
	Name        string    `json:"name" bson:"name"`
	Email       string    `json:"email" bson:"email"`
	LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
	Version     int64     `json:"version" bson:"version"`
}
