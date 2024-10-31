# User-API

このプロジェクトは、Go言語とMySQLを使用して構築されたユーザー管理APIです。基本的なCRUD操作（Create, Read, Update, Delete）を提供します。

## 使用技術

- **Go バージョン**: 1.23
- **MySQL バージョン**: 8.0

## エンドポイントとサンプルリクエスト

### 1. Create User

- **HTTP メソッド**: `POST`
- **リクエストボディ**:
  ```bash
  curl -X POST http://localhost:8081/users \
     -H "Content-Type: application/json" \
     -d '{
          "account_id": "12345",
          "first_name": "geen",
          "last_name": "100",
          "age": 20
         }'
  ```
- **レスポンスボディ**:
  ```bash
  curl -X POST http://localhost:8081/users \
     -H "Content-Type: application/json" \
     -d '{
            "id": 1,
            "account_id": "12345",
            "first_name": "geen",
            "last_name": "100",
            "age": 20
         }'
  ```

### 2. Get All Users

- **HTTP メソッド**: `GET`
- **サンプルリクエスト**:
  ```bash
  curl -X GET http://localhost:8081/users

### 3. Get User by ID

- **HTTP メソッド**: `GET`
- **サンプルリクエスト**:
  ```bash
  curl -X GET http://localhost:8081/users/1

### 4. Update User

- **HTTP メソッド**: `PUT`
- **リクエストボディ**:
```bash
  curl -X PUT http://localhost:8081/users/1 \
     -H "Content-Type: application/json" \
     -d '{
          "age": 25
         }'
  ```
- **レスポンスボディ**:
```bash
  curl -X PUT http://localhost:8081/users/1 \
     -H "Content-Type: application/json" \
     -d '{
          "id": 1,
          "account_id": "12345",
          "first_name": "geen",
          "last_name": "100",
          "age": 25
         }'
  ```
### Delete User

- **HTTP メソッド**: `DELETE`
- **リクエストボディ**:
  ```bash
  curl -X DELETE http://localhost:8081/users/1
  ```



