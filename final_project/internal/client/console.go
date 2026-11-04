package client

import (
	"errors"
	"fmt"
	"strconv"
	"strings"

	"tictactoe/internal/api"
	"tictactoe/internal/game"
)

// ErrBadInput - пользователь ввёл ход не в том формате.
var ErrBadInput = errors.New("введите номер строки и столбца через пробел, например: 1 3")

// RenderBoard рисует поле в виде текста. Строки и столбцы нумеруются с 1.
// Пример для поля 3x3:
//
//	    1   2   3
//	1  X |   | O
//	  ---+---+---
//	2    | X |
//
// Пустые клетки рисуются пробелами.
func RenderBoard(board [][]string) string {
	var sb strings.Builder

	sb.WriteString("  ")
	for i := range board {
		fmt.Fprintf(&sb, "  %d ", i+1)
	}
	sb.WriteString("\n")

	for i, row := range board {
		fmt.Fprintf(&sb, "%d ", i+1)
		for j, cell := range row {
			if cell == "" {
				cell = " "
			}
			fmt.Fprintf(&sb, " %s ", cell)
			if j < len(row)-1 {
				sb.WriteString("|")
			}
		}
		sb.WriteString("\n")

		if i < len(board)-1 {
			sb.WriteString("  ")
			sb.WriteString(strings.Repeat("---+", len(row)-1))
			sb.WriteString("---\n")
		}
	}
	return sb.String()
}

// ParseMove разбирает строку вида "2 3" и возвращает координаты, начиная с 0.
func ParseMove(input string, size int) (row, col int, err error) {
	parts := strings.Fields(input)
	if len(parts) != 2 {
		return 0, 0, ErrBadInput
	}

	row, err = strconv.Atoi(parts[0])
	if err != nil {
		return 0, 0, ErrBadInput
	}
	col, err = strconv.Atoi(parts[1])
	if err != nil {
		return 0, 0, ErrBadInput
	}

	if row < 1 || row > size || col < 1 || col > size {
		return 0, 0, fmt.Errorf("строка и столбец должны быть от 1 до %d", size)
	}
	return row - 1, col - 1, nil
}

// ResultText возвращает текст итога игры для игрока mySymbol или пустую строку, если игра не закончена.
func ResultText(state api.GameState, mySymbol string) string {
	switch game.Status(state.Status) {
	case game.StatusDraw:
		return "Ничья!"
	case game.StatusXWon:
		return winText(game.X, mySymbol)
	case game.StatusOWon:
		return winText(game.O, mySymbol)
	case game.StatusPlaying:
	}
	return ""
}

func winText(winner, mySymbol string) string {
	if winner == mySymbol {
		return "Вы победили!"
	}
	return "Вы проиграли. Победил " + winner + "."
}
