# HW_11 - шахматы + swagger + jwt

## запуск

```bash
cp env.example .env          # и поменять пароль и JWT_SECRET
go mod tidy
go install github.com/swaggo/swag/cmd/swag@v1.16.6
swag init -d cmd/server,internal -g main.go -o docs
go run ./cmd/server
```

swagger: http://localhost:8080/swagger/index.html

после изменения комментариев над хендлерами нужно заново запустить `swag init`

## как проверить авторизацию в swagger

1. POST /login с логином и паролем из .env - придёт token
2. нажать Authorize, вписать `Bearer <token>`
3. теперь работают POST / PUT / DELETE, без токена будет 401
4. GET запросы работают без токена

## клиент

```bash
go run ./cmd/client -game 1
```

спросит логин и пароль, потом команды: `E2 E4`, `сдался`, `выход`, Enter - обновить доску
