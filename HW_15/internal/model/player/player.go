package player

const (
	White = "white"
	Black = "black"
)

type Player struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

func NewPlayer(id int, name string, color string) *Player {
	return &Player{ID: id, Name: name, Color: color}
}
