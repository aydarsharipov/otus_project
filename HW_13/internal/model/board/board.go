package board

import (
	"fmt"

	"HW_13/internal/model/piece"
)

const Size = 8

// клетки доски
type Board struct {
	Cells [][]*piece.Piece `json:"cells"`
}

func NewBoard() *Board {
	b := &Board{}
	for i := 0; i < Size; i++ {
		b.Cells = append(b.Cells, make([]*piece.Piece, Size))
	}
	return b
}

func SquareName(row, col int) string {
	return fmt.Sprintf("%c%d", 'A'+col, Size-row)
}

// "E2" -> строка 6, столбец 4
func ParseSquare(s string) (int, int, bool) {
	if len(s) != 2 {
		return 0, 0, false
	}
	col := int(s[0]) - 'A'
	rank := int(s[1]) - '0'
	if col < 0 || col >= Size || rank < 1 || rank > Size {
		return 0, 0, false
	}
	return Size - rank, col, true
}
