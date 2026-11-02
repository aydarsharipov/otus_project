package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"HW_13/internal/repository"
	"HW_13/internal/server"
	"HW_13/internal/service"
)

func main() {
	err := repository.LoadAll()
	if err != nil {
		fmt.Println("ошибка загрузки данных:", err)
		return
	}
	fmt.Println("загружено игроков:", len(repository.Players.Items))
	fmt.Println("загружено игр:", len(repository.Games.Items))
	fmt.Println("загружено ходов:", len(repository.Moves.Items))

	// создаём сервис
	svc := service.New(&repository.Players, &repository.Games, &repository.Moves)

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
