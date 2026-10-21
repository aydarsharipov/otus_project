package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sync"

	"HW_9/internal/model"
	"HW_9/internal/model/board"
	"HW_9/internal/model/game"
	"HW_9/internal/model/move"
	"HW_9/internal/model/piece"
	"HW_9/internal/model/player"
)

// store хранит слайс сущностей и дублирует его в JSON Lines файл:
// одна строка — одна сущность, поэтому добавление — это дозапись в конец файла.
type store[T model.Entity] struct {
	mu    sync.Mutex
	items []T
	path  string
}

func newStore[T model.Entity](fileName string) *store[T] {
	return &store[T]{items: make([]T, 0), path: fileName}
}

// Add сохраняет сущность в слайс и дописывает её в файл.
func (s *store[T]) Add(v T) error {
	data, err := json.Marshal(v)
	if err != nil {
		return fmt.Errorf("сериализация %s: %w", v.EntityName(), err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.OpenFile(s.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
	if err != nil {
		return err
	}
	if _, err := f.Write(append(data, '\n')); err != nil {
		f.Close()
		return err
	}
	if err := f.Close(); err != nil {
		return err
	}

	s.items = append(s.items, v)
	return nil
}

// Load заполняет слайс данными из файла. Отсутствие файла — не ошибка.
func (s *store[T]) Load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	f, err := os.Open(s.path)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	if err != nil {
		return err
	}
	defer f.Close()

	items := make([]T, 0)
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 64*1024), 10*1024*1024)
	for line := 1; scanner.Scan(); line++ {
		if len(scanner.Bytes()) == 0 {
			continue
		}
		var v T
		if err := json.Unmarshal(scanner.Bytes(), &v); err != nil {
			return fmt.Errorf("%s, строка %d: %w", s.path, line, err)
		}
		items = append(items, v)
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	s.items = items
	return nil
}

func (s *store[T]) All() []T {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]T(nil), s.items...)
}

func (s *store[T]) Len() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.items)
}

var (
	players *store[*player.Player]
	pieces  *store[*piece.Piece]
	moves   *store[*move.Move]
	boards  *store[*board.Board]
	games   *store[*game.Game]
)

// Init создаёт каталог с данными и при рестарте наполняет слайсы из файлов.
func Init(dir string) error {
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}

	players = newStore[*player.Player](filepath.Join(dir, "players.jsonl"))
	pieces = newStore[*piece.Piece](filepath.Join(dir, "pieces.jsonl"))
	moves = newStore[*move.Move](filepath.Join(dir, "moves.jsonl"))
	boards = newStore[*board.Board](filepath.Join(dir, "boards.jsonl"))
	games = newStore[*game.Game](filepath.Join(dir, "games.jsonl"))

	return errors.Join(players.Load(), pieces.Load(), moves.Load(), boards.Load(), games.Load())
}

func Save(e model.Entity) error {
	switch v := e.(type) {
	case *player.Player:
		return players.Add(v)
	case *piece.Piece:
		return pieces.Add(v)
	case *move.Move:
		return moves.Add(v)
	case *board.Board:
		return boards.Add(v)
	case *game.Game:
		return games.Add(v)
	default:
		return fmt.Errorf("неизвестный тип сущности: %T", e)
	}
}

func Run(in <-chan model.Entity) {
	for e := range in {
		if err := Save(e); err != nil {
			log.Println("Ошибка сохранения:", err)
		}
	}
}

// MaxGameID нужен, чтобы после рестарта новые игры не получали уже занятые id.
func MaxGameID() int {
	maxID := 0
	for _, g := range games.All() {
		maxID = max(maxID, g.ID())
	}
	return maxID
}

func Stats() string {
	return fmt.Sprintf("игроков: %d, фигур: %d, ходов: %d, досок: %d, игр: %d",
		players.Len(), pieces.Len(), moves.Len(), boards.Len(), games.Len())
}
