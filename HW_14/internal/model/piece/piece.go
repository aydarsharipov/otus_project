package piece

const (
	Pawn   = "pawn"
	Rook   = "rook"
	Knight = "knight"
	Bishop = "bishop"
	Queen  = "queen"
	King   = "king"
)

var symbols = map[string]map[string]string{
	Pawn:   {"white": "♟", "black": "♙"},
	Rook:   {"white": "♜", "black": "♖"},
	Knight: {"white": "♞", "black": "♘"},
	Bishop: {"white": "♝", "black": "♗"},
	Queen:  {"white": "♛", "black": "♕"},
	King:   {"white": "♚", "black": "♔"},
}

type Piece struct {
	Type   string `json:"type"`
	Color  string `json:"color"`
	Symbol string `json:"symbol"`
}

func NewPiece(pieceType string, color string) *Piece {
	return &Piece{Type: pieceType, Color: color, Symbol: symbols[pieceType][color]}
}
