package piece

type PieceType string

const (
	Pawn   PieceType = "pawn"
	Rook   PieceType = "rook"
	Knight PieceType = "knight"
	Bishop PieceType = "bishop"
	Queen  PieceType = "queen"
	King   PieceType = "king"
)

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
		Pawn:   {"white": '♙', "black": '♟'},
		Rook:   {"white": '♖', "black": '♜'},
		Knight: {"white": '♘', "black": '♜'},
		Bishop: {"white": '♗', "black": '♝'},
		Queen:  {"white": '♕', "black": '♛'},
		King:   {"white": '♔', "black": '♚'},
	}
	return symbols[pieceType][color]
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
