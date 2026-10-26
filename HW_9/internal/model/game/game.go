package game

import (
	"encoding/json"
	"errors"
	"sync"
	"time"

	"HW_9/internal/model/board"
	"HW_9/internal/model/move"
	"HW_9/internal/model/piece"
	"HW_9/internal/model/player"
)

const BoardSize = board.StandardSize

type GameStatus string

const (
	Active    GameStatus = "active"
	Completed GameStatus = "completed"
)

type PlayerTime struct {
	Last  time.Duration
	Total time.Duration
}

type Game struct {
	mu            sync.RWMutex
	id            int
	whitePlayer   *player.Player
	blackPlayer   *player.Player
	board         *board.Board
	moves         []*move.Move
	currentTurn   *player.Player
	status        GameStatus
	createdAt     time.Time
	winner        *player.Player
	endReason     string
	turnStartedAt time.Time
	times         map[player.Color]*PlayerTime
}

func NewGame(id int, whitePlayer, blackPlayer *player.Player) *Game {
	now := time.Now()
	g := &Game{
		id:            id,
		whitePlayer:   whitePlayer,
		blackPlayer:   blackPlayer,
		board:         board.NewBoard(BoardSize),
		moves:         make([]*move.Move, 0),
		currentTurn:   whitePlayer,
		status:        Active,
		createdAt:     now,
		turnStartedAt: now,
		times: map[player.Color]*PlayerTime{
			player.White: {},
			player.Black: {},
		},
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

func (g *Game) CurrentTurn() *player.Player {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.currentTurn
}

func (g *Game) Status() GameStatus {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.status
}

func (g *Game) MovesCount() int {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return len(g.moves)
}

func (g *Game) CanMove(fromRow, fromCol, toRow, toCol int) error {
	g.mu.RLock()
	defer g.mu.RUnlock()
	return g.canMove(fromRow, fromCol, toRow, toCol)
}

func (g *Game) canMove(fromRow, fromCol, toRow, toCol int) error {
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

func (g *Game) PossibleMoves() [][4]int {
	g.mu.RLock()
	defer g.mu.RUnlock()

	possible := make([][4]int, 0)
	for fromRow := 0; fromRow < BoardSize; fromRow++ {
		for fromCol := 0; fromCol < BoardSize; fromCol++ {
			for toRow := 0; toRow < BoardSize; toRow++ {
				for toCol := 0; toCol < BoardSize; toCol++ {
					if g.canMove(fromRow, fromCol, toRow, toCol) == nil {
						possible = append(possible, [4]int{fromRow, fromCol, toRow, toCol})
					}
				}
			}
		}
	}
	return possible
}

func (g *Game) MakeMove(fromRow, fromCol, toRow, toCol int) (*move.Move, error) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.status != Active {
		return nil, errors.New("игра уже закончена")
	}
	if err := g.canMove(fromRow, fromCol, toRow, toCol); err != nil {
		return nil, err
	}

	p := g.board.GetCell(fromRow, fromCol).Piece()

	g.board.SetPiece(toRow, toCol, p)
	g.board.SetPiece(fromRow, fromCol, nil)

	now := time.Now()
	spent := now.Sub(g.turnStartedAt)
	t := g.times[g.currentTurn.Color()]
	t.Last = spent
	t.Total += spent
	g.turnStartedAt = now

	m := move.NewMove(len(g.moves)+1, fromRow, fromCol, toRow, toCol, p)
	m.SetDuration(spent)
	m.Complete()
	g.moves = append(g.moves, m)

	if g.currentTurn == g.whitePlayer {
		g.currentTurn = g.blackPlayer
	} else {
		g.currentTurn = g.whitePlayer
	}
	return m, nil
}

func (g *Game) Surrender() {
	g.mu.Lock()
	defer g.mu.Unlock()

	loser := g.currentTurn
	winner := g.whitePlayer
	if loser == g.whitePlayer {
		winner = g.blackPlayer
	}
	g.finish(winner, loser.Name()+" сдаётся")
}

func (g *Game) Draw(reason string) {
	g.mu.Lock()
	defer g.mu.Unlock()

	if g.status == Active {
		g.finish(nil, reason)
	}
}

func (g *Game) finish(winner *player.Player, reason string) {
	g.status = Completed
	g.winner = winner
	g.endReason = reason
}

type Snapshot struct {
	ID         int
	WhiteName  string
	BlackName  string
	Cells      [BoardSize][BoardSize]rune
	MovesCount int
	LastMove   string
	Status     GameStatus
	TurnName   string
	TurnColor  player.Color
	Thinking   time.Duration
	WinnerName string
	EndReason  string
	WhiteTime  PlayerTime
	BlackTime  PlayerTime
}

func (g *Game) Snapshot() Snapshot {
	g.mu.RLock()
	defer g.mu.RUnlock()

	s := Snapshot{
		ID:         g.id,
		WhiteName:  g.whitePlayer.Name(),
		BlackName:  g.blackPlayer.Name(),
		MovesCount: len(g.moves),
		Status:     g.status,
		TurnName:   g.currentTurn.Name(),
		TurnColor:  g.currentTurn.Color(),
		Thinking:   time.Since(g.turnStartedAt),
		EndReason:  g.endReason,
		WhiteTime:  *g.times[player.White],
		BlackTime:  *g.times[player.Black],
	}
	for row := 0; row < BoardSize; row++ {
		for col := 0; col < BoardSize; col++ {
			if p := g.board.GetCell(row, col).Piece(); p != nil {
				s.Cells[row][col] = p.Symbol()
			}
		}
	}
	if len(g.moves) > 0 {
		s.LastMove = g.moves[len(g.moves)-1].Notation()
	}
	if g.winner != nil {
		s.WinnerName = g.winner.Name()
	}
	return s
}

type gameJSON struct {
	ID          int                          `json:"id"`
	White       *player.Player               `json:"white"`
	Black       *player.Player               `json:"black"`
	Board       *board.Board                 `json:"board"`
	Moves       []*move.Move                 `json:"moves"`
	CurrentTurn player.Color                 `json:"current_turn"`
	Status      GameStatus                   `json:"status"`
	CreatedAt   time.Time                    `json:"created_at"`
	Winner      player.Color                 `json:"winner,omitempty"`
	EndReason   string                       `json:"end_reason,omitempty"`
	Times       map[player.Color]*PlayerTime `json:"times"`
}

func (g *Game) MarshalJSON() ([]byte, error) {
	g.mu.RLock()
	defer g.mu.RUnlock()

	v := gameJSON{
		ID: g.id, White: g.whitePlayer, Black: g.blackPlayer, Board: g.board,
		Moves: g.moves, CurrentTurn: g.currentTurn.Color(), Status: g.status,
		CreatedAt: g.createdAt, EndReason: g.endReason, Times: g.times,
	}
	if g.winner != nil {
		v.Winner = g.winner.Color()
	}
	return json.Marshal(v)
}

func (g *Game) UnmarshalJSON(data []byte) error {
	var v gameJSON
	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	g.id = v.ID
	g.whitePlayer = v.White
	g.blackPlayer = v.Black
	g.board = v.Board
	g.moves = v.Moves
	g.status = v.Status
	g.createdAt = v.CreatedAt
	g.endReason = v.EndReason
	g.turnStartedAt = time.Now()
	g.times = v.Times

	if v.CurrentTurn == player.Black {
		g.currentTurn = g.blackPlayer
	} else {
		g.currentTurn = g.whitePlayer
	}
	if v.Winner == player.White {
		g.winner = g.whitePlayer
	}
	if v.Winner == player.Black {
		g.winner = g.blackPlayer
	}
	return nil
}
