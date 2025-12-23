#!/bin/bash

echo "localstack setup started"

echo "sqs setup started"

# SQSキューの作成
awslocal sqs create-queue --queue-name receive-queue
awslocal sqs create-queue --queue-name send-queue

# 作成したキューの確認
echo "Listing SQS Queues:"
awslocal sqs list-queues

echo "sqs setup completed"

echo "dyanamodb setup started"
# DynamoDBテーブルの作成
awslocal dynamodb create-table --table-name SampleTable \
    --attribute-definitions \
        AttributeName=ID,AttributeType=S \
        AttributeName=City,AttributeType=S \
        AttributeName=CreatedAt,AttributeType=S \
    --key-schema AttributeName=ID,KeyType=HASH \
    --global-secondary-indexes \
        IndexName=City-CreatedAt-index,KeySchema='[{AttributeName=City,KeyType=HASH},{AttributeName=CreatedAt,KeyType=RANGE}]',Projection='{ProjectionType=ALL}' \
    --billing-mode PAY_PER_REQUEST

# TTLの設定
awslocal dynamodb update-time-to-live \
    --table-name SampleTable \
    --time-to-live-specification "Enabled=true,AttributeName=ExpiresAt"

# 作成したテーブルの確認
echo "Listing DynamoDB Tables"

awslocal dynamodb list-tables

# SampleTableにアイテムを追加
awslocal dynamodb put-item --table-name SampleTable --item '{
    "ID": {"S": "123"}, 
    "Name": {"S": "Test Item"}, 
    "City": {"S": "Tokyo"}, 
    "CreatedAt": {"S": "2024-12-23T10:00:00Z"}
}'

awslocal dynamodb put-item --table-name SampleTable --item '{
    "ID": {"S": "124"}, 
    "Name": {"S": "Test Item 2"}, 
    "City": {"S": "Osaka"}, 
    "CreatedAt": {"S": "2024-12-23T11:00:00Z"}
}'

echo "dynamodb setup completed"

echo "localstack setup completed"