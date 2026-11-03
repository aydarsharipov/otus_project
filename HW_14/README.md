docker compose up -d
go run ./cmd/server

Адреса баз можно поменять через переменные MONGO_URI и REDIS_ADDR

curl -X POST localhost:8080/api/player -d '{"name":"Аня","color":"white"}'
curl localhost:8080/api/players
curl localhost:8080/api/logs

docker compose exec redis redis-cli keys 'log:*'
docker compose exec redis redis-cli ttl <ключ>

go test ./...
