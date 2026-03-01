package config

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	LocalStackEndpoint = "http://localhost.localstack.cloud:4566"

	DefaultRegion = "ap-northeast-1"

	DummyAccessKey    = "dummy_access_key"
	DummySecretKey    = "dummy_secret_key"
	DummySessionToken = ""

	MongoDBURI = "mongodb://localhost:27017/?replicaSet=rs0&directConnection=true"
)

func GetLocalStackConfig(ctx context.Context) aws.Config {
	cfg, _ := config.LoadDefaultConfig(ctx, config.WithRegion(DefaultRegion),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(DummyAccessKey, DummySecretKey, DummySessionToken)),
	)
	return cfg
}
