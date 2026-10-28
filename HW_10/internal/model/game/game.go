package game

import (
	"encoding/json"
	"errors"

	"HW_10/internal/model/board"
	"HW_10/internal/model/move"
	"HW_10/internal/model/piece"
	"HW_10/internal/model/player"
)

type Game struct {
	id     int
	white  *player.Player
	black  *player.Player
	board  *board.Board
	moves  []*move.Move
	turn   player.Color
	status string // active / finished
	winner player.Color
}

func NewGame(id int, white, black *player.Player) *Game {
	g := &Game{id: id, white: white, black: black, board: board.NewBoard(),
		moves: []*move.Move{}, turn: player.White, status: "active"}
	back := []piece.PieceType{piece.Rook, piece.Knight, piece.Bishop, piece.Queen,
		piece.King, piece.Bishop, piece.Knight, piece.Rook}
	for col, t := range back {
		g.board.Set(0, col, piece.NewPiece(t, "black"))
		g.board.Set(1, col, piece.NewPiece(piece.Pawn, "black"))
		g.board.Set(6, col, piece.NewPiece(piece.Pawn, "white"))
		g.board.Set(7, col, piece.NewPiece(t, "white"))
	}
	return g
}

func (g *Game) EntityName() string { return "game" }
func (g *Game) ID() int            { return g.id }

func (g *Game) HasPlayer(id int) bool {
	return g.white.ID() == id || g.black.ID() == id
}

func (g *Game) MakeMove(moveID, fromRow, fromCol, toRow, toCol int) (*move.Move, error) {
	if g.status != "active" {
		return nil, errors.New("игра закончена")
	}
	p := g.board.Get(fromRow, fromCol)
	if p == nil {
		return nil, errors.New("там нет фигуры")
	}
	if p.Color() != string(g.turn) {
		return nil, errors.New("сейчас не ваш ход")
	}
	if dst := g.board.Get(toRow, toCol); dst != nil && dst.Color() == string(g.turn) {
		return nil, errors.New("там своя фигура")
	}

	g.board.Set(toRow, toCol, p)
	g.board.Set(fromRow, fromCol, nil)
	m := move.NewMove(moveID, g.id, fromRow, fromCol, toRow, toCol, p)
	g.moves = append(g.moves, m)

	if g.turn == player.White {
		g.turn = player.Black
	} else {
		g.turn = player.White
	}
	return m, nil
}

// сдаётся тот кто ходит
func (g *Game) Surrender() error {
	if g.status != "active" {
		return errors.New("игра закончена")
	}
	g.status = "finished"
	if g.turn == player.White {
		g.winner = player.Black
	} else {
		g.winner = player.White
	}
	return nil
}

type gameJSON struct {
	ID     int            `json:"id"`
	White  *player.Player `json:"white"`
	Black  *player.Player `json:"black"`
	Board  *board.Board   `json:"board"`
	Moves  []*move.Move   `json:"moves"`
	Turn   player.Color   `json:"turn"`
	Status string         `json:"status"`
	Winner player.Color   `json:"winner,omitempty"`
}

func (g *Game) MarshalJSON() ([]byte, error) {
	return json.Marshal(gameJSON{g.id, g.white, g.black, g.board, g.moves, g.turn, g.status, g.winner})
}

func (g *Game) UnmarshalJSON(data []byte) error {
	var v gameJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*g = Game{v.ID, v.White, v.Black, v.Board, v.Moves, v.Turn, v.Status, v.Winner}
	return nil
}
