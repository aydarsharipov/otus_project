package game

import (
	"errors"
	"testing"
)

type move struct {
	player   string
	row, col int
}

// play делает несколько ходов подряд и падает, если какой-то ход не прошёл.
func play(t *testing.T, g *Game, moves []move) {
	t.Helper()
	for _, m := range moves {
		if err := g.Move(m.player, m.row, m.col); err != nil {
			t.Fatalf("ход %v вернул ошибку: %v", m, err)
		}
	}
}

func TestNew(t *testing.T) {
	for _, size := range []int{3, 4} {
		g, err := New(size)
		if err != nil {
			t.Fatalf("size %d: неожиданная ошибка %v", size, err)
		}
		if g.Size() != size {
			t.Errorf("размер %d, ожидали %d", g.Size(), size)
		}
		if g.Turn() != X {
			t.Errorf("первым должен ходить X, а ходит %s", g.Turn())
		}
		if g.Status() != StatusPlaying {
			t.Errorf("статус %s, ожидали %s", g.Status(), StatusPlaying)
		}
		for _, row := range g.Board() {
			for _, cell := range row {
				if cell != Empty {
					t.Fatalf("новое поле должно быть пустым")
				}
			}
		}
	}
}

func TestNewBadSize(t *testing.T) {
	for _, size := range []int{-1, 0, 1, 2, 5, 10} {
		if _, err := New(size); !errors.Is(err, ErrBadSize) {
			t.Errorf("size %d: ожидали ErrBadSize, получили %v", size, err)
		}
	}
}

func TestMoveErrors(t *testing.T) {
	tests := []struct {
		name    string
		before  []move
		move    move
		wantErr error
	}{
		{name: "чужой игрок", move: move{"Z", 0, 0}, wantErr: ErrBadPlayer},
		{name: "первым ходит O", move: move{O, 0, 0}, wantErr: ErrNotYourTurn},
		{name: "X ходит два раза", before: []move{{X, 0, 0}}, move: move{X, 1, 1}, wantErr: ErrNotYourTurn},
		{name: "строка меньше нуля", move: move{X, -1, 0}, wantErr: ErrOutOfBoard},
		{name: "столбец меньше нуля", move: move{X, 0, -1}, wantErr: ErrOutOfBoard},
		{name: "строка за полем", move: move{X, 3, 0}, wantErr: ErrOutOfBoard},
		{name: "столбец за полем", move: move{X, 0, 3}, wantErr: ErrOutOfBoard},
		{name: "занятая клетка", before: []move{{X, 1, 1}}, move: move{O, 1, 1}, wantErr: ErrCellBusy},
		{
			name:    "ход после победы",
			before:  []move{{X, 0, 0}, {O, 1, 0}, {X, 0, 1}, {O, 1, 1}, {X, 0, 2}},
			move:    move{O, 2, 2},
			wantErr: ErrGameOver,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(3)
			if err != nil {
				t.Fatal(err)
			}
			play(t, g, tt.before)

			err = g.Move(tt.move.player, tt.move.row, tt.move.col)
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("ожидали %v, получили %v", tt.wantErr, err)
			}
		})
	}
}

func TestMoveSwitchesTurn(t *testing.T) {
	g, err := New(3)
	if err != nil {
		t.Fatal(err)
	}

	play(t, g, []move{{X, 0, 0}})
	if g.Turn() != O {
		t.Errorf("после X должен ходить O")
	}
	play(t, g, []move{{O, 1, 1}})
	if g.Turn() != X {
		t.Errorf("после O должен ходить X")
	}

	board := g.Board()
	if board[0][0] != X || board[1][1] != O {
		t.Errorf("ходы не попали на поле: %v", board)
	}
}

func TestBoardIsCopy(t *testing.T) {
	g, err := New(3)
	if err != nil {
		t.Fatal(err)
	}

	board := g.Board()
	board[0][0] = O

	if g.Board()[0][0] != Empty {
		t.Errorf("изменение копии поменяло саму игру")
	}
}

func TestWinner(t *testing.T) {
	tests := []struct {
		name  string
		size  int
		moves []move
		want  Status
	}{
		{
			name:  "3x3 X по строке",
			size:  3,
			moves: []move{{X, 1, 0}, {O, 0, 0}, {X, 1, 1}, {O, 0, 1}, {X, 1, 2}},
			want:  StatusXWon,
		},
		{
			name:  "3x3 O по столбцу",
			size:  3,
			moves: []move{{X, 0, 0}, {O, 0, 2}, {X, 1, 0}, {O, 1, 2}, {X, 2, 1}, {O, 2, 2}},
			want:  StatusOWon,
		},
		{
			name:  "3x3 X по главной диагонали",
			size:  3,
			moves: []move{{X, 0, 0}, {O, 0, 1}, {X, 1, 1}, {O, 0, 2}, {X, 2, 2}},
			want:  StatusXWon,
		},
		{
			name:  "3x3 X по побочной диагонали",
			size:  3,
			moves: []move{{X, 0, 2}, {O, 0, 0}, {X, 1, 1}, {O, 0, 1}, {X, 2, 0}},
			want:  StatusXWon,
		},
		{
			name: "3x3 ничья",
			size: 3,
			// X O X
			// X O O
			// O X X
			moves: []move{
				{X, 0, 0}, {O, 0, 1}, {X, 0, 2},
				{O, 1, 1}, {X, 1, 0}, {O, 1, 2},
				{X, 2, 1}, {O, 2, 0}, {X, 2, 2},
			},
			want: StatusDraw,
		},
		{
			name:  "3x3 игра ещё идёт",
			size:  3,
			moves: []move{{X, 0, 0}, {O, 1, 1}},
			want:  StatusPlaying,
		},
		{
			name: "4x4 X по строке",
			size: 4,
			moves: []move{
				{X, 3, 0}, {O, 0, 0}, {X, 3, 1}, {O, 0, 1},
				{X, 3, 2}, {O, 0, 2}, {X, 3, 3},
			},
			want: StatusXWon,
		},
		{
			name: "4x4 три в ряд ещё не победа",
			size: 4,
			moves: []move{
				{X, 0, 0}, {O, 1, 0}, {X, 0, 1}, {O, 1, 1}, {X, 0, 2},
			},
			want: StatusPlaying,
		},
		{
			name: "4x4 O по диагонали",
			size: 4,
			moves: []move{
				{X, 0, 1}, {O, 0, 0}, {X, 0, 2}, {O, 1, 1},
				{X, 0, 3}, {O, 2, 2}, {X, 1, 0}, {O, 3, 3},
			},
			want: StatusOWon,
		},
		{
			name: "4x4 ничья",
			size: 4,
			// X X O O
			// O O X X
			// X X O O
			// O O X X
			moves: []move{
				{X, 0, 0}, {O, 0, 2}, {X, 0, 1}, {O, 0, 3},
				{X, 1, 2}, {O, 1, 0}, {X, 1, 3}, {O, 1, 1},
				{X, 2, 0}, {O, 2, 2}, {X, 2, 1}, {O, 2, 3},
				{X, 3, 2}, {O, 3, 0}, {X, 3, 3}, {O, 3, 1},
			},
			want: StatusDraw,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			g, err := New(tt.size)
			if err != nil {
				t.Fatal(err)
			}
			play(t, g, tt.moves)

			if g.Status() != tt.want {
				t.Errorf("статус %s, ожидали %s", g.Status(), tt.want)
			}
		})
	}
}
