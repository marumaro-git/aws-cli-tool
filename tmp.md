# DocumentDB を利用した結果整合性の実現

## 概要

LocalStack上でDocumentDB（MongoDB互換）を利用し、イベント時刻をIDに含めることで結果整合性を実現する手法について調査・検証を行う。

## DocumentDB と LocalStack

### DocumentDBとは
- AWSが提供するMongoDB互換のマネージドドキュメントデータベース
- MongoDBのAPIバージョン3.6と4.0をサポート
- 高い可用性とスケーラビリティを提供

### LocalStackでのDocumentDB
- LocalStack Proで利用可能
- MongoDB互換エンジンとして動作
- 本番環境への移行が容易

## 設計方針

### IDにイベント時刻を持つ設計

#### ObjectIdベース設計
```
ObjectId: <timestamp><machineId><processId><counter>
例: 673d5a2b123456789abcdef0
```

#### カスタムIDベース設計
```
CustomID: <timestamp_ms>-<event_type>-<sequence>
例: 1709251200000-user_created-001
```

### 結果整合性のアプローチ

#### タイムスタンプベース順序保証
- イベント発生時刻をIDに含める
- タイムスタンプソートによる順序保証
- 分散環境での時刻同期が課題
- event_timeによるLast Write Wins戦略

## LocalStack環境構築

### docker-compose.yml設定
```yaml
version: '3.8'
services:
  localstack:
    image: localstack/localstack-pro:latest
    ports:
      - "4566:4566"
      - "27017:27017"  # DocumentDB/MongoDB ポート
    environment:
      - SERVICES=docdb
      - LOCALSTACK_API_KEY=${LOCALSTACK_API_KEY}
      - DEBUG=1
      - PERSISTENCE=1
      - MONGO_EXTERNAL_PORT=27017
    volumes:
      - "./tmp/localstack:/var/lib/localstack"
      - "/var/run/docker.sock:/var/run/docker.sock"
```

### 初期化スクリプト
```bash
#!/bin/bash
# init-docdb.sh

# DocumentDBクラスター作成
awslocal docdb create-db-cluster \
  --db-cluster-identifier test-cluster \
  --engine docdb \
  --master-username admin \
  --master-user-password password123

# DocumentDBインスタンス作成
awslocal docdb create-db-instance \
  --db-instance-identifier test-instance \
  --db-instance-class db.t3.medium \
  --engine docdb \
  --db-cluster-identifier test-cluster
```

## Go実装例

### 接続設定
```go
package main

import (
    "context"
    "fmt"
    "time"
    "go.mongodb.org/mongo-driver/mongo"
    "go.mongodb.org/mongo-driver/mongo/options"
    "go.mongodb.org/mongo-driver/bson/primitive"
)

type EventDocument struct {
    ID          string                 `json:"id" bson:"_id"`
    EventTime   time.Time             `json:"event_time" bson:"event_time"`
    EventType   string                `json:"event_type" bson:"event_type"`
    Data        map[string]interface{} `json:"data" bson:"data"`
    Version     int                   `json:"version" bson:"version"`
}

func connectDocumentDB() (*mongo.Client, error) {
    // LocalStack DocumentDB接続
    connectionURI := "mongodb://admin:password123@localhost:27017"
    clientOptions := options.Client().ApplyURI(connectionURI)
    
    client, err := mongo.Connect(context.TODO(), clientOptions)
    if err != nil {
        return nil, fmt.Errorf("failed to connect: %w", err)
    }
    
    return client, nil
}
```

### タイムスタンプベースID生成
```go
func generateTimestampBasedID(eventType string, nodeID string) string {
    timestamp := time.Now().UnixMilli()
    return fmt.Sprintf("%d-%s-%s", timestamp, eventType, nodeID)
}

func insertEvent(client *mongo.Client, event EventDocument) error {
    collection := client.Database("eventstore").Collection("events")
    
    // ID生成
    event.ID = generateTimestampBasedID(event.EventType, "node_01")
    event.EventTime = time.Now()
    
    _, err := collection.InsertOne(context.TODO(), event)
    return err
}
```

