package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"syscall"

	"HW_15/internal/repository"
	"HW_15/internal/server"
	"HW_15/internal/service"
)

// берём настройку из переменной окружения, если её нет то значение по умолчанию
func getEnv(name string, def string) string {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	return val
}

func main() {
	// пароль не храним в коде, берём только из переменной окружения
	password := os.Getenv("POSTGRES_PASSWORD")
	if password == "" {
		fmt.Println("задайте переменную POSTGRES_PASSWORD")
		return
	}
	// собираем адрес базы, url.URL сам экранирует спецсимволы в пароле
	dbURL := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("chess", password),
		Host:     getEnv("POSTGRES_HOST", "localhost:5432"),
		Path:     "chess",
		RawQuery: "sslmode=disable",
	}

	pool, err := repository.Connect(dbURL.String())
	if err != nil {
		fmt.Println("не удалось подключиться к postgres:", err)
		return
	}
	defer pool.Close()
	fmt.Println("подключились к postgres")

	err = repository.Migrate(pool)
	if err != nil {
		fmt.Println("ошибка миграций:", err)
		return
	}

	players := repository.NewPlayers(pool)
	games := repository.NewGames(pool)
	moves := repository.NewMoves(pool)

	// создаём сервис
	svc := service.New(players, games, moves)

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
