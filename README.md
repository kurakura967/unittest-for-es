## unittest-for-es
Elasticsearchへのアクセスレイヤーに対する単体テストをmockせず、docker container上に立てた本物のElasticsearchへ接続し実行する。

## 単体テスト

```bash
export COMPOSE_FILE=docker-compose.test.yaml
docker compose up -d --build
go test -v -count=1 ./...
```
