package main

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"HW_10/internal/repository"
	"HW_10/internal/server"
)

// сколько ждём завершения запросов при остановке сервера
const shutdownTimeout = 5 * time.Second

func main() {
	err := repository.LoadAll()
	if err != nil {
		fmt.Println("ошибка загрузки данных:", err)
		return
	}
	fmt.Println("загружено игроков:", len(repository.Players.Items))
	fmt.Println("загружено игр:", len(repository.Games.Items))
	fmt.Println("загружено ходов:", len(repository.Moves.Items))

	srv := &http.Server{
		Addr:    "localhost:8080",
		Handler: server.NewRouter(),
	}

	// ошибку сервера передаём в main через канал, а не через os.Exit
	serverErr := make(chan error, 1)
	go func() {
		fmt.Println("сервер запущен на", srv.Addr)
		err := srv.ListenAndServe()
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	// ждём Ctrl+C или ошибку сервера
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// если сервер упал - выходим, если пришёл сигнал - останавливаем сервер
	select {
	case err := <-serverErr:
		if err != nil {
			fmt.Println("ошибка сервера:", err)
		}
		return
	case <-ctx.Done():
	}

	fmt.Println("останавливаем сервер")
	// останавливаем сервер с таймаутом, чтобы не зависнуть
	shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancel()
	err = srv.Shutdown(shutdownCtx)
	if err != nil {
		fmt.Println("ошибка остановки сервера:", err)
	}
}
