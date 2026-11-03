// Package api описывает JSON-структуры, которыми обмениваются клиент и сервер.
package api

// TokenHeader - заголовок, в котором клиент передаёт свой токен игрока.
const TokenHeader = "X-Player-Token"

// CreateGameRequest - тело запроса POST /games.
type CreateGameRequest struct {
	Size int `json:"size"`
}

// JoinResponse - ответ на создание игры или присоединение к ней.
type JoinResponse struct {
	GameID      string `json:"gameId"`
	PlayerToken string `json:"playerToken"`
	Symbol      string `json:"symbol"`
}

// MoveRequest - тело запроса POST /games/{id}/moves. Координаты начинаются с 0.
type MoveRequest struct {
	Row int `json:"row"`
	Col int `json:"col"`
}

// GameState - текущее состояние игры.
type GameState struct {
	GameID string     `json:"gameId"`
	Size   int        `json:"size"`
	Board  [][]string `json:"board"`
	Turn   string     `json:"turn"`
	Status string     `json:"status"`
}

// ErrorResponse - тело ответа с ошибкой.
type ErrorResponse struct {
	Error string `json:"error"`
}
