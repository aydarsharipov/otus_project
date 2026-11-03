// Package main - консольный клиент игры "Крестики-нолики".
package main

import (
	"bufio"
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"tictactoe/internal/api"
	"tictactoe/internal/client"
	"tictactoe/internal/service"
)

// pollInterval - как часто спрашиваем сервер, не сходил ли соперник.
const pollInterval = time.Second

// errInputClosed - пользователь закрыл ввод (Ctrl+D).
var errInputClosed = errors.New("ввод закончился")

func main() {
	defaultURL := os.Getenv("SERVER_URL")
	if defaultURL == "" {
		defaultURL = "http://localhost:8080"
	}
	serverURL := flag.String("server", defaultURL, "адрес сервера")
	flag.Parse()

	in := bufio.NewScanner(os.Stdin)
	c := client.New(*serverURL)
	ctx := context.Background()

	if err := startGame(ctx, c, in); err != nil {
		printError(err)
		os.Exit(1)
	}

	fmt.Printf("Вы играете за %s. ID игры: %s\n", c.Symbol, c.GameID)

	if err := playGame(ctx, c, in); err != nil {
		printError(err)
		os.Exit(1)
	}
}

// startGame спрашивает, создать новую игру или подключиться к существующей.
func startGame(ctx context.Context, c *client.Client, in *bufio.Scanner) error {
	for {
		choice, err := ask(in, "1 - создать игру, 2 - присоединиться к игре: ")
		if err != nil {
			return err
		}

		switch choice {
		case "1":
			answer, err := ask(in, "Размер поля (3 или 4): ")
			if err != nil {
				return err
			}
			size, err := strconv.Atoi(answer)
			if err != nil {
				fmt.Println("Нужно ввести число 3 или 4")
				continue
			}
			if err := c.CreateGame(ctx, size); err != nil {
				printError(err)
				continue
			}
			fmt.Println("Игра создана. Сообщите сопернику ID:", c.GameID)
			return nil
		case "2":
			id, err := ask(in, "Введите ID игры: ")
			if err != nil {
				return err
			}
			if err := c.JoinGame(ctx, id); err != nil {
				printError(err)
				continue
			}
			return nil
		default:
			fmt.Println("Введите 1 или 2")
		}
	}
}

// playGame - основной цикл: показываем поле, ждём свой ход, отправляем ход.
func playGame(ctx context.Context, c *client.Client, in *bufio.Scanner) error {
	waitingPrinted := false

	for {
		state, err := c.State(ctx)
		if err != nil {
			return err
		}

		if result := client.ResultText(state, c.Symbol); result != "" {
			fmt.Print("\n" + client.RenderBoard(state.Board))
			fmt.Println(result)
			return nil
		}

		if state.Status == service.StatusWaiting || state.Turn != c.Symbol {
			if !waitingPrinted {
				printWaiting(state)
				waitingPrinted = true
			}
			time.Sleep(pollInterval)
			continue
		}
		waitingPrinted = false

		fmt.Print("\n" + client.RenderBoard(state.Board))
		if err := makeMove(ctx, c, in, state.Size); err != nil {
			return err
		}
	}
}

// makeMove спрашивает ход у игрока, пока сервер его не примет.
func makeMove(ctx context.Context, c *client.Client, in *bufio.Scanner, size int) error {
	for {
		line, err := ask(in, fmt.Sprintf("Ваш ход (%s), строка и столбец: ", c.Symbol))
		if err != nil {
			return err
		}

		row, col, err := client.ParseMove(line, size)
		if err != nil {
			fmt.Println(err)
			continue
		}

		if _, err := c.Move(ctx, row, col); err != nil {
			fmt.Println("Ход не принят:", err)
			continue
		}
		return nil
	}
}

func printWaiting(state api.GameState) {
	if state.Status == service.StatusWaiting {
		fmt.Println("Ждём, пока соперник присоединится...")
		return
	}
	fmt.Print("\n" + client.RenderBoard(state.Board))
	fmt.Println("Ход соперника, ждём...")
}

func printError(err error) {
	fmt.Println("Ошибка:", err)
}

// ask печатает вопрос и читает одну строку ответа.
func ask(in *bufio.Scanner, question string) (string, error) {
	fmt.Print(question)
	if !in.Scan() {
		return "", errInputClosed
	}
	return strings.TrimSpace(in.Text()), nil
}
