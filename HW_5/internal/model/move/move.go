package move

import (
	"time"

	"HW_5/internal/model/piece"
)

type MoveStatus string

const (
	Pending   MoveStatus = "pending"
	Completed MoveStatus = "completed"
	Invalid   MoveStatus = "invalid"
)

type Move struct {
	id          int
	fromRow     int
	fromCol     int
	toRow       int
	toCol       int
	piece       *piece.Piece
	status      MoveStatus
	createdAt   time.Time
	description string
}

func NewMove(id, fromRow, fromCol, toRow, toCol int, p *piece.Piece) *Move {
	return &Move{
		id:          id,
		fromRow:     fromRow,
		fromCol:     fromCol,
		toRow:       toRow,
		toCol:       toCol,
		piece:       p,
		status:      Pending,
		createdAt:   time.Now(),
		description: "",
	}
}

func (m *Move) ID() int {
	return m.id
}

func (m *Move) FromRow() int {
	return m.fromRow
}

func (m *Move) FromCol() int {
	return m.fromCol
}

func (m *Move) ToRow() int {
	return m.toRow
}

func (m *Move) ToCol() int {
	return m.toCol
}

func (m *Move) Piece() *piece.Piece {
	return m.piece
}

func (m *Move) Status() MoveStatus {
	return m.status
}

func (m *Move) CreatedAt() time.Time {
	return m.createdAt
}

func (m *Move) Description() string {
	return m.description
}

func (m *Move) Complete() {
	m.status = Completed
	m.description = "Move completed"
}

func (m *Move) Invalidate(reason string) {
	m.status = Invalid
	m.description = reason
}
