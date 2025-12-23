package model

type Item struct {
	ID        string `dynamodbav:"ID"`
	Name      string `dynamodbav:"Name"`
	City      string `dynamodbav:"City"`
	CreatedAt string `dynamodbav:"CreatedAt"`
}
