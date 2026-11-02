package service

import (
	"errors"
	"strings"

	"HW_13/internal/model/board"
	"HW_13/internal/model/game"
	"HW_13/internal/model/move"
	"HW_13/internal/model/player"
)

// интерфейсы для хранилищ

type PlayerStorage interface {
	NextID() int
	Add(p *player.Player) error
	Get(id int) *player.Player
	All() []*player.Player
	Delete(id int) bool
	Save() error
}

type GameStorage interface {
	NextID() int
	Add(g *game.Game) error
	Get(id int) *game.Game
	All() []*game.Game
	Delete(id int) bool
	Save() error
}

type MoveStorage interface {
	NextID() int
	Add(m *move.Move) error
	Get(id int) *move.Move
	All() []*move.Move
}

var (
	ErrEmptyName      = errors.New("не указано имя")
	ErrBadColor       = errors.New("цвет должен быть white или black")
	ErrBadAction      = errors.New("action может быть только surrender")
	ErrBadSquare      = errors.New("клетки нужно писать так: E2, E4")
	ErrPlayerNotFound = errors.New("игрок не найден")
	ErrGameNotFound   = errors.New("игра не найдена")
	ErrMoveNotFound   = errors.New("ход не найден")
	ErrSave           = errors.New("ошибка сохранения")
)

type Service struct {
	players PlayerStorage
	games   GameStorage
	moves   MoveStorage
}

func New(players PlayerStorage, games GameStorage, moves MoveStorage) *Service {
	return &Service{players: players, games: games, moves: moves}
}

// игроки

func (s *Service) CreatePlayer(name string, color string) (*player.Player, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}
	if color != player.White && color != player.Black {
		return nil, ErrBadColor
	}

	p := player.NewPlayer(s.players.NextID(), name, color)
	err := s.players.Add(p)
	if err != nil {
		return nil, ErrSave
	}
	return p, nil
}

func (s *Service) GetPlayer(id int) (*player.Player, error) {
	p := s.players.Get(id)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	return p, nil
}

func (s *Service) ListPlayers() []*player.Player {
	return s.players.All()
}

// меняем только имя
func (s *Service) RenamePlayer(id int, name string) (*player.Player, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, ErrEmptyName
	}
	p := s.players.Get(id)
	if p == nil {
		return nil, ErrPlayerNotFound
	}
	p.Name = name
	err := s.players.Save()
	if err != nil {
		return nil, ErrSave
	}
	return p, nil
}

func (s *Service) DeletePlayer(id int) error {
	if !s.players.Delete(id) {
		return ErrPlayerNotFound
	}
	err := s.players.Save()
	if err != nil {
		return ErrSave
	}
	return nil
}

// игры

func (s *Service) CreateGame(whiteID int, blackID int) (*game.Game, error) {
	white := s.players.Get(whiteID)
	black := s.players.Get(blackID)
	if white == nil || black == nil {
		return nil, ErrPlayerNotFound
	}

	g := game.NewGame(s.games.NextID(), white, black)
	err := s.games.Add(g)
	if err != nil {
		return nil, ErrSave
	}
	return g, nil
}

func (s *Service) GetGame(id int) (*game.Game, error) {
	g := s.games.Get(id)
	if g == nil {
		return nil, ErrGameNotFound
	}
	return g, nil
}

func (s *Service) ListGames() []*game.Game {
	return s.games.All()
}

// игру можно только сдать
func (s *Service) UpdateGame(id int, action string) (*game.Game, error) {
	if action != "surrender" {
		return nil, ErrBadAction
	}
	g := s.games.Get(id)
	if g == nil {
		return nil, ErrGameNotFound
	}
	err := g.Surrender()
	if err != nil {
		return nil, err
	}
	err = s.games.Save()
	if err != nil {
		return nil, ErrSave
	}
	return g, nil
}

func (s *Service) DeleteGame(id int) error {
	if !s.games.Delete(id) {
		return ErrGameNotFound
	}
	err := s.games.Save()
	if err != nil {
		return ErrSave
	}
	return nil
}

// ходы

func (s *Service) MakeMove(gameID int, from string, to string) (*move.Move, error) {
	fromRow, fromCol, ok1 := board.ParseSquare(strings.ToUpper(from))
	toRow, toCol, ok2 := board.ParseSquare(strings.ToUpper(to))
	if !ok1 || !ok2 {
		return nil, ErrBadSquare
	}

	g := s.games.Get(gameID)
	if g == nil {
		return nil, ErrGameNotFound
	}
	m, err := g.MakeMove(s.moves.NextID(), fromRow, fromCol, toRow, toCol)
	if err != nil {
		return nil, err
	}
	err = s.moves.Add(m)
	if err != nil {
		return nil, ErrSave
	}
	// игру тоже сохраняем
	err = s.games.Save()
	if err != nil {
		return nil, ErrSave
	}
	return m, nil
}

func (s *Service) GetMove(id int) (*move.Move, error) {
	m := s.moves.Get(id)
	if m == nil {
		return nil, ErrMoveNotFound
	}
	return m, nil
}

func (s *Service) ListMoves() []*move.Move {
	return s.moves.All()
}
