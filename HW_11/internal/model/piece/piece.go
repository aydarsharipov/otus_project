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
	Type   string `json:"type" example:"pawn"`
	Color  string `json:"color" example:"white"`
	Symbol string `json:"symbol" example:"♟"`
}

func NewPiece(pieceType string, color string) *Piece {
	return &Piece{Type: pieceType, Color: color, Symbol: symbols[pieceType][color]}
}
