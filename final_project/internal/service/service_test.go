package service

import (
	"errors"
	"sync"
	"testing"

	"tictactoe/internal/game"
)

// newStartedGame создаёт игру 3x3 и сразу добавляет второго игрока.
func newStartedGame(t *testing.T, s *Service) (id, tokenX, tokenO string) {
	t.Helper()

	created, err := s.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}
	joined, err := s.JoinGame(created.GameID)
	if err != nil {
		t.Fatal(err)
	}
	return created.GameID, created.PlayerToken, joined.PlayerToken
}

func TestCreateGame(t *testing.T) {
	s := New()

	resp, err := s.CreateGame(4)
	if err != nil {
		t.Fatal(err)
	}
	if resp.GameID == "" || resp.PlayerToken == "" {
		t.Fatalf("пустой id или токен: %+v", resp)
	}
	if resp.Symbol != game.X {
		t.Errorf("создатель должен играть за X, а играет за %s", resp.Symbol)
	}

	state, err := s.State(resp.GameID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != StatusWaiting {
		t.Errorf("новая игра в статусе %s, ожидали %s", state.Status, StatusWaiting)
	}
	if state.Size != 4 {
		t.Errorf("размер %d, ожидали 4", state.Size)
	}
}

func TestCreateGameBadSize(t *testing.T) {
	s := New()
	if _, err := s.CreateGame(7); !errors.Is(err, game.ErrBadSize) {
		t.Errorf("ожидали ErrBadSize, получили %v", err)
	}
}

func TestCreateGameLimit(t *testing.T) {
	s := New()
	for i := 0; i < MaxGames; i++ {
		if _, err := s.CreateGame(3); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := s.CreateGame(3); !errors.Is(err, ErrTooManyGames) {
		t.Errorf("ожидали ErrTooManyGames, получили %v", err)
	}
}

func TestJoinGame(t *testing.T) {
	s := New()
	created, err := s.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}

	joined, err := s.JoinGame(created.GameID)
	if err != nil {
		t.Fatal(err)
	}
	if joined.Symbol != game.O {
		t.Errorf("второй игрок должен играть за O, а играет за %s", joined.Symbol)
	}
	if joined.PlayerToken == created.PlayerToken {
		t.Errorf("у игроков одинаковые токены")
	}

	state, err := s.State(created.GameID)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != string(game.StatusPlaying) {
		t.Errorf("после join статус %s, ожидали %s", state.Status, game.StatusPlaying)
	}

	if _, err := s.JoinGame(created.GameID); !errors.Is(err, ErrGameFull) {
		t.Errorf("третий игрок: ожидали ErrGameFull, получили %v", err)
	}
}

func TestGameNotFound(t *testing.T) {
	const id = "nope"
	s := New()
	if _, err := s.JoinGame(id); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("JoinGame: ожидали ErrGameNotFound, получили %v", err)
	}
	if _, err := s.State(id); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("State: ожидали ErrGameNotFound, получили %v", err)
	}
	if _, err := s.Move(id, "token", 0, 0); !errors.Is(err, ErrGameNotFound) {
		t.Errorf("Move: ожидали ErrGameNotFound, получили %v", err)
	}
}

func TestMoveBeforeSecondPlayer(t *testing.T) {
	s := New()
	created, err := s.CreateGame(3)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := s.Move(created.GameID, created.PlayerToken, 0, 0); !errors.Is(err, ErrWaiting) {
		t.Errorf("ожидали ErrWaiting, получили %v", err)
	}
}

func TestMoveBadToken(t *testing.T) {
	s := New()
	id, _, _ := newStartedGame(t, s)

	for _, token := range []string{"", "wrong"} {
		if _, err := s.Move(id, token, 0, 0); !errors.Is(err, ErrBadToken) {
			t.Errorf("токен %q: ожидали ErrBadToken, получили %v", token, err)
		}
	}
}

func TestMoveTokenFromOtherGame(t *testing.T) {
	s := New()
	id1, _, _ := newStartedGame(t, s)
	_, tokenX2, _ := newStartedGame(t, s)

	if _, err := s.Move(id1, tokenX2, 0, 0); !errors.Is(err, ErrBadToken) {
		t.Errorf("ожидали ErrBadToken, получили %v", err)
	}
}

func TestMoveFullGame(t *testing.T) {
	s := New()
	id, tokenX, tokenO := newStartedGame(t, s)

	if _, err := s.Move(id, tokenO, 0, 0); !errors.Is(err, game.ErrNotYourTurn) {
		t.Errorf("O ходит первым: ожидали ErrNotYourTurn, получили %v", err)
	}

	moves := []struct {
		token    string
		row, col int
	}{
		{tokenX, 0, 0}, {tokenO, 1, 0}, {tokenX, 0, 1}, {tokenO, 1, 1}, {tokenX, 0, 2},
	}
	for _, m := range moves {
		if _, err := s.Move(id, m.token, m.row, m.col); err != nil {
			t.Fatalf("ход (%d,%d): %v", m.row, m.col, err)
		}
	}

	state, err := s.State(id)
	if err != nil {
		t.Fatal(err)
	}
	if state.Status != string(game.StatusXWon) {
		t.Errorf("статус %s, ожидали %s", state.Status, game.StatusXWon)
	}
	if state.Board[0][2] != game.X {
		t.Errorf("ход не записался на поле: %v", state.Board)
	}
}

// Оба игрока одновременно пытаются занять одну клетку - ход должен пройти ровно один раз.
func TestConcurrentMoves(t *testing.T) {
	s := New()
	id, tokenX, tokenO := newStartedGame(t, s)

	var wg sync.WaitGroup
	var mu sync.Mutex
	success := 0

	for i := 0; i < 20; i++ {
		token := tokenX
		if i%2 == 1 {
			token = tokenO
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			if _, err := s.Move(id, token, 1, 1); err == nil {
				mu.Lock()
				success++
				mu.Unlock()
			}
			_, _ = s.State(id)
		}()
	}
	wg.Wait()

	if success != 1 {
		t.Errorf("успешных ходов %d, ожидали 1", success)
	}
}
