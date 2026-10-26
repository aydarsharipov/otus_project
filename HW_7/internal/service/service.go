package service

import (
	"context"
	"fmt"
	"math/rand"
	"sync/atomic"
	"time"

	"HW_7/internal/model"
	"HW_7/internal/model/board"
	"HW_7/internal/model/game"
	"HW_7/internal/model/move"
	"HW_7/internal/model/piece"
	"HW_7/internal/model/player"
)

type Service struct {
	out    chan<- model.Entity
	nextID atomic.Int64
}

func New(out chan<- model.Entity) *Service {
	return &Service{out: out}
}

func (s *Service) newID() int {
	return int(s.nextID.Add(1))
}

func (s *Service) send(entities ...model.Entity) {
	for _, e := range entities {
		s.out <- e
	}
}

func (s *Service) NewGame(whiteName, blackName string) *game.Game {
	id := s.newID()

	white := player.NewPlayer(id*2-1, whiteName, player.White)
	black := player.NewPlayer(id*2, blackName, player.Black)
	g := game.NewGame(id, white, black)

	s.send(white, black, g)
	return g
}

func (s *Service) GenerateEntities() {
	id := s.newID()

	white := player.NewPlayer(id*2-1, fmt.Sprintf("bot%d", id*2-1), player.White)
	black := player.NewPlayer(id*2, fmt.Sprintf("bot%d", id*2), player.Black)
	p := piece.NewPiece(piece.AllTypes[rand.Intn(len(piece.AllTypes))], randomColor())

	s.send(
		white,
		black,
		p,
		move.NewMove(id, rand.Intn(game.BoardSize), rand.Intn(game.BoardSize),
			rand.Intn(game.BoardSize), rand.Intn(game.BoardSize), p),
		board.NewBoard(game.BoardSize),
		game.NewGame(id, white, black),
	)
}

func (s *Service) RunGenerator(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.GenerateEntities()
		}
	}
}

func (s *Service) MakeMove(g *game.Game, fromRow, fromCol, toRow, toCol int) error {
	m, err := g.MakeMove(fromRow, fromCol, toRow, toCol)
	if err != nil {
		return err
	}
	s.send(m)
	return nil
}

func (s *Service) AutoMove(g *game.Game) {
	possible := g.PossibleMoves()
	if len(possible) == 0 {
		g.Draw("нет ходов")
		return
	}
	m := possible[rand.Intn(len(possible))]
	_ = s.MakeMove(g, m[0], m[1], m[2], m[3])
}

func (s *Service) RunSimulation(ctx context.Context, g *game.Game, maxMoves int, minDelay, maxDelay time.Duration) {
	for g.Status() == game.Active {
		if g.MovesCount() >= maxMoves {
			g.Draw("лимит ходов")
			return
		}
		if !Sleep(ctx, RandomDelay(minDelay, maxDelay)) {
			return
		}
		s.AutoMove(g)
	}
}

func RandomDelay(min, max time.Duration) time.Duration {
	return min + time.Duration(rand.Int63n(int64(max-min)+1))
}

func Sleep(ctx context.Context, d time.Duration) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

func randomColor() string {
	if rand.Intn(2) == 0 {
		return string(player.White)
	}
	return string(player.Black)
}
