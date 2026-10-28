package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/joho/godotenv"

	_ "HW_11/docs" // swagger
	"HW_11/internal/repository"
	"HW_11/internal/server"
)

//	@title			Шахматы API
//	@version		1.0
//	@description	Сервер для игры в шахматы. Сначала POST /login, потом нажать Authorize и вставить "Bearer <токен>"

//	@host		localhost:8080
//	@BasePath	/api

//	@securityDefinitions.apikey	BearerAuth
//	@in							header
//	@name						Authorization
//	@description				Введите: Bearer <токен>

func main() {
	err := godotenv.Load()
	if err != nil {
		fmt.Println("не найден файл .env, скопируйте env.example в .env")
		return
	}
	if os.Getenv("LOGIN") == "" || os.Getenv("PASSWORD") == "" || os.Getenv("JWT_SECRET") == "" {
		fmt.Println("в .env должны быть LOGIN, PASSWORD и JWT_SECRET")
		return
	}

	err = repository.LoadAll()
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

	go func() {
		fmt.Println("сервер запущен на", srv.Addr)
		fmt.Println("swagger: http://localhost:8080/swagger/index.html")
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
