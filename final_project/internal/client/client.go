// Package client - клиент для API игры и вспомогательные функции для консоли.
package client

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"tictactoe/internal/api"
)

// Client ходит в API сервера. После CreateGame или JoinGame запоминает игру и токен.
type Client struct {
	baseURL string
	http    *http.Client

	GameID string
	Symbol string
	token  string
}

// New создаёт клиента для сервера по адресу baseURL, например http://localhost:8080.
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		http:    &http.Client{Timeout: 5 * time.Second},
	}
}

// CreateGame создаёт новую игру с полем size x size.
func (c *Client) CreateGame(ctx context.Context, size int) error {
	var resp api.JoinResponse
	err := c.do(ctx, http.MethodPost, "/games", api.CreateGameRequest{Size: size}, &resp)
	if err != nil {
		return err
	}
	c.remember(resp)
	return nil
}

// JoinGame присоединяется к уже созданной игре.
func (c *Client) JoinGame(ctx context.Context, gameID string) error {
	var resp api.JoinResponse
	if err := c.do(ctx, http.MethodPost, gamePath(gameID)+"/join", nil, &resp); err != nil {
		return err
	}
	c.remember(resp)
	return nil
}

// State получает текущее состояние игры.
func (c *Client) State(ctx context.Context) (api.GameState, error) {
	var state api.GameState
	err := c.do(ctx, http.MethodGet, gamePath(c.GameID), nil, &state)
	return state, err
}

// Move делает ход. Координаты начинаются с 0.
func (c *Client) Move(ctx context.Context, row, col int) (api.GameState, error) {
	var state api.GameState
	err := c.do(ctx, http.MethodPost, gamePath(c.GameID)+"/moves", api.MoveRequest{Row: row, Col: col}, &state)
	return state, err
}

// gamePath возвращает путь к игре. ID экранируем, чтобы нельзя было подставить чужой путь.
func gamePath(gameID string) string {
	return "/games/" + url.PathEscape(gameID)
}

func (c *Client) remember(resp api.JoinResponse) {
	c.GameID = resp.GameID
	c.Symbol = resp.Symbol
	c.token = resp.PlayerToken
}

func (c *Client) do(ctx context.Context, method, path string, body, result interface{}) error {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return err
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	if c.token != "" {
		req.Header.Set(api.TokenHeader, c.token)
	}

	resp, err := c.http.Do(req)
	if err != nil {
		return fmt.Errorf("сервер недоступен: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		var errResp api.ErrorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil || errResp.Error == "" {
			return fmt.Errorf("сервер вернул код %d", resp.StatusCode)
		}
		return errors.New(errResp.Error)
	}

	return json.NewDecoder(resp.Body).Decode(result)
}