### 結果整合性チェック
```go
func checkEventualConsistency(client *mongo.Client, maxWaitTime time.Duration) error {
    collection := client.Database("eventstore").Collection("events")
    
    startTime := time.Now()
    for time.Since(startTime) < maxWaitTime {
        // 最新イベントを時系列順で取得
        cursor, err := collection.Find(
            context.TODO(),
            bson.M{},
            options.Find().SetSort(bson.D{{"event_time", 1}}),
        )
        if err != nil {
            return err
        }
        
        var events []EventDocument
        if err := cursor.All(context.TODO(), &events); err != nil {
            return err
        }
        
        // 時系列順序の検証
        if validateEventOrder(events) {
            return nil // 整合性OK
        }
        
        time.Sleep(100 * time.Millisecond)
    }
    
    return fmt.Errorf("eventual consistency not achieved within %v", maxWaitTime)
}

func validateEventOrder(events []EventDocument) bool {
    for i := 1; i < len(events); i++ {
        prev := events[i-1]
        curr := events[i]
        
        // タイムスタンプの順序チェック
        if curr.EventTime.Before(prev.EventTime) {
            return false
        }
    }
    return true
}
```

## ユースケース：event_timeベースの結果整合性パターン

### 1. 単一イベント更新（Single Event Update）

#### 概要
1件ずつのイベント更新で、event_timeベースの競合解決を行う。

#### 実装例
```go
type UserState struct {
    ID          string    `json:"id" bson:"_id"`
    Name        string    `json:"name" bson:"name"`
    Email       string    `json:"email" bson:"email"`
    LastUpdated time.Time `json:"last_updated" bson:"last_updated"`
    Version     int64     `json:"version" bson:"version"`
}

func updateUserSingle(client *mongo.Client, userID string, event EventDocument) error {
    collection := client.Database("userstore").Collection("users")
    
    // 現在の状態を取得
    var currentUser UserState
    err := collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&currentUser)
    if err != nil && err != mongo.ErrNoDocuments {
        return err
    }
    
    // イベント時刻チェック：古いイベントは破棄
    if !currentUser.LastUpdated.IsZero() && event.EventTime.Before(currentUser.LastUpdated) {
        return fmt.Errorf("event is older than current state, discarding: event_time=%v, current_time=%v", 
            event.EventTime, currentUser.LastUpdated)
    }
    
    // 更新処理（Last Write Wins）
    update := bson.M{
        "$set": bson.M{
            "name":         event.Data["name"],
            "email":        event.Data["email"],
            "last_updated": event.EventTime,
        },
        "$inc": bson.M{"version": 1},
    }
    
    opts := options.UpdateOptions{}
    opts.SetUpsert(true)
    
    _, err = collection.UpdateOne(context.TODO(), bson.M{"_id": userID}, update, &opts)
    return err
}
```

### 2. バッチイベント更新（Batch Event Update）

#### 概要
複数のイベントを一括処理し、event_time順で適用することで結果整合性を保証する。

#### 実装例
```go
func updateUserBatch(client *mongo.Client, userID string, events []EventDocument) error {
    // イベントをevent_time順でソート
    sort.Slice(events, func(i, j int) bool {
        return events[i].EventTime.Before(events[j].EventTime)
    })
    
    collection := client.Database("userstore").Collection("users")
    
    // 現在の状態を取得
    var currentUser UserState
    err := collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&currentUser)
    if err != nil && err != mongo.ErrNoDocuments {
        return err
    }
    
    // トランザクション開始
    session, err := client.StartSession()
    if err != nil {
        return err
    }
    defer session.EndSession(context.TODO())
    
    callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
        latestTime := currentUser.LastUpdated
        appliedEvents := 0
        
        for _, event := range events {
            // 古いイベントをスキップ
            if !latestTime.IsZero() && event.EventTime.Before(latestTime) {
                continue
            }
            
            // イベント適用
            update := bson.M{
                "$set": bson.M{
                    "name":         event.Data["name"],
                    "email":        event.Data["email"],
                    "last_updated": event.EventTime,
                },
                "$inc": bson.M{"version": 1},
            }
            
            opts := options.UpdateOptions{}
            opts.SetUpsert(true)
            
            _, err := collection.UpdateOne(sessCtx, bson.M{"_id": userID}, update, &opts)
            if err != nil {
                return nil, err
            }
            
            latestTime = event.EventTime
            appliedEvents++
        }
        
        return appliedEvents, nil
    }
    
    _, err = session.WithTransaction(context.TODO(), callback)
    return err
}
```

