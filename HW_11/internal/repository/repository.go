package repository

import (
	"encoding/json"
	"os"

	"HW_11/internal/model/game"
	"HW_11/internal/model/move"
	"HW_11/internal/model/player"
)

// папка с файлами
var dataDir = "data"

// сохраняем слайс в файл
func saveToFile(fileName string, data any) error {
	bytes, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(dataDir+"/"+fileName, bytes, 0644)
}

// читаем слайс из файла
func loadFromFile(fileName string, data any) error {
	bytes, err := os.ReadFile(dataDir + "/" + fileName)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(bytes, data)
}

// игроки

type PlayersSlice struct {
	Items []*player.Player
}

func (s *PlayersSlice) Load() error {
	err := loadFromFile("players.json", &s.Items)
	// чтобы был [] а не null
	if s.Items == nil {
		s.Items = []*player.Player{}
	}
	return err
}

func (s *PlayersSlice) Save() error {
	return saveToFile("players.json", s.Items)
}

func (s *PlayersSlice) NextID() int {
	maxID := 0
	for _, p := range s.Items {
		if p.ID > maxID {
			maxID = p.ID
		}
	}
	return maxID + 1
}

func (s *PlayersSlice) Add(p *player.Player) error {
	s.Items = append(s.Items, p)
	return s.Save()
}

func (s *PlayersSlice) Get(id int) *player.Player {
	for _, p := range s.Items {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (s *PlayersSlice) Delete(id int) bool {
	for i, p := range s.Items {
		if p.ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

// игры

type GamesSlice struct {
	Items []*game.Game
}

func (s *GamesSlice) Load() error {
	err := loadFromFile("games.json", &s.Items)
	// чтобы был [] а не null
	if s.Items == nil {
		s.Items = []*game.Game{}
	}
	return err
}

func (s *GamesSlice) Save() error {
	return saveToFile("games.json", s.Items)
}

func (s *GamesSlice) NextID() int {
	maxID := 0
	for _, g := range s.Items {
		if g.ID > maxID {
			maxID = g.ID
		}
	}
	return maxID + 1
}

func (s *GamesSlice) Add(g *game.Game) error {
	s.Items = append(s.Items, g)
	return s.Save()
}

func (s *GamesSlice) Get(id int) *game.Game {
	for _, g := range s.Items {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func (s *GamesSlice) Delete(id int) bool {
	for i, g := range s.Items {
		if g.ID == id {
			s.Items = append(s.Items[:i], s.Items[i+1:]...)
			return true
		}
	}
	return false
}

// ходы

type MovesSlice struct {
	Items []*move.Move
}

func (s *MovesSlice) Load() error {
	err := loadFromFile("moves.json", &s.Items)
	// чтобы был [] а не null
	if s.Items == nil {
		s.Items = []*move.Move{}
	}
	return err
}

func (s *MovesSlice) Save() error {
	return saveToFile("moves.json", s.Items)
}

func (s *MovesSlice) NextID() int {
	maxID := 0
	for _, m := range s.Items {
		if m.ID > maxID {
			maxID = m.ID
		}
	}
	return maxID + 1
}

func (s *MovesSlice) Add(m *move.Move) error {
	s.Items = append(s.Items, m)
	return s.Save()
}

func (s *MovesSlice) Get(id int) *move.Move {
	for _, m := range s.Items {
		if m.ID == id {
			return m
		}
	}
	return nil
}

var Players PlayersSlice
var Games GamesSlice
var Moves MovesSlice

// загружаем все слайсы из файлов
func LoadAll() error {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return err
	}

	err = Players.Load()
	if err != nil {
		return err
	}
	err = Games.Load()
	if err != nil {
		return err
	}
	err = Moves.Load()
	if err != nil {
		return err
	}
	return nil
}
