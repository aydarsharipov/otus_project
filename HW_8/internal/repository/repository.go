package repository

import (
	"fmt"
	"log"
	"sync"

	"HW_8/internal/model"
	"HW_8/internal/model/board"
	"HW_8/internal/model/game"
	"HW_8/internal/model/move"
	"HW_8/internal/model/piece"
	"HW_8/internal/model/player"
)

type store[T model.Entity] struct {
	mu    sync.Mutex
	items []T
}

func newStore[T model.Entity]() *store[T] {
	return &store[T]{items: make([]T, 0)}
}

func (s *store[T]) add(v T) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = append(s.items, v)
}

var (
	players = newStore[*player.Player]()
	pieces  = newStore[*piece.Piece]()
	moves   = newStore[*move.Move]()
	boards  = newStore[*board.Board]()
	games   = newStore[*game.Game]()
)

func Save(e model.Entity) error {
	switch v := e.(type) {
	case *player.Player:
		players.add(v)
	case *piece.Piece:
		pieces.add(v)
	case *move.Move:
		moves.add(v)
	case *board.Board:
		boards.add(v)
	case *game.Game:
		games.add(v)
	default:
		return fmt.Errorf("неизвестный тип сущности: %T", e)
	}
	return nil
}

func Run(in <-chan model.Entity) {
	for e := range in {
		if err := Save(e); err != nil {
			log.Println("Ошибка сохранения:", err)
		}
	}
}
