package move

import (
	"encoding/json"
	"fmt"
	"time"

	"HW_9/internal/model/board"
	"HW_9/internal/model/piece"
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
	duration    time.Duration
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

func (m *Move) EntityName() string {
	return "move"
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

func (m *Move) Duration() time.Duration {
	return m.duration
}

func (m *Move) SetDuration(d time.Duration) {
	m.duration = d
}

func (m *Move) Complete() {
	m.status = Completed
	m.description = "Move completed"
}

func (m *Move) Invalidate(reason string) {
	m.status = Invalid
	m.description = reason
}

func (m *Move) Notation() string {
	return fmt.Sprintf("%c %s-%s", m.piece.Symbol(),
		board.SquareName(m.fromRow, m.fromCol), board.SquareName(m.toRow, m.toCol))
}

type moveJSON struct {
	ID          int           `json:"id"`
	FromRow     int           `json:"from_row"`
	FromCol     int           `json:"from_col"`
	ToRow       int           `json:"to_row"`
	ToCol       int           `json:"to_col"`
	Piece       *piece.Piece  `json:"piece"`
	Status      MoveStatus    `json:"status"`
	CreatedAt   time.Time     `json:"created_at"`
	Description string        `json:"description"`
	Duration    time.Duration `json:"duration"`
}

func (m *Move) MarshalJSON() ([]byte, error) {
	return json.Marshal(moveJSON{
		ID: m.id, FromRow: m.fromRow, FromCol: m.fromCol, ToRow: m.toRow, ToCol: m.toCol,
		Piece: m.piece, Status: m.status, CreatedAt: m.createdAt,
		Description: m.description, Duration: m.duration,
	})
}

func (m *Move) UnmarshalJSON(data []byte) error {
	var v moveJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*m = Move{
		id: v.ID, fromRow: v.FromRow, fromCol: v.FromCol, toRow: v.ToRow, toCol: v.ToCol,
		piece: v.Piece, status: v.Status, createdAt: v.CreatedAt,
		description: v.Description, duration: v.Duration,
	}
	return nil
}
