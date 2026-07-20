package player

import "time"

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
