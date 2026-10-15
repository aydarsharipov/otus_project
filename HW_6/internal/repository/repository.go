package repository

import (
	"fmt"

	"HW_6/internal/model"
	"HW_6/internal/model/board"
	"HW_6/internal/model/game"
	"HW_6/internal/model/move"
	"HW_6/internal/model/piece"
	"HW_6/internal/model/player"
)

// Слайсы для хранения сущностей, предварительно инициализированные в пакете
var (
	players = make([]*player.Player, 0)
	pieces  = make([]*piece.Piece, 0)
	moves   = make([]*move.Move, 0)
	boards  = make([]*board.Board, 0)
	games   = make([]*game.Game, 0)
)

func Save(e model.Entity) error {
	switch v := e.(type) {
	case *player.Player:
		players = append(players, v)
	case *piece.Piece:
		pieces = append(pieces, v)
	case *move.Move:
		moves = append(moves, v)
	case *board.Board:
		boards = append(boards, v)
	case *game.Game:
		games = append(games, v)
	default:
		return fmt.Errorf("неизвестный тип сущности: %T", e)
	}
	return nil
}
func Stats() map[string]int {
	return map[string]int{
		"player": len(players),
		"piece":  len(pieces),
		"move":   len(moves),
		"board":  len(boards),
		"game":   len(games),
	}
}
