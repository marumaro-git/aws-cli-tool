#!/bin/bash

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
