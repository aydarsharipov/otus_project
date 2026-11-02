package game

import (
	"errors"

	"HW_13/internal/model/board"
	"HW_13/internal/model/move"
	"HW_13/internal/model/piece"
	"HW_13/internal/model/player"
)

// ошибки
var (
	ErrGameFinished = errors.New("игра закончена")
	ErrNoPiece      = errors.New("там нет фигуры")
	ErrNotYourTurn  = errors.New("сейчас не ваш ход")
	ErrOwnPiece     = errors.New("там своя фигура")
)

type Game struct {
	ID     int            `json:"id"`
	White  *player.Player `json:"white"`
	Black  *player.Player `json:"black"`
	Board  *board.Board   `json:"board"`
	Moves  []*move.Move   `json:"moves"`
	Turn   string         `json:"turn"`
	Status string         `json:"status"` // active или finished
	Winner string         `json:"winner,omitempty"`
}

func NewGame(id int, white, black *player.Player) *Game {
	g := &Game{
		ID:     id,
		White:  white,
		Black:  black,
		Board:  board.NewBoard(),
		Moves:  []*move.Move{},
		Turn:   player.White,
		Status: "active",
	}

	backRow := []string{piece.Rook, piece.Knight, piece.Bishop, piece.Queen,
		piece.King, piece.Bishop, piece.Knight, piece.Rook}
	for col := 0; col < board.Size; col++ {
		g.Board.Cells[0][col] = piece.NewPiece(backRow[col], "black")
		g.Board.Cells[1][col] = piece.NewPiece(piece.Pawn, "black")
		g.Board.Cells[6][col] = piece.NewPiece(piece.Pawn, "white")
		g.Board.Cells[7][col] = piece.NewPiece(backRow[col], "white")
	}
	return g
}

func (g *Game) MakeMove(moveID, fromRow, fromCol, toRow, toCol int) (*move.Move, error) {
	if g.Status != "active" {
		return nil, ErrGameFinished
	}
	p := g.Board.Cells[fromRow][fromCol]
	if p == nil {
		return nil, ErrNoPiece
	}
	if p.Color != g.Turn {
		return nil, ErrNotYourTurn
	}
	target := g.Board.Cells[toRow][toCol]
	if target != nil && target.Color == g.Turn {
		return nil, ErrOwnPiece
	}

	g.Board.Cells[toRow][toCol] = p
	g.Board.Cells[fromRow][fromCol] = nil
	m := move.NewMove(moveID, g.ID, fromRow, fromCol, toRow, toCol, p)
	g.Moves = append(g.Moves, m)

	if g.Turn == player.White {
		g.Turn = player.Black
	} else {
		g.Turn = player.White
	}
	return m, nil
}

// сдаётся тот кто ходит
func (g *Game) Surrender() error {
	if g.Status != "active" {
		return ErrGameFinished
	}
	g.Status = "finished"
	if g.Turn == player.White {
		g.Winner = player.Black
	} else {
		g.Winner = player.White
	}
	return nil
}
