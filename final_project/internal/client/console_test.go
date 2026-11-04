package client

import (
	"errors"
	"testing"

	"tictactoe/internal/api"
)

func TestRenderBoard(t *testing.T) {
	board := [][]string{
		{"X", "", "O"},
		{"", "X", ""},
		{"O", "", ""},
	}
	want := "" +
		"    1   2   3 \n" +
		"1  X |   | O \n" +
		"  ---+---+---\n" +
		"2    | X |   \n" +
		"  ---+---+---\n" +
		"3  O |   |   \n"

	if got := RenderBoard(board); got != want {
		t.Errorf("получили:\n%s\nожидали:\n%s", got, want)
	}
}

func TestRenderBoard4x4(t *testing.T) {
	const line = "  ---+---+---+---\n"
	board := [][]string{
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", ""},
		{"", "", "", "X"},
	}
	want := "" +
		"    1   2   3   4 \n" +
		"1    |   |   |   \n" + line +
		"2    |   |   |   \n" + line +
		"3    |   |   |   \n" + line +
		"4    |   |   | X \n"

	if got := RenderBoard(board); got != want {
		t.Errorf("получили:\n%s\nожидали:\n%s", got, want)
	}
}

func TestParseMove(t *testing.T) {
	tests := []struct {
		input   string
		size    int
		row     int
		col     int
		wantErr bool
	}{
		{input: "1 1", size: 3, row: 0, col: 0},
		{input: "3 2", size: 3, row: 2, col: 1},
		{input: "  2   3  ", size: 3, row: 1, col: 2},
		{input: "4 4", size: 4, row: 3, col: 3},
		{input: "4 4", size: 3, wantErr: true},
		{input: "0 1", size: 3, wantErr: true},
		{input: "-1 1", size: 3, wantErr: true},
		{input: "", size: 3, wantErr: true},
		{input: "1", size: 3, wantErr: true},
		{input: "1 2 3", size: 3, wantErr: true},
		{input: "a b", size: 3, wantErr: true},
		{input: "1 b", size: 3, wantErr: true},
	}

	for _, tt := range tests {
		row, col, err := ParseMove(tt.input, tt.size)
		if tt.wantErr {
			if err == nil {
				t.Errorf("%q: ожидали ошибку", tt.input)
			}
			continue
		}
		if err != nil {
			t.Errorf("%q: неожиданная ошибка %v", tt.input, err)
			continue
		}
		if row != tt.row || col != tt.col {
			t.Errorf("%q: получили (%d,%d), ожидали (%d,%d)", tt.input, row, col, tt.row, tt.col)
		}
	}
}

func TestParseMoveBadFormat(t *testing.T) {
	if _, _, err := ParseMove("abc", 3); !errors.Is(err, ErrBadInput) {
		t.Errorf("ожидали ErrBadInput, получили %v", err)
	}
}

func TestResultText(t *testing.T) {
	const win = "Вы победили!"

	tests := []struct {
		status string
		me     string
		want   string
	}{
		{"playing", "X", ""},
		{"waiting", "X", ""},
		{"draw", "O", "Ничья!"},
		{"x_won", "X", win},
		{"x_won", "O", "Вы проиграли. Победил X."},
		{"o_won", "O", win},
		{"o_won", "X", "Вы проиграли. Победил O."},
	}

	for _, tt := range tests {
		got := ResultText(api.GameState{Status: tt.status}, tt.me)
		if got != tt.want {
			t.Errorf("%s/%s: получили %q, ожидали %q", tt.status, tt.me, got, tt.want)
		}
	}
}
