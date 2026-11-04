# Крестики-нолики по сети

Игра для двух игроков в консоли. Есть сервер и клиент. Каждый игрок запускает свой клиент,
а сервер следит за ходами и определяет победителя.

Поле 3x3 или 4x4. Чтобы победить, нужно заполнить строку, столбец или диагональ целиком.

## Как запустить

cd final_project
docker compose up -d --build server

Первый игрок в одном терминале:
docker compose run --rm client

Выбрать `1`, ввести размер поля и отправить второму игроку ID игры.

Второй игрок в другом терминале:
docker compose run --rm client

Выбрать `2` и ввести ID.

Ход — это номер строки и столбца через пробел, например `2 3`.


## Без Docker

go run ./cmd/server
go run ./cmd/client

## Тесты

go test ./...
golangci-lint run ./...


## Что внутри

- `cmd/server` — сервер
- `cmd/client` — клиент
- `internal/game` — правила игры
- `internal/service` — партии и игроки
- `internal/transport/httpapi` — HTTP API
- `internal/client` — код клиента
- `tests/integration` — интеграционные тесты

