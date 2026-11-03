export POSTGRES_PASSWORD=1234
docker compose up -d
go run ./cmd/server

sqlc generate

curl -X POST localhost:8080/api/player -d '{"name":"Аня","color":"white"}'
curl -X POST localhost:8080/api/player -d '{"name":"Боря","color":"black"}'
curl -X POST localhost:8080/api/game -d '{"white_id":1,"black_id":2}'
curl -X POST localhost:8080/api/move -d '{"game_id":1,"from":"E2","to":"E4"}'
curl localhost:8080/api/game/1

docker compose exec postgres psql -U chess -c 'select id, name, created_at from players'

go test ./...