### 3. 古いイベント破棄ロジック（Stale Event Rejection）

#### タイムスタンプベース破棄
```go
func isEventStale(currentState UserState, incomingEvent EventDocument, gracePeriod time.Duration) bool {
    if currentState.LastUpdated.IsZero() {
        return false // 初回イベント
    }
    
    // グレースピリオドを考慮した破棄判定
    staleThreshold := currentState.LastUpdated.Add(-gracePeriod)
    return incomingEvent.EventTime.Before(staleThreshold)
}

func processEventWithStaleCheck(client *mongo.Client, userID string, event EventDocument) error {
    collection := client.Database("userstore").Collection("users")
    
    var currentUser UserState
    err := collection.FindOne(context.TODO(), bson.M{"_id": userID}).Decode(&currentUser)
    if err != nil && err != mongo.ErrNoDocuments {
        return err
    }
    
    // 5秒のグレースピリオドで古いイベントチェック
    if isEventStale(currentUser, event, 5*time.Second) {
        return &StaleEventError{
            EventTime:   event.EventTime,
            CurrentTime: currentUser.LastUpdated,
            Message:     "Event is too old to be applied",
        }
    }
    
    return updateUserSingle(client, userID, event)
}

type StaleEventError struct {
    EventTime   time.Time
    CurrentTime time.Time
    Message     string
}

func (e *StaleEventError) Error() string {
    return fmt.Sprintf("%s: event=%v, current=%v", e.Message, e.EventTime, e.CurrentTime)
}
```

### 4. 競合解決戦略

#### Last Write Wins (LWW)
```go
func applyLastWriteWins(currentState UserState, incomingEvent EventDocument) UserState {
    if incomingEvent.EventTime.After(currentState.LastUpdated) {
        // 新しいイベントが勝利
        return UserState{
            ID:          currentState.ID,
            Name:        incomingEvent.Data["name"].(string),
            Email:       incomingEvent.Data["email"].(string),
            LastUpdated: incomingEvent.EventTime,
            Version:     currentState.Version + 1,
        }
    }
    // 現在の状態を維持
    return currentState
}
```



## 検証シナリオ

### 4. 古いイベント破棄テスト
```go
func TestStaleEventRejection() {
    // 時系列が逆転したイベントの投入
    // 適切に破棄されることを確認
    // グレースピリオド内の許容範囲テスト
}
```

### 5. バッチ処理整合性テスト
```go
func TestBatchEventConsistency() {
    // 順序が混在したイベントバッチ
    // ソート後の正しい適用順序確認
    // 部分失敗時のロールバック確認
}
```

### 1. 基本的な時系列保証テスト
```go
func TestBasicTemporalOrder() {
    // 複数のイベントを同時期に投入
    // タイムスタンプ順でソートして順序確認
}
```

### 2. 分散環境での結果整合性テスト
```go
func TestDistributedConsistency() {
    // 複数ノードからの同時書き込み
    // event_timeによる順序保証確認
}
```

### 3. ネットワーク分断耐性テスト
```go
func TestNetworkPartitionTolerance() {
    // 一時的な接続断絶をシミュレート
    // 復旧後の整合性確認
}
```

## 実装上の注意点

### 時刻同期の課題
- システム間の時刻ずれ対策
- NTPサーバー同期の重要性
- 物理ラムバート時計の利用検討

### パフォーマンス考慮事項
- インデックス戦略（event_time, event_type）
- 大量データでの検索最適化
- 適切なコレクション分割

### 監視とメトリクス
```go
type ConsistencyMetrics struct {
    EventsProcessed   int64
    OutOfOrderEvents  int64
    MaxDelaySeconds   float64
    ConsistencyRatio  float64
}
```

## ベストプラクティス

1. **イベントスキーマの設計**
   - 不変なイベント構造
   - バージョニング戦略
   - 下位互換性の確保

2. **エラーハンドリング**
   - 重複イベントの検出と排除
   - 失敗イベントの再試行戦略
   - デッドレターキューの活用

3. **スケーラビリティ**
   - シャーディング戦略
   - 読み書き分離
   - キャッシュ戦略

## 次のステップ

1. PoC実装の作成
2. パフォーマンステストの実行
3. 本番環境への移行計画策定
4. 運用監視体制の構築