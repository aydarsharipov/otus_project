package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HW_14/internal/repository"
	"HW_14/internal/server"
	"HW_14/internal/service"
)

// сколько живёт каждая запись лога в redis
const logTTL = 24 * time.Hour

// берём настройку из переменной окружения, если её нет то значение по умолчанию
func getEnv(name string, def string) string {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	return val
}

func main() {
	mongoURI := getEnv("MONGO_URI", "mongodb://localhost:27017")
	redisAddr := getEnv("REDIS_ADDR", "localhost:6379")

	mongoClient, err := repository.ConnectMongo(mongoURI)
	if err != nil {
		fmt.Println("не удалось подключиться к mongodb:", err)
		return
	}
	defer mongoClient.Disconnect(context.Background())
	fmt.Println("подключились к mongodb")

	redisClient, err := repository.ConnectRedis(redisAddr)
	if err != nil {
		fmt.Println("не удалось подключиться к redis:", err)
		return
	}
	defer redisClient.Close()
	fmt.Println("подключились к redis")

	db := mongoClient.Database("chess")
	players, err := repository.NewPlayers(db)
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	games, err := repository.NewGames(db)
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	moves, err := repository.NewMoves(db)
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	logger := repository.NewRedisLogger(redisClient, logTTL)

	// создаём сервис
	svc := service.New(players, games, moves, logger)

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: server.NewRouter(svc),
	}

	go func() {
		fmt.Println("сервер запущен на", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && err != http.ErrServerClosed {
			fmt.Println("ошибка сервера:", err)
			os.Exit(1)
		}
	}()

	// ждём Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fmt.Println("останавливаем сервер")
	srv.Shutdown(context.Background())
}
