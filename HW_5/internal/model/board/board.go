package board

import "HW_5/internal/model/piece"

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
	size int
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

func (b *Board) Size() int {
	return b.size
}

func (b *Board) GetCell(row, col int) *Cell {
	if row < 0 || row >= b.size || col < 0 || col >= b.size {
		return nil
	}
	return b.cells[row][col]
}

func (b *Board) SetPiece(row, col int, p *piece.Piece) bool {
	if row < 0 || row >= b.size || col < 0 || col >= b.size {
		return false
	}
	b.cells[row][col].SetPiece(p)
	return true
}

func (b *Board) GetAllCells() [][]*Cell {
	return b.cells
}
