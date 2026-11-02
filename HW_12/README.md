# HW_12 - шахматы через gRPC

## генерация кода из proto

нужны protoc и плагины:

```bash
brew install protobuf
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

потом:

```bash
make generate     # создаст pkg/chess_api/chess.pb.go и chess_grpc.pb.go
go mod tidy
```

## запуск

```bash
go run ./cmd/server     # в одном терминале
go run ./cmd/client     # в другом, help - список команд
```

данные сохраняются в папку data/ (players.json, games.json, moves.json) и загружаются при рестарте сервера

## проверка

```
addplayer Аня white
addplayer Боб black
addgame 1 2
move 1 E2 E4
game 1
renameplayer 1 Анна
surrender 1
moves
```

потом перезапустить сервер и сделать `players` / `games` - данные должны остаться

на сервере включён reflection, поэтому можно проверять ещё через Postman (New -> gRPC, localhost:5001)
