package game

import (
	"errors"
	"time"
	"HW_6/internal/model/board"
	"HW_6/internal/model/move"
	"HW_6/internal/model/piece"
	"HW_6/internal/model/player"
)

const BoardSize = 8

type GameStatus string

const (
	Active    GameStatus = "active"
	Completed GameStatus = "completed"
)

type Game struct {
	id          int
	whitePlayer *player.Player
	blackPlayer *player.Player
	board       *board.Board
	moves       []*move.Move
	currentTurn *player.Player
	status      GameStatus
	createdAt   time.Time
	winner      *player.Player
	endReason   string
}

func NewGame(id int, whitePlayer, blackPlayer *player.Player) *Game {
	g := &Game{
		id:          id,
		whitePlayer: whitePlayer,
		blackPlayer: blackPlayer,
		board:       board.NewBoard(BoardSize),
		moves:       make([]*move.Move, 0),
		currentTurn: whitePlayer,
		status:      Active,
		createdAt:   time.Now(),
	}
	g.setupInitialPosition()
	return g
}

func (g *Game) setupInitialPosition() {
	backRank := []piece.PieceType{
		piece.Rook, piece.Knight, piece.Bishop, piece.Queen,
		piece.King, piece.Bishop, piece.Knight, piece.Rook,
	}
	for col, pieceType := range backRank {
		g.board.SetPiece(0, col, piece.NewPiece(pieceType, "black"))
		g.board.SetPiece(1, col, piece.NewPiece(piece.Pawn, "black"))
		g.board.SetPiece(6, col, piece.NewPiece(piece.Pawn, "white"))
		g.board.SetPiece(7, col, piece.NewPiece(pieceType, "white"))
	}
}

func (g *Game) EntityName() string {
	return "game"
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

func (g *Game) Winner() *player.Player {
	return g.winner
}

func (g *Game) EndReason() string {
	return g.endReason
}

// проверяет ход текущего игрока
func (g *Game) CanMove(fromRow, fromCol, toRow, toCol int) error {
	from := g.board.GetCell(fromRow, fromCol)
	to := g.board.GetCell(toRow, toCol)
	myColor := string(g.currentTurn.Color())

	if from == nil || to == nil {
		return errors.New("клетка вне доски")
	}
	if from.IsEmpty() {
		return errors.New("в начальной клетке нет фигуры")
	}
	if from.Piece().Color() != myColor {
		return errors.New("это фигура соперника")
	}
	if !to.IsEmpty() && to.Piece().Color() == myColor {
		return errors.New("в конечной клетке стоит ваша фигура")
	}
	return nil
}

// переставляет фигуру на доске и добавляет ход в слайс ходов
func (g *Game) MakeMove(fromRow, fromCol, toRow, toCol int) (*move.Move, error) {
	if g.status != Active {
		return nil, errors.New("игра уже закончена")
	}
	if err := g.CanMove(fromRow, fromCol, toRow, toCol); err != nil {
		return nil, err
	}

	p := g.board.GetCell(fromRow, fromCol).Piece()
	captured := g.board.GetCell(toRow, toCol).Piece()

	g.board.SetPiece(toRow, toCol, p)
	g.board.SetPiece(fromRow, fromCol, nil)

	m := move.NewMove(len(g.moves)+1, fromRow, fromCol, toRow, toCol, p)
	m.Complete()
	g.moves = append(g.moves, m)

	if captured != nil && captured.Type() == piece.King {
		g.finish(g.currentTurn, "король взят")
		return m, nil
	}

	// Передаём ход сопернику
	if g.currentTurn == g.whitePlayer {
		g.currentTurn = g.blackPlayer
	} else {
		g.currentTurn = g.whitePlayer
	}
	return m, nil
}

// текущий игрок сдаётся, победа сопернику
func (g *Game) Surrender() {
	loser := g.currentTurn
	winner := g.whitePlayer
	if loser == g.whitePlayer {
		winner = g.blackPlayer
	}
	g.finish(winner, loser.Name()+" сдаётся")
}

func (g *Game) finish(winner *player.Player, reason string) {
	g.status = Completed
	g.winner = winner
	g.endReason = reason
}
