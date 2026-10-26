package move

import (
	"HW_11/internal/model/board"
	"HW_11/internal/model/piece"
)

type Move struct {
	ID       int          `json:"id" example:"1"`
	GameID   int          `json:"game_id" example:"1"`
	FromRow  int          `json:"from_row" example:"6"`
	FromCol  int          `json:"from_col" example:"4"`
	ToRow    int          `json:"to_row" example:"4"`
	ToCol    int          `json:"to_col" example:"4"`
	Piece    *piece.Piece `json:"piece"`
	Notation string       `json:"notation" example:"♟ E2-E4"`
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
