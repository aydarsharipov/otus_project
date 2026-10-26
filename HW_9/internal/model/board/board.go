package board

import (
	"encoding/json"
	"fmt"

	"HW_9/internal/model/piece"
)

const StandardSize = 8

type Cell struct {
	row   int
	col   int
	piece *piece.Piece
}

func NewCell(row, col int) *Cell {
	return &Cell{
		row:   row,
		col:   col,
		piece: nil,
	}
}

func (c *Cell) Row() int {
	return c.row
}

func (c *Cell) Col() int {
	return c.col
}

func (c *Cell) Piece() *piece.Piece {
	return c.piece
}

func (c *Cell) SetPiece(p *piece.Piece) {
	c.piece = p
}

func (c *Cell) IsEmpty() bool {
	return c.piece == nil
}

type Board struct {
	size  int
	cells [][]*Cell
}

func NewBoard(size int) *Board {
	b := &Board{
		size:  size,
		cells: make([][]*Cell, size),
	}
	for i := 0; i < size; i++ {
		b.cells[i] = make([]*Cell, size)
		for j := 0; j < size; j++ {
			b.cells[i][j] = NewCell(i, j)
		}
	}
	return b
}

func (b *Board) EntityName() string {
	return "board"
}

func (b *Board) Size() int {
	return b.size
}

func (b *Board) InBounds(row, col int) bool {
	return row >= 0 && row < b.size && col >= 0 && col < b.size
}

func (b *Board) GetCell(row, col int) *Cell {
	if !b.InBounds(row, col) {
		return nil
	}
	return b.cells[row][col]
}

func (b *Board) SetPiece(row, col int, p *piece.Piece) bool {
	if !b.InBounds(row, col) {
		return false
	}
	b.cells[row][col].SetPiece(p)
	return true
}

func (b *Board) GetAllCells() [][]*Cell {
	return b.cells
}

func SquareName(row, col int) string {
	return fmt.Sprintf("%c%d", 'A'+col, StandardSize-row)
}

func ParseSquare(s string) (row, col int, ok bool) {
	if len(s) != 2 {
		return 0, 0, false
	}
	col = int(s[0]) - 'A'
	rank := int(s[1]) - '0'
	if col < 0 || col >= StandardSize || rank < 1 || rank > StandardSize {
		return 0, 0, false
	}
	return StandardSize - rank, col, true
}

type cellJSON struct {
	Row   int          `json:"row"`
	Col   int          `json:"col"`
	Piece *piece.Piece `json:"piece"`
}

type boardJSON struct {
	Size  int        `json:"size"`
	Cells []cellJSON `json:"cells"`
}

func (b *Board) MarshalJSON() ([]byte, error) {
	v := boardJSON{Size: b.size, Cells: make([]cellJSON, 0)}
	for _, row := range b.cells {
		for _, c := range row {
			if !c.IsEmpty() {
				v.Cells = append(v.Cells, cellJSON{Row: c.row, Col: c.col, Piece: c.piece})
			}
		}
	}
	return json.Marshal(v)
}

func (b *Board) UnmarshalJSON(data []byte) error {
	var v boardJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*b = *NewBoard(v.Size)
	for _, c := range v.Cells {
		b.SetPiece(c.Row, c.Col, c.Piece)
	}
	return nil
}
