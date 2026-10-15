package service

import (
	"fmt"
	"math/rand"

	"HW_6/internal/model"
	"HW_6/internal/model/board"
	"HW_6/internal/model/game"
	"HW_6/internal/model/move"
	"HW_6/internal/model/piece"
	"HW_6/internal/model/player"
	"HW_6/internal/repository"
)

var nextID = 1

// создаёт партию, сохраняет игроков и игру в репозиторий
func NewGame(whiteName, blackName string) *game.Game {
	id := nextID
	nextID++

	white := player.NewPlayer(id*2-1, whiteName, player.White)
	black := player.NewPlayer(id*2, blackName, player.Black)
	g := game.NewGame(id, white, black)

	for _, e := range []model.Entity{white, black, g} {
		_ = repository.Save(e)
	}
	return g
}

func GenerateEntities() error {
	id := nextID
	nextID++

	white := player.NewPlayer(id*2-1, fmt.Sprintf("bot%d", id*2-1), player.White)
	black := player.NewPlayer(id*2, fmt.Sprintf("bot%d", id*2), player.Black)
	p := piece.NewPiece(piece.AllTypes[rand.Intn(len(piece.AllTypes))], randomColor())

	entities := []model.Entity{
		white,
		black,
		p,
		move.NewMove(id, rand.Intn(game.BoardSize), rand.Intn(game.BoardSize),
			rand.Intn(game.BoardSize), rand.Intn(game.BoardSize), p),
		board.NewBoard(game.BoardSize),
		game.NewGame(id, white, black),
	}

	for _, e := range entities {
		if err := repository.Save(e); err != nil {
			return err
		}
	}
	return nil
}

func randomColor() string {
	if rand.Intn(2) == 0 {
		return string(player.White)
	}
	return string(player.Black)
}
