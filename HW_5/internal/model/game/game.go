package game

import (
	"time"

	"HW_5/internal/model/board"
	"HW_5/internal/model/move"
	"HW_5/internal/model/piece"
	"HW_5/internal/model/player"
)

type GameStatus string

const (
	Active    GameStatus = "active"
	Paused    GameStatus = "paused"
	Completed GameStatus = "completed"
)

type Game struct {
	id           int
	whitePlayer  *player.Player
	blackPlayer  *player.Player
	board        *board.Board
	moves        []*move.Move
	currentTurn  *player.Player
	status       GameStatus
	createdAt    time.Time
	updatedAt    time.Time
	winner       *player.Player
}

func NewGame(id int, whitePlayer, blackPlayer *player.Player) *Game {
	g := &Game{
		id:          id,
		whitePlayer: whitePlayer,
		blackPlayer: blackPlayer,
		board:       board.NewBoard(8),
		moves:       make([]*move.Move, 0),
		currentTurn: whitePlayer,
		status:      Active,
		createdAt:   time.Now(),
		updatedAt:   time.Now(),
	}
	g.setupInitialPosition()
	return g
}

func (g *Game) setupInitialPosition() {
	for col := 0; col < 8; col++ {
		pawn := piece.NewPiece("pawn", "black")
		g.board.SetPiece(1, col, pawn)
	}

	for col := 0; col < 8; col++ {
		pawn := piece.NewPiece("pawn", "white")
		g.board.SetPiece(6, col, pawn)
	}
}

func (g *Game) ID() int {
	return g.id
}

func (g *Game) WhitePlayer() *player.Player {
	return g.whitePlayer
}

func (g *Game) BlackPlayer() *player.Player {
	return g.blackPlayer
}

func (g *Game) Board() *board.Board {
	return g.board
}

func (g *Game) Moves() []*move.Move {
	return g.moves
}

func (g *Game) CurrentTurn() *player.Player {
	return g.currentTurn
}

func (g *Game) Status() GameStatus {
	return g.status
}

func (g *Game) CreatedAt() time.Time {
	return g.createdAt
}

func (g *Game) UpdatedAt() time.Time {
	return g.updatedAt
}

func (g *Game) Winner() *player.Player {
	return g.winner
}

func (g *Game) MakeMove(m *move.Move) bool {
	if g.status != Active {
		return false
	}

	if m.Status() == move.Invalid {
		return false
	}

	g.moves = append(g.moves, m)
	m.Complete()
	g.updatedAt = time.Now()

	if g.currentTurn == g.whitePlayer {
		g.currentTurn = g.blackPlayer
	} else {
		g.currentTurn = g.whitePlayer
	}

	return true
}

func (g *Game) Pause() {
	if g.status == Active {
		g.status = Paused
		g.updatedAt = time.Now()
	}
}

func (g *Game) Resume() {
	if g.status == Paused {
		g.status = Active
		g.updatedAt = time.Now()
	}
}

func (g *Game) EndGame(winner *player.Player) {
	g.status = Completed
	g.winner = winner
	g.updatedAt = time.Now()
}
