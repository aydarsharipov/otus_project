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

var AllTypes = []PieceType{Pawn, Rook, Knight, Bishop, Queen, King}

type Piece struct {
	pieceType PieceType
	color     string
	symbol    rune
}

func NewPiece(pieceType PieceType, color string) *Piece {
	return &Piece{
		pieceType: pieceType,
		color:     color,
		symbol:    getSymbol(pieceType, color),
	}
}

func getSymbol(pieceType PieceType, color string) rune {
	symbols := map[PieceType]map[string]rune{
		Pawn:   {"white": '♟', "black": '♙'},
		Rook:   {"white": '♜', "black": '♖'},
		Knight: {"white": '♞', "black": '♘'},
		Bishop: {"white": '♝', "black": '♗'},
		Queen:  {"white": '♛', "black": '♕'},
		King:   {"white": '♚', "black": '♔'},
	}
	return symbols[pieceType][color]
}

func (p *Piece) EntityName() string {
	return "piece"
}

func (p *Piece) Type() PieceType {
	return p.pieceType
}

func (p *Piece) Color() string {
	return p.color
}

func (p *Piece) Symbol() rune {
	return p.symbol
}

type pieceJSON struct {
	Type  PieceType `json:"type"`
	Color string    `json:"color"`
}

func (p *Piece) MarshalJSON() ([]byte, error) {
	return json.Marshal(pieceJSON{Type: p.pieceType, Color: p.color})
}

func (p *Piece) UnmarshalJSON(data []byte) error {
	var v pieceJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*p = *NewPiece(v.Type, v.Color)
	return nil
}
