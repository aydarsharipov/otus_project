package httpapi

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"tictactoe/internal/api"
	"tictactoe/internal/service"
)

const (
	gamesURL = "/games"
	moveBody = `{"row":0,"col":0}`
)

// doRequest отправляет запрос прямо в роутер, без настоящей сети.
func doRequest(t *testing.T, h http.Handler, method, url, body, token string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequestWithContext(context.Background(), method, url, strings.NewReader(body))
	if token != "" {
		req.Header.Set(api.TokenHeader, token)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec
}

func checkCode(t *testing.T, rec *httptest.ResponseRecorder, want int) {
	t.Helper()
	if rec.Code != want {
		t.Errorf("код ответа %d, ожидали %d, тело: %s", rec.Code, want, rec.Body.String())
	}
}

func decode(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("не удалось разобрать ответ: %v", err)
	}
}

func TestCreateGameHandler(t *testing.T) {
	h := New(service.New()).Routes()

	rec := doRequest(t, h, http.MethodPost, gamesURL, `{"size":3}`, "")
	checkCode(t, rec, http.StatusCreated)
	if ct := rec.Header().Get("Content-Type"); ct != "application/json" {
		t.Errorf("Content-Type %q", ct)
	}

	var resp api.JoinResponse
	decode(t, rec, &resp)
	if resp.GameID == "" || resp.PlayerToken == "" || resp.Symbol != "X" {
		t.Errorf("плохой ответ: %+v", resp)
	}
}

func TestHandlerErrors(t *testing.T) {
	svc := service.New()
	h := New(svc).Routes()

	waiting, err := svc.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}
	started, err := svc.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.JoinGame(started.GameID); err != nil {
		t.Fatal(err)
	}

	startedURL := gamesURL + "/" + started.GameID
	movesURL := startedURL + "/moves"
	waitingMovesURL := gamesURL + "/" + waiting.GameID + "/moves"
	tokenX := started.PlayerToken
	bigBody := `{"size":` + strings.Repeat("1", 2000) + `}`

	tests := []struct {
		name   string
		method string
		url    string
		body   string
		token  string
		want   int
	}{
		{"битый JSON", http.MethodPost, gamesURL, `{size:`, "", http.StatusBadRequest},
		{"лишнее поле", http.MethodPost, gamesURL, `{"size":3,"foo":1}`, "", http.StatusBadRequest},
		{"слишком большое тело", http.MethodPost, gamesURL, bigBody, "", http.StatusBadRequest},
		{"плохой размер", http.MethodPost, gamesURL, `{"size":5}`, "", http.StatusBadRequest},
		{"нет игры", http.MethodGet, gamesURL + "/nope", "", "", http.StatusNotFound},
		{"join нет игры", http.MethodPost, gamesURL + "/nope/join", "", "", http.StatusNotFound},
		{"join полная игра", http.MethodPost, startedURL + "/join", "", "", http.StatusConflict},
		{"ход без токена", http.MethodPost, movesURL, moveBody, "", http.StatusForbidden},
		{"ход чужим токеном", http.MethodPost, movesURL, moveBody, "bad", http.StatusForbidden},
		{"ход за поле", http.MethodPost, movesURL, `{"row":9,"col":0}`, tokenX, http.StatusBadRequest},
		{"ход битый JSON", http.MethodPost, movesURL, `[]`, tokenX, http.StatusBadRequest},
		{"ход без соперника", http.MethodPost, waitingMovesURL, moveBody, waiting.PlayerToken, http.StatusConflict},
		{"неверный метод", http.MethodDelete, startedURL, "", "", http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, h, tt.method, tt.url, tt.body, tt.token)
			checkCode(t, rec, tt.want)
		})
	}
}

func TestMoveHandler(t *testing.T) {
	svc := service.New()
	h := New(svc).Routes()

	created, err := svc.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}
	joined, err := svc.JoinGame(created.GameID)
	if err != nil {
		t.Fatal(err)
	}
	url := gamesURL + "/" + created.GameID + "/moves"
	center := `{"row":1,"col":1}`

	rec := doRequest(t, h, http.MethodPost, url, center, created.PlayerToken)
	checkCode(t, rec, http.StatusOK)
	var state api.GameState
	decode(t, rec, &state)
	if state.Board[1][1] != "X" || state.Turn != "O" {
		t.Errorf("неожиданное состояние: %+v", state)
	}

	// Повторный ход X - не его очередь.
	rec = doRequest(t, h, http.MethodPost, url, moveBody, created.PlayerToken)
	checkCode(t, rec, http.StatusConflict)

	// O пытается занять ту же клетку.
	rec = doRequest(t, h, http.MethodPost, url, center, joined.PlayerToken)
	checkCode(t, rec, http.StatusConflict)
	var errResp api.ErrorResponse
	decode(t, rec, &errResp)
	if errResp.Error == "" {
		t.Errorf("в ответе нет текста ошибки")
	}
}
