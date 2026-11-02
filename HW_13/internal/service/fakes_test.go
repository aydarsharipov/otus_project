package service_test

import (
	"errors"

	"HW_13/internal/model/game"
	"HW_13/internal/model/move"
	"HW_13/internal/model/player"
	"HW_13/internal/service"
)

// ошибка которую возвращают фейки если надо сломать сохранение
var errDisk = errors.New("диск сломался")

// фейковые хранилища в памяти, вместо файлов

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

func (f *fakePlayers) Delete(id int) bool {
	for i, p := range f.items {
		if p.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return true
		}
	}
	return false
}

func (f *fakePlayers) Save() error {
	return f.saveErr
}

type fakeGames struct {
	items   []*game.Game
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

func (f *fakeGames) Delete(id int) bool {
	for i, g := range f.items {
		if g.ID == id {
			f.items = append(f.items[:i], f.items[i+1:]...)
			return true
		}
	}
	return false
}

func (f *fakeGames) Save() error {
	return f.saveErr
}

type fakeMoves struct {
	items  []*move.Move
	addErr error
}

func (f *fakeMoves) NextID() int {
	return len(f.items) + 1
}

func (f *fakeMoves) Add(m *move.Move) error {
	if f.addErr != nil {
		return f.addErr
	}
	f.items = append(f.items, m)
	return nil
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
		games:   &fakeGames{items: []*game.Game{game.NewGame(1, ivan, petr)}},
		moves:   &fakeMoves{},
	}
	e.svc = service.New(e.players, e.games, e.moves)
	return e
}
