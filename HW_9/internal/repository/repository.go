package repository

import (
	"encoding/json"
	"fmt"
	"os"

	"HW_9/internal/model"
	"HW_9/internal/model/board"
	"HW_9/internal/model/game"
	"HW_9/internal/model/move"
	"HW_9/internal/model/piece"
	"HW_9/internal/model/player"
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

func (s *PlayersSlice) Add(p *player.Player) error {
	s.Items = append(s.Items, p)
	return saveToFile("players.json", s.Items)
}

func (s *PlayersSlice) Load() error {
	return loadFromFile("players.json", &s.Items)
}

// фигуры

type PiecesSlice struct {
	Items []*piece.Piece
}

func (s *PiecesSlice) Add(p *piece.Piece) error {
	s.Items = append(s.Items, p)
	return saveToFile("pieces.json", s.Items)
}

func (s *PiecesSlice) Load() error {
	return loadFromFile("pieces.json", &s.Items)
}

// ходы

type MovesSlice struct {
	Items []*move.Move
}

func (s *MovesSlice) Add(m *move.Move) error {
	s.Items = append(s.Items, m)
	return saveToFile("moves.json", s.Items)
}

func (s *MovesSlice) Load() error {
	return loadFromFile("moves.json", &s.Items)
}

// доски

type BoardsSlice struct {
	Items []*board.Board
}

func (s *BoardsSlice) Add(b *board.Board) error {
	s.Items = append(s.Items, b)
	return saveToFile("boards.json", s.Items)
}

func (s *BoardsSlice) Load() error {
	return loadFromFile("boards.json", &s.Items)
}

// игры

type GamesSlice struct {
	Items []*game.Game
}

func (s *GamesSlice) Add(g *game.Game) error {
	s.Items = append(s.Items, g)
	return saveToFile("games.json", s.Items)
}

func (s *GamesSlice) Load() error {
	return loadFromFile("games.json", &s.Items)
}

var players PlayersSlice
var pieces PiecesSlice
var moves MovesSlice
var boards BoardsSlice
var games GamesSlice

// загружаем все слайсы из файлов
func LoadAll() error {
	err := os.MkdirAll(dataDir, 0755)
	if err != nil {
		return err
	}

	err = players.Load()
	if err != nil {
		return err
	}
	err = pieces.Load()
	if err != nil {
		return err
	}
	err = moves.Load()
	if err != nil {
		return err
	}
	err = boards.Load()
	if err != nil {
		return err
	}
	err = games.Load()
	if err != nil {
		return err
	}
	return nil
}

func Save(e model.Entity) error {
	switch v := e.(type) {
	case *player.Player:
		return players.Add(v)
	case *piece.Piece:
		return pieces.Add(v)
	case *move.Move:
		return moves.Add(v)
	case *board.Board:
		return boards.Add(v)
	case *game.Game:
		return games.Add(v)
	}
	return fmt.Errorf("неизвестный тип: %T", e)
}

func Run(in <-chan model.Entity) {
	for e := range in {
		err := Save(e)
		if err != nil {
			fmt.Println("ошибка сохранения:", err)
		}
	}
}

func PrintCount() {
	fmt.Println("игроков:", len(players.Items))
	fmt.Println("фигур:", len(pieces.Items))
	fmt.Println("ходов:", len(moves.Items))
	fmt.Println("досок:", len(boards.Items))
	fmt.Println("игр:", len(games.Items))
}
