package player

const (
	White = "white"
	Black = "black"
)

type Player struct {
	ID    int    `json:"id" example:"1"`
	Name  string `json:"name" example:"Аня"`
	Color string `json:"color" example:"white"`
}

func NewPlayer(id int, name string, color string) *Player {
	return &Player{ID: id, Name: name, Color: color}
}
