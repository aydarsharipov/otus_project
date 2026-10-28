package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const serverURL = "http://localhost:8080/api"

// команды пользователя, статус игры и параметры доски
const (
	cmdQuit      = "выход"
	cmdSurrender = "сдался"
	statusActive = "active"
	boardSize    = 8
	boardHeader  = "  A B C D E F G H"
)

// фигура на клетке
type Piece struct {
	Symbol string `json:"symbol"`
}

// клетка доски с фигурой
type Cell struct {
	Row   int   `json:"row"`
	Col   int   `json:"col"`
	Piece Piece `json:"piece"`
}

// доска - список занятых клеток
type Board struct {
	Cells []Cell `json:"cells"`
}

// игра: доска, чей ход, статус и победитель
type Game struct {
	Board  Board  `json:"board"`
	Turn   string `json:"turn"`
	Status string `json:"status"`
	Winner string `json:"winner"`
}

// проверка, что клетка не выходит за пределы доски
func (c Cell) validate() error {
	if c.Row < 0 || c.Row >= boardSize || c.Col < 0 || c.Col >= boardSize {
		return fmt.Errorf("клетка за пределами доски: row=%d col=%d", c.Row, c.Col)
	}
	return nil
}

// отправка запроса на сервер
func sendRequest(method string, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequest(method, serverURL+path, reqBody)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode >= 300 {
		return nil, fmt.Errorf("%d %s", resp.StatusCode, strings.TrimSpace(string(respBody)))
	}
	return respBody, nil
}

// разбор ответа сервера с игрой
func parseGame(data []byte) (Game, error) {
	var g Game
	err := json.Unmarshal(data, &g)
	return g, err
}

func printBoard(g Game) error {
	var grid [boardSize][boardSize]string
	for _, c := range g.Board.Cells {
		// проверяем клетку, чтобы не было паники при выходе за пределы массива
		if err := c.validate(); err != nil {
			return err
		}
		grid[c.Row][c.Col] = c.Piece.Symbol
	}

	// заголовок с буквами столбцов
	fmt.Println(boardHeader)
	for r := 0; r < boardSize; r++ {
		fmt.Print(boardSize-r, " ")
		for c := 0; c < boardSize; c++ {
			if grid[r][c] == "" {
				fmt.Print(". ")
			} else {
				fmt.Print(grid[r][c], " ")
			}
		}
		fmt.Println()
	}

	if g.Status == statusActive {
		fmt.Println("ходят:", g.Turn)
	} else {
		fmt.Println("игра окончена, победили:", g.Winner)
	}
	return nil
}

// запуск: go run ./cmd/client -game 1
// команды: E2 E4 | Enter - обновить доску | сдался | выход
func main() {
	gameID := flag.Int("game", 0, "номер игры")
	flag.Parse()
	if *gameID <= 0 {
		fmt.Println("укажите номер игры: -game 1")
		return
	}

	gamePath := fmt.Sprintf("/game/%d", *gameID)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		data, err := sendRequest("GET", gamePath, nil)
		if err != nil {
			fmt.Println("ошибка:", err)
			return
		}
		g, err := parseGame(data)
		if err != nil {
			fmt.Println("ошибка:", err)
			return
		}
		if err = printBoard(g); err != nil {
			fmt.Println("ошибка:", err)
			return
		}
		// игра закончилась - больше не опрашиваем сервер
		if g.Status != statusActive {
			return
		}

		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		command := strings.TrimSpace(scanner.Text())

		if command == "" {
			continue
		}
		if command == cmdQuit {
			return
		}
		// сдача: при успехе выходим, при ошибке остаёмся в игре
		if command == cmdSurrender {
			err = surrender(gamePath)
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			return
		}

		parts := strings.Fields(command)
		if len(parts) != 2 {
			fmt.Println("ход нужно писать так: E2 E4")
			continue
		}
		moveData := map[string]any{"game_id": *gameID, "from": parts[0], "to": parts[1]}
		_, err = sendRequest("POST", "/move", moveData)
		if err != nil {
			fmt.Println("ошибка:", err)
		}
	}
}

// сдача: сервер возвращает обновлённую игру, по ней проверяем,
// что сдача применилась, и показываем итог
func surrender(gamePath string) error {
	data, err := sendRequest("PUT", gamePath, map[string]string{"action": "surrender"})
	if err != nil {
		return err
	}
	g, err := parseGame(data)
	if err != nil {
		return err
	}
	// если статус не поменялся, значит сдача не применилась
	if g.Status == statusActive {
		return errors.New("сервер не применил сдачу, игра всё ещё идёт")
	}
	fmt.Println("вы сдались")
	return printBoard(g)
}
