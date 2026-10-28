package player

import "encoding/json"

type Color string

const (
	White Color = "white"
	Black Color = "black"
)

type Player struct {
	id    int
	name  string
	color Color
}

func NewPlayer(id int, name string, color Color) *Player {
	return &Player{id, name, color}
}

func (p *Player) EntityName() string  { return "player" }
func (p *Player) ID() int             { return p.id }
func (p *Player) Color() Color        { return p.color }
func (p *Player) SetName(name string) { p.name = name }

type playerJSON struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color Color  `json:"color"`
}

func (p *Player) MarshalJSON() ([]byte, error) {
	return json.Marshal(playerJSON{p.id, p.name, p.color})
}

func (p *Player) UnmarshalJSON(data []byte) error {
	var v playerJSON
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	*p = Player{v.ID, v.Name, v.Color}
	return nil
}
