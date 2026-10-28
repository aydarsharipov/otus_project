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

// токен после входа
var token string

// поля для вывода доски
type Game struct {
	Board struct {
		Cells [][]*struct {
			Symbol string `json:"symbol"`
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
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
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
	fmt.Println("  A B C D E F G H")
	for r := 0; r < len(g.Board.Cells); r++ {
		fmt.Print(8-r, " ")
		for c := 0; c < len(g.Board.Cells[r]); c++ {
			if g.Board.Cells[r][c] == nil {
				fmt.Print(". ")
			} else {
				fmt.Print(g.Board.Cells[r][c].Symbol, " ")
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

	scanner := bufio.NewScanner(os.Stdin)

	// вход
	fmt.Print("логин: ")
	scanner.Scan()
	loginText := strings.TrimSpace(scanner.Text())
	fmt.Print("пароль: ")
	scanner.Scan()
	passwordText := strings.TrimSpace(scanner.Text())

	data, err := sendRequest("POST", "/login", map[string]string{"login": loginText, "password": passwordText})
	if err != nil {
		fmt.Println("не получилось войти:", err)
		return
	}
	var loginResp struct {
		Token string `json:"token"`
	}
	err = json.Unmarshal(data, &loginResp)
	if err != nil {
		fmt.Println("ошибка:", err)
		return
	}
	token = loginResp.Token
	fmt.Println("вход выполнен")

	gamePath := fmt.Sprintf("/game/%d", *gameID)

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
