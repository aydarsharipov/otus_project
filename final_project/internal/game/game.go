// Package game содержит правила игры "Крестики-нолики".
package game

import "errors"

// Символы игроков и пустой клетки.
const (
	Empty = ""
	X     = "X"
	O     = "O"
)

// Status - состояние партии.
type Status string

// Возможные состояния партии.
const (
	StatusPlaying Status = "playing"
	StatusXWon    Status = "x_won"
	StatusOWon    Status = "o_won"
	StatusDraw    Status = "draw"
)

// Ошибки, которые возвращает игра.
var (
	ErrBadSize     = errors.New("размер поля должен быть 3 или 4")
	ErrBadPlayer   = errors.New("неизвестный игрок")
	ErrOutOfBoard  = errors.New("клетка за пределами поля")
	ErrCellBusy    = errors.New("клетка уже занята")
	ErrNotYourTurn = errors.New("сейчас ход другого игрока")
	ErrGameOver    = errors.New("игра уже закончена")
)

// Game - одна партия крестиков-ноликов.
type Game struct {
	size   int
	board  [][]string
	turn   string
	status Status
	moves  int
}

// New создаёт новую игру. Размер поля может быть 3 или 4. Первым ходит X.
func New(size int) (*Game, error) {
	if size != 3 && size != 4 {
		return nil, ErrBadSize
	}

	board := make([][]string, size)
	for i := range board {
		board[i] = make([]string, size)
	}

	return &Game{
		size:   size,
		board:  board,
		turn:   X,
		status: StatusPlaying,
	}, nil
}

// Size возвращает размер поля.
func (g *Game) Size() int {
	return g.size
}

// Turn возвращает символ игрока, который сейчас ходит.
func (g *Game) Turn() string {
	return g.turn
}

// Status возвращает текущее состояние партии.
func (g *Game) Status() Status {
	return g.status
}

// Board возвращает копию поля, чтобы снаружи нельзя было его поменять.
func (g *Game) Board() [][]string {
	result := make([][]string, g.size)
	for i := range g.board {
		result[i] = make([]string, g.size)
		copy(result[i], g.board[i])
	}
	return result
}

// Move делает ход игрока player в клетку (row, col). Координаты начинаются с 0.
func (g *Game) Move(player string, row, col int) error {
	if player != X && player != O {
		return ErrBadPlayer
	}
	if g.status != StatusPlaying {
		return ErrGameOver
	}
	if player != g.turn {
		return ErrNotYourTurn
	}
	if row < 0 || row >= g.size || col < 0 || col >= g.size {
		return ErrOutOfBoard
	}
	if g.board[row][col] != Empty {
		return ErrCellBusy
	}

	g.board[row][col] = player
	g.moves++

	if g.isWinner(player) {
		if player == X {
			g.status = StatusXWon
		} else {
			g.status = StatusOWon
		}
		return nil
	}

	if g.moves == g.size*g.size {
		g.status = StatusDraw
		return nil
	}

	if g.turn == X {
		g.turn = O
	} else {
		g.turn = X
	}
	return nil
}

// isWinner проверяет, заполнил ли игрок целую строку, столбец или диагональ.
func (g *Game) isWinner(player string) bool {
	diag1 := true
	diag2 := true

	for i := 0; i < g.size; i++ {
		rowFull := true
		colFull := true
		for j := 0; j < g.size; j++ {
			if g.board[i][j] != player {
				rowFull = false
			}
			if g.board[j][i] != player {
				colFull = false
			}
		}
		if rowFull || colFull {
			return true
		}

		if g.board[i][i] != player {
			diag1 = false
		}
		if g.board[i][g.size-1-i] != player {
			diag2 = false
		}
	}

	return diag1 || diag2
}
