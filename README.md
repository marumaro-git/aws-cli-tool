# 概要
このリポジトリは、AWS CLIツールを使用して、localstack上でSQSおよびDynamoDBの操作を行うためのサンプルコードと手順を提供します。

## sqs
sqsからメッセージを受信し、別のsqsに送信するアプリケーションです。

```bash
go run main.go sqs
```

### dynamodb
dynamodbのTTL（Time To Live）機能を確認するアプリケーションです。

```bash
go run main.go dynamodb
```


# セットアップ
このリポジトリをクローンし、Dockerコンテナを起動します。

```bash
docker compose -f docker/docker-compose.yaml up -d
```

# localstackの操作方法

## SQS
### メッセージの送受信
- メッセージの追加
```bash
awslocal sqs send-message \
  --queue-url http://localhost.localstack.cloud:4566/000000000000/receive-queue \
  --message-body '{
    "name": "test",
    "age": 30,
    "city": "Tokyo",
    "country": "Japan"
  }'
  ``` 

- メッセージの受信
```bash
awslocal sqs receive-message \
    --queue-url http://localhost.localstack.cloud:4566/000000000000/send-queue \
    --max-number-of-messages 10 \
    --wait-time-seconds 1 \
    --visibility-timeout 1
```

- 件数の確認
```bash
awslocal sqs get-queue-attributes \
  --queue-url http://localhost.localstack.cloud:4566/000000000000/receive-queue \
  --attribute-names ApproximateNumberOfMessages
```

## DynamoDB
### 全件取得
```bash
awslocal dynamodb scan --table-name SampleTable
```
