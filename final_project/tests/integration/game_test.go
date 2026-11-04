// Package integration проверяет клиент и сервер вместе через настоящее HTTP-соединение.
package integration

import (
	"context"
	"net/http/httptest"
	"strings"
	"testing"

	"tictactoe/internal/client"
	"tictactoe/internal/game"
	"tictactoe/internal/service"
	"tictactoe/internal/transport/httpapi"
)

// startServer поднимает сервер на случайном порту и останавливает его в конце теста.
func startServer(t *testing.T) string {
	t.Helper()
	srv := httptest.NewServer(httpapi.New(service.New()).Routes())
	t.Cleanup(srv.Close)
	return srv.URL
}

// startGame создаёт игру первым клиентом и подключает второго.
func startGame(t *testing.T, url string, size int) (playerX, playerO *client.Client) {
	t.Helper()
	ctx := context.Background()

	playerX = client.New(url)
	if err := playerX.CreateGame(ctx, size); err != nil {
		t.Fatalf("создание игры: %v", err)
	}

	playerO = client.New(url)
	if err := playerO.JoinGame(ctx, playerX.GameID); err != nil {
		t.Fatalf("присоединение: %v", err)
	}
	return playerX, playerO
}

func TestFullGameXWins(t *testing.T) {
	ctx := context.Background()
	url := startServer(t)
	playerX, playerO := startGame(t, url, 3)

	if playerX.Symbol != game.X || playerO.Symbol != game.O {
		t.Fatalf("символы игроков: %s и %s", playerX.Symbol, playerO.Symbol)
	}

	moves := []struct {
		player   *client.Client
		row, col int
	}{
		{playerX, 0, 0}, {playerO, 1, 0}, {playerX, 1, 1}, {playerO, 2, 0}, {playerX, 2, 2},
	}
	for _, m := range moves {
		if _, err := m.player.Move(ctx, m.row, m.col); err != nil {
			t.Fatalf("ход %s в (%d,%d): %v", m.player.Symbol, m.row, m.col, err)
		}
	}

	// Оба игрока должны увидеть одинаковый итог.
	for _, p := range []*client.Client{playerX, playerO} {
		state, err := p.State(ctx)
		if err != nil {
			t.Fatal(err)
		}
		if state.Status != string(game.StatusXWon) {
			t.Errorf("игрок %s видит статус %s", p.Symbol, state.Status)
		}
	}

	stateO, err := playerO.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if got := client.ResultText(stateO, playerO.Symbol); got != "Вы проиграли. Победил X." {
		t.Errorf("текст для O: %q", got)
	}

	if _, err := playerO.Move(ctx, 0, 2); err == nil {
		t.Errorf("после конца игры ход должен быть запрещён")
	}
}

func TestFullGameDraw4x4(t *testing.T) {
	ctx := context.Background()
	url := startServer(t)
	playerX, playerO := startGame(t, url, 4)

	// Итоговое поле:
	// X X O O
	// O O X X
	// X X O O
	// O O X X
	cells := [][2]int{
		{0, 0},
		{0, 2},
		{0, 1},
		{0, 3},
		{1, 2},
		{1, 0},
		{1, 3},
		{1, 1},
		{2, 0},
		{2, 2},
		{2, 1},
		{2, 3},
		{3, 2},
		{3, 0},
		{3, 3},
		{3, 1},
	}
	for i, cell := range cells {
		p := playerX
		if i%2 == 1 {
			p = playerO
		}
		if _, err := p.Move(ctx, cell[0], cell[1]); err != nil {
			t.Fatalf("ход %d: %v", i, err)
		}
	}

	state, err := playerX.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != string(game.StatusDraw) {
		t.Errorf("статус %s, ожидали ничью", state.Status)
	}
	if client.ResultText(state, playerX.Symbol) != "Ничья!" {
		t.Errorf("неверный текст итога")
	}
}

func TestWaitingForSecondPlayer(t *testing.T) {
	ctx := context.Background()
	url := startServer(t)

	playerX := client.New(url)
	if err := playerX.CreateGame(ctx, 3); err != nil {
		t.Fatal(err)
	}

	state, err := playerX.State(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != service.StatusWaiting {
		t.Errorf("статус %s, ожидали %s", state.Status, service.StatusWaiting)
	}

	if _, err := playerX.Move(ctx, 0, 0); err == nil {
		t.Errorf("без второго игрока ходить нельзя")
	}
}

func TestErrorsThroughNetwork(t *testing.T) {
	ctx := context.Background()
	url := startServer(t)
	playerX, playerO := startGame(t, url, 3)

	tests := []struct {
		name     string
		do       func() error
		contains string
	}{
		{
			name:     "плохой размер",
			do:       func() error { return client.New(url).CreateGame(ctx, 5) },
			contains: game.ErrBadSize.Error(),
		},
		{
			name:     "нет такой игры",
			do:       func() error { return client.New(url).JoinGame(ctx, "unknown") },
			contains: service.ErrGameNotFound.Error(),
		},
		{
			name:     "третий игрок",
			do:       func() error { return client.New(url).JoinGame(ctx, playerX.GameID) },
			contains: service.ErrGameFull.Error(),
		},
		{
			name: "O ходит первым",
			do: func() error {
				_, err := playerO.Move(ctx, 0, 0)
				return err
			},
			contains: game.ErrNotYourTurn.Error(),
		},
		{
			name: "ход за поле",
			do: func() error {
				_, err := playerX.Move(ctx, 3, 3)
				return err
			},
			contains: game.ErrOutOfBoard.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.do()
			if err == nil || !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("ожидали ошибку с текстом %q, получили %v", tt.contains, err)
			}
		})
	}
}

func TestBusyCellThroughNetwork(t *testing.T) {
	ctx := context.Background()
	url := startServer(t)
	playerX, playerO := startGame(t, url, 3)

	if _, err := playerX.Move(ctx, 1, 1); err != nil {
		t.Fatal(err)
	}
	_, err := playerO.Move(ctx, 1, 1)
	if err == nil || !strings.Contains(err.Error(), game.ErrCellBusy.Error()) {
		t.Errorf("ожидали ошибку о занятой клетке, получили %v", err)
	}
}

func TestServerUnavailable(t *testing.T) {
	srv := httptest.NewServer(httpapi.New(service.New()).Routes())
	url := srv.URL
	srv.Close()

	if err := client.New(url).CreateGame(context.Background(), 3); err == nil {
		t.Errorf("ожидали ошибку, сервер выключен")
	}
}
