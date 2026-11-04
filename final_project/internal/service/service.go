// Package service хранит партии и связывает игроков с их символами.
package service

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"sync"

	"tictactoe/internal/api"
	"tictactoe/internal/game"
)

// StatusWaiting - партия создана, но второй игрок ещё не присоединился.
const StatusWaiting = "waiting"

// MaxGames - сколько партий сервер может держать в памяти одновременно.
const MaxGames = 1000

// Ошибки сервиса.
var (
	ErrGameNotFound = errors.New("игра не найдена")
	ErrGameFull     = errors.New("в игре уже два игрока")
	ErrBadToken     = errors.New("неверный токен игрока")
	ErrWaiting      = errors.New("ждём второго игрока")
	ErrTooManyGames = errors.New("слишком много игр на сервере")
)

type session struct {
	game   *game.Game
	tokenX string
	tokenO string
}

// Service управляет всеми партиями. Безопасен для использования из нескольких горутин.
type Service struct {
	mu    sync.Mutex
	games map[string]*session
}

// New создаёт пустой сервис.
func New() *Service {
	return &Service{games: make(map[string]*session)}
}

// CreateGame создаёт новую партию. Создатель играет за X.
func (s *Service) CreateGame(size int) (api.JoinResponse, error) {
	g, err := game.New(size)
	if err != nil {
		return api.JoinResponse{}, err
	}

	id, err := randomString(8)
	if err != nil {
		return api.JoinResponse{}, err
	}
	token, err := randomString(16)
	if err != nil {
		return api.JoinResponse{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if len(s.games) >= MaxGames {
		return api.JoinResponse{}, ErrTooManyGames
	}
	s.games[id] = &session{game: g, tokenX: token}

	return api.JoinResponse{GameID: id, PlayerToken: token, Symbol: game.X}, nil
}

// JoinGame добавляет второго игрока (O) в партию.
func (s *Service) JoinGame(id string) (api.JoinResponse, error) {
	token, err := randomString(16)
	if err != nil {
		return api.JoinResponse{}, err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.games[id]
	if !ok {
		return api.JoinResponse{}, ErrGameNotFound
	}
	if sess.tokenO != "" {
		return api.JoinResponse{}, ErrGameFull
	}
	sess.tokenO = token

	return api.JoinResponse{GameID: id, PlayerToken: token, Symbol: game.O}, nil
}

// State возвращает состояние партии.
func (s *Service) State(id string) (api.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.games[id]
	if !ok {
		return api.GameState{}, ErrGameNotFound
	}
	return makeState(id, sess), nil
}

// Move делает ход от имени игрока с токеном token.
func (s *Service) Move(id, token string, row, col int) (api.GameState, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	sess, ok := s.games[id]
	if !ok {
		return api.GameState{}, ErrGameNotFound
	}

	var player string
	switch {
	case sameToken(token, sess.tokenX):
		player = game.X
	case sameToken(token, sess.tokenO):
		player = game.O
	default:
		return api.GameState{}, ErrBadToken
	}

	if sess.tokenO == "" {
		return api.GameState{}, ErrWaiting
	}

	if err := sess.game.Move(player, row, col); err != nil {
		return api.GameState{}, err
	}
	return makeState(id, sess), nil
}

func makeState(id string, sess *session) api.GameState {
	status := string(sess.game.Status())
	if sess.tokenO == "" {
		status = StatusWaiting
	}

	return api.GameState{
		GameID: id,
		Size:   sess.game.Size(),
		Board:  sess.game.Board(),
		Turn:   sess.game.Turn(),
		Status: status,
	}
}

// sameToken сравнивает токены за постоянное время. Пустой токен никогда не подходит.
func sameToken(got, want string) bool {
	if got == "" || want == "" {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(got), []byte(want)) == 1
}

func randomString(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
