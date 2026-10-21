package player

import (
	"encoding/json"
	"time"
)

type Color string

const (
	White Color = "white"
	Black Color = "black"
)

type Player struct {
	id        int
	name      string
	color     Color
	createdAt time.Time
}

func NewPlayer(id int, name string, color Color) *Player {
	return &Player{
		id:        id,
		name:      name,
		color:     color,
		createdAt: time.Now(),
	}
}

func (p *Player) EntityName() string {
	return "player"
}

func (p *Player) ID() int {
	return p.id
}

func (p *Player) Name() string {
	return p.name
}

func (p *Player) Color() Color {
	return p.color
}

func (p *Player) CreatedAt() time.Time {
	return p.createdAt
}

func (p *Player) SetName(name string) {
	p.name = name
}

type playerJSON struct {
	ID        int       `json:"id"`
	Name      string    `json:"name"`
	Color     Color     `json:"color"`
	CreatedAt time.Time `json:"created_at"`
}

func (p *Player) MarshalJSON() ([]byte, error) {
	return json.Marshal(playerJSON{ID: p.id, Name: p.name, Color: p.color, CreatedAt: p.createdAt})
}

func (p *Player) UnmarshalJSON(data []byte) error {
	var v playerJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*p = Player{id: v.ID, name: v.Name, color: v.Color, createdAt: v.CreatedAt}
	return nil
}
