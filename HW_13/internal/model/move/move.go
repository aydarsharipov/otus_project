package move

import (
	"HW_13/internal/model/board"
	"HW_13/internal/model/piece"
)

type Move struct {
	ID       int          `json:"id"`
	GameID   int          `json:"game_id"`
	FromRow  int          `json:"from_row"`
	FromCol  int          `json:"from_col"`
	ToRow    int          `json:"to_row"`
	ToCol    int          `json:"to_col"`
	Piece    *piece.Piece `json:"piece"`
	Notation string       `json:"notation"`
}

func NewMove(id, gameID, fromRow, fromCol, toRow, toCol int, p *piece.Piece) *Move {
	m := &Move{
		ID:      id,
		GameID:  gameID,
		FromRow: fromRow,
		FromCol: fromCol,
		ToRow:   toRow,
		ToCol:   toCol,
		Piece:   p,
	}
	m.Notation = p.Symbol + " " + board.SquareName(fromRow, fromCol) + "-" + board.SquareName(toRow, toCol)
	return m
}
