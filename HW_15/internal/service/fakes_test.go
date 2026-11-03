package service_test

import (
	"errors"

	"HW_15/internal/model/game"
	"HW_15/internal/model/move"
	"HW_15/internal/model/player"
	"HW_15/internal/service"
)

// ошибка которую возвращают фейки если надо сломать сохранение
var errDB = errors.New("база сломалась")

// фейковые хранилища в памяти, вместо базы

type fakePlayers struct {
	items   []*player.Player
	saveErr error
}

func (f *fakePlayers) NextID() int {
	return len(f.items) + 1
}

func (f *fakePlayers) Add(p *player.Player) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.items = append(f.items, p)
	return nil
}

func (f *fakePlayers) Get(id int) *player.Player {
	for _, p := range f.items {
		if p.ID == id {
			return p
		}
	}
	return nil
}

func (f *fakePlayers) All() []*player.Player {
	return f.items
}

func (f *fakePlayers) Update(p *player.Player) error {
	return f.saveErr
}

func (f *fakePlayers) Delete(id int) (bool, error) {
	if f.saveErr != nil {
		return false, f.saveErr
	}
	for i, p := range f.items {
		if p.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

type fakeGames struct {
	items   []*game.Game
	moves   *fakeMoves // сюда AddMove складывает ходы
	saveErr error
}

func (f *fakeGames) NextID() int {
	return len(f.items) + 1
}

func (f *fakeGames) Add(g *game.Game) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.items = append(f.items, g)
	return nil
}

func (f *fakeGames) Get(id int) *game.Game {
	for _, g := range f.items {
		if g.ID == id {
			return g
		}
	}
	return nil
}

func (f *fakeGames) All() []*game.Game {
	return f.items
}

func (f *fakeGames) Update(g *game.Game) error {
	return f.saveErr
}

func (f *fakeGames) Delete(id int) (bool, error) {
	if f.saveErr != nil {
		return false, f.saveErr
	}
	for i, g := range f.items {
		if g.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return true, nil
		}
	}
	return false, nil
}

// как транзакция: при ошибке ход не сохраняется
func (f *fakeGames) AddMove(g *game.Game, m *move.Move) error {
	if f.saveErr != nil {
		return f.saveErr
	}
	f.moves.items = append(f.moves.items, m)
	return nil
}

type fakeMoves struct {
	items []*move.Move
}

func (f *fakeMoves) NextID() int {
	return len(f.items) + 1
}

func (f *fakeMoves) Get(id int) *move.Move {
	for _, m := range f.items {
		if m.ID == id {
			return m
		}
	}
	return nil
}

func (f *fakeMoves) All() []*move.Move {
	return f.items
}

// всё что нужно для теста
type testEnv struct {
	svc     *service.Service
	players *fakePlayers
	games   *fakeGames
	moves   *fakeMoves
}

// два игрока и одна игра между ними
func newEnv() *testEnv {
	ivan := player.NewPlayer(1, "Иван", player.White)
	petr := player.NewPlayer(2, "Пётр", player.Black)

	e := &testEnv{
		players: &fakePlayers{items: []*player.Player{ivan, petr}},
		moves:   &fakeMoves{},
	}
	e.games = &fakeGames{items: []*game.Game{game.NewGame(1, ivan, petr)}, moves: e.moves}
	e.svc = service.New(e.players, e.games, e.moves)
	return e
}
