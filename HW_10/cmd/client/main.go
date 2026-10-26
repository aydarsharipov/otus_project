package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

const serverURL = "http://localhost:8080/api"

// поля для вывода доски
type Game struct {
	Board struct {
		Cells []struct {
			Row   int `json:"row"`
			Col   int `json:"col"`
			Piece struct {
				Symbol string `json:"symbol"`
			} `json:"piece"`
		} `json:"cells"`
	} `json:"board"`
	Turn   string `json:"turn"`
	Status string `json:"status"`
	Winner string `json:"winner"`
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

func printBoard(g Game) {
	var grid [8][8]string
	for _, c := range g.Board.Cells {
		grid[c.Row][c.Col] = c.Piece.Symbol
	}

	fmt.Println("  A B C D E F G H")
	for r := 0; r < 8; r++ {
		fmt.Print(8-r, " ")
		for c := 0; c < 8; c++ {
			if grid[r][c] == "" {
				fmt.Print(". ")
			} else {
				fmt.Print(grid[r][c], " ")
			}
		}
		fmt.Println()
	}

	if g.Status == "active" {
		fmt.Println("ходят:", g.Turn)
	} else {
		fmt.Println("игра окончена, победили:", g.Winner)
	}
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
		var g Game
		err = json.Unmarshal(data, &g)
		if err != nil {
			fmt.Println("ошибка:", err)
			return
		}
		printBoard(g)

		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		command := strings.TrimSpace(scanner.Text())

		if command == "" {
			continue
		}
		if command == "выход" {
			return
		}
		if command == "сдался" {
			_, err = sendRequest("PUT", gamePath, map[string]string{"action": "surrender"})
			if err != nil {
				fmt.Println("ошибка:", err)
			}
			continue
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
