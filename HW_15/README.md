# HW_15 — хранение данных в PostgreSQL

- игроки, игры и ходы хранятся в PostgreSQL (таблицы `players`, `games`, `moves`)
- подключение к базе через `pgx`, миграции через `goose`
- миграции в `migrations/`: `001` создаёт таблицы, `002` добавляет поле `created_at`
- SQL-запросы лежат в `queries/queries.sql`, по ним `sqlc` генерирует код в `internal/db`
- ход сохраняется в транзакции: за один запрос `POST /api/move` добавляется строка в `moves` и обновляется доска в `games`

## Запуск

Пароль от базы в репозитории не хранится, его нужно задать в терминале:

```
export POSTGRES_PASSWORD=придумай_пароль
docker compose up -d
go run ./cmd/server
```

Миграции применяются сами при запуске сервера.

## Генерация кода

После изменения `queries/queries.sql` или миграций:

```
sqlc generate
```

## Проверка

```
curl -X POST localhost:8080/api/player -d '{"name":"Аня","color":"white"}'
curl -X POST localhost:8080/api/player -d '{"name":"Боря","color":"black"}'
curl -X POST localhost:8080/api/game -d '{"white_id":1,"black_id":2}'
curl -X POST localhost:8080/api/move -d '{"game_id":1,"from":"E2","to":"E4"}'
curl localhost:8080/api/game/1

docker compose exec postgres psql -U chess -c 'select id, name, created_at from players'
```

## Тесты

```
go test ./...
```

Тесты в `internal/repository` работают с настоящей базой, без `docker compose up` и `POSTGRES_PASSWORD` они пропускаются.
