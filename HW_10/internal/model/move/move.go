package move

import (
	"encoding/json"
	"fmt"

	"HW_10/internal/model/board"
	"HW_10/internal/model/piece"
)

type Move struct {
	id                             int
	gameID                         int
	fromRow, fromCol, toRow, toCol int
	piece                          *piece.Piece
}

func NewMove(id, gameID, fromRow, fromCol, toRow, toCol int, p *piece.Piece) *Move {
	return &Move{id, gameID, fromRow, fromCol, toRow, toCol, p}
}

func (m *Move) EntityName() string { return "move" }
func (m *Move) ID() int            { return m.id }
func (m *Move) GameID() int        { return m.gameID }

func (m *Move) Notation() string {
	return fmt.Sprintf("%c %s-%s", m.piece.Symbol(),
		board.SquareName(m.fromRow, m.fromCol), board.SquareName(m.toRow, m.toCol))
}

type moveJSON struct {
	ID       int          `json:"id"`
	GameID   int          `json:"game_id"`
	FromRow  int          `json:"from_row"`
	FromCol  int          `json:"from_col"`
	ToRow    int          `json:"to_row"`
	ToCol    int          `json:"to_col"`
	Piece    *piece.Piece `json:"piece"`
	Notation string       `json:"notation"`
}

func (m *Move) MarshalJSON() ([]byte, error) {
	return json.Marshal(moveJSON{m.id, m.gameID, m.fromRow, m.fromCol, m.toRow, m.toCol, m.piece, m.Notation()})
}

func (m *Move) UnmarshalJSON(data []byte) error {
	var v moveJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Move{v.ID, v.GameID, v.FromRow, v.FromCol, v.ToRow, v.ToCol, v.Piece}
	return nil
}
