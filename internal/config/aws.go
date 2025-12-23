package config

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
)

const (
	LocalStackEndpoint = "http://localhost.localstack.cloud:4566"
	
	DefaultRegion = "ap-northeast-1"
	
	DummyAccessKey    = "dummy_access_key"
	DummySecretKey    = "dummy_secret_key"
	DummySessionToken = ""
)

func GetLocalStackConfig() aws.Config {
	return aws.Config{
		Region:      DefaultRegion,
		Credentials: credentials.NewStaticCredentialsProvider(DummyAccessKey, DummySecretKey, DummySessionToken),
	}
}
