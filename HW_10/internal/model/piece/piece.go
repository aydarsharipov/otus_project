package piece

import "encoding/json"

type PieceType string

const (
	Pawn   PieceType = "pawn"
	Rook   PieceType = "rook"
	Knight PieceType = "knight"
	Bishop PieceType = "bishop"
	Queen  PieceType = "queen"
	King   PieceType = "king"
)

var symbols = map[PieceType]map[string]rune{
	Pawn:   {"white": '♟', "black": '♙'},
	Rook:   {"white": '♜', "black": '♖'},
	Knight: {"white": '♞', "black": '♘'},
	Bishop: {"white": '♝', "black": '♗'},
	Queen:  {"white": '♛', "black": '♕'},
	King:   {"white": '♚', "black": '♔'},
}

type Piece struct {
	pieceType PieceType
	color     string
}

func NewPiece(t PieceType, color string) *Piece {
	return &Piece{pieceType: t, color: color}
}

func (p *Piece) EntityName() string { return "piece" }
func (p *Piece) Color() string      { return p.color }
func (p *Piece) Symbol() rune       { return symbols[p.pieceType][p.color] }

type pieceJSON struct {
	Type   PieceType `json:"type"`
	Color  string    `json:"color"`
	Symbol string    `json:"symbol"`
}

func (p *Piece) MarshalJSON() ([]byte, error) {
	return json.Marshal(pieceJSON{p.pieceType, p.color, string(p.Symbol())})
}

func (p *Piece) UnmarshalJSON(data []byte) error {
	var v pieceJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*p = Piece{v.Type, v.Color}
	return nil
}
