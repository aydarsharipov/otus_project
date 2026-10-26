package board

import (
	"encoding/json"
	"fmt"

	"HW_10/internal/model/piece"
)

const Size = 8

type Board struct {
	cells [Size][Size]*piece.Piece
}

func NewBoard() *Board {
	return &Board{}
}

func (b *Board) EntityName() string { return "board" }

func (b *Board) Get(row, col int) *piece.Piece {
	return b.cells[row][col]
}

func (b *Board) Set(row, col int, p *piece.Piece) {
	b.cells[row][col] = p
}

func SquareName(row, col int) string {
	return fmt.Sprintf("%c%d", 'A'+col, Size-row)
}

// "E2" -> row 6, col 4
func ParseSquare(s string) (row, col int, ok bool) {
	if len(s) != 2 {
		return 0, 0, false
	}
	col = int(s[0]) - 'A'
	rank := int(s[1]) - '0'
	if col < 0 || col >= Size || rank < 1 || rank > Size {
		return 0, 0, false
	}
	return Size - rank, col, true
}

type cellJSON struct {
	Row   int          `json:"row"`
	Col   int          `json:"col"`
	Piece *piece.Piece `json:"piece"`
}

func (b *Board) MarshalJSON() ([]byte, error) {
	cells := []cellJSON{}
	for r := 0; r < Size; r++ {
		for c := 0; c < Size; c++ {
			if b.cells[r][c] != nil {
				cells = append(cells, cellJSON{r, c, b.cells[r][c]})
			}
		}
	}
	return json.Marshal(map[string]any{"cells": cells})
}

func (b *Board) UnmarshalJSON(data []byte) error {
	var v struct {
		Cells []cellJSON `json:"cells"`
	}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*b = Board{}
	for _, c := range v.Cells {
		b.cells[c.Row][c.Col] = c.Piece
	}
	return nil
}
