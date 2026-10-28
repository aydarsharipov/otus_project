package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HW_13/internal/model/game"
	"HW_13/internal/model/player"
	"HW_13/internal/service"
)

func TestMakeMove(t *testing.T) {
	tests := []struct {
		name         string
		gameID       int
		from         string
		to           string
		prepare      func(e *testEnv)
		wantErr      error
		wantNotation string
		wantMoves    int // сколько ходов в хранилище после вызова
	}{
		{name: "пешка E2-E4", gameID: 1, from: "E2", to: "E4", wantNotation: "♟ E2-E4", wantMoves: 1},
		{name: "маленькие буквы", gameID: 1, from: "e2", to: "e4", wantNotation: "♟ E2-E4", wantMoves: 1},
		{name: "конь G1-F3", gameID: 1, from: "G1", to: "F3", wantNotation: "♞ G1-F3", wantMoves: 1},
		{
			name: "ход чёрных после белых", gameID: 1, from: "E7", to: "E5",
			prepare:      func(e *testEnv) { mustMove(e, "E2", "E4") },
			wantNotation: "♙ E7-E5", wantMoves: 2,
		},
		{
			name: "взятие чужой фигуры", gameID: 1, from: "E4", to: "D5",
			prepare: func(e *testEnv) {
				mustMove(e, "E2", "E4")
				mustMove(e, "D7", "D5")
			},
			wantNotation: "♟ E4-D5", wantMoves: 3,
		},
		{name: "плохая клетка откуда", gameID: 1, from: "Z9", to: "E4", wantErr: service.ErrBadSquare},
		{name: "плохая клетка куда", gameID: 1, from: "E2", to: "E", wantErr: service.ErrBadSquare},
		{name: "ряд 0", gameID: 1, from: "E0", to: "E4", wantErr: service.ErrBadSquare},
		{name: "ряд 9", gameID: 1, from: "E2", to: "E9", wantErr: service.ErrBadSquare},
		{name: "пустые клетки", gameID: 1, from: "", to: "", wantErr: service.ErrBadSquare},
		{name: "нет такой игры", gameID: 100, from: "E2", to: "E4", wantErr: service.ErrGameNotFound},
		{name: "пустая клетка", gameID: 1, from: "E4", to: "E5", wantErr: game.ErrNoPiece},
		{name: "сейчас ход белых", gameID: 1, from: "E7", to: "E5", wantErr: game.ErrNotYourTurn},
		{name: "на свою фигуру", gameID: 1, from: "A1", to: "A2", wantErr: game.ErrOwnPiece},
		{
			name: "игра закончена", gameID: 1, from: "E2", to: "E4",
			prepare: func(e *testEnv) {
				_, err := e.svc.UpdateGame(1, "surrender")
				if err != nil {
					panic(err)
				}
			},
			wantErr: game.ErrGameFinished,
		},
		{
			name: "ошибка сохранения хода", gameID: 1, from: "E2", to: "E4",
			prepare: func(e *testEnv) { e.moves.addErr = errDisk },
			wantErr: service.ErrSave,
		},
		{
			name: "ошибка сохранения игры", gameID: 1, from: "E2", to: "E4",
			prepare: func(e *testEnv) { e.games.saveErr = errDisk },
			wantErr: service.ErrSave, wantMoves: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			m, err := e.svc.MakeMove(tt.gameID, tt.from, tt.to)

			assert.Len(t, e.svc.ListMoves(), tt.wantMoves)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, m)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNotation, m.Notation)
			assert.Equal(t, tt.gameID, m.GameID)
			assert.Equal(t, tt.wantMoves, m.ID)
		})
	}
}

// после хода фигура должна переехать, а очередь смениться
func TestMakeMoveChangesBoard(t *testing.T) {
	e := newEnv()

	_, err := e.svc.MakeMove(1, "E2", "E4")
	require.NoError(t, err)

	g, err := e.svc.GetGame(1)
	require.NoError(t, err)
	assert.Nil(t, g.Board.Cells[6][4], "на E2 должно быть пусто")
	require.NotNil(t, g.Board.Cells[4][4], "на E4 должна быть пешка")
	assert.Equal(t, "pawn", g.Board.Cells[4][4].Type)
	assert.Equal(t, player.Black, g.Turn)
	assert.Len(t, g.Moves, 1)
}

func TestGetMove(t *testing.T) {
	tests := []struct {
		name         string
		id           int
		wantNotation string
		wantErr      error
	}{
		{name: "первый ход", id: 1, wantNotation: "♟ E2-E4"},
		{name: "второй ход", id: 2, wantNotation: "♙ E7-E5"},
		{name: "нет такого", id: 100, wantErr: service.ErrMoveNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			mustMove(e, "E2", "E4")
			mustMove(e, "E7", "E5")

			m, err := e.svc.GetMove(tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, m)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantNotation, m.Notation)
		})
	}
}

func TestListMoves(t *testing.T) {
	e := newEnv()
	assert.Empty(t, e.svc.ListMoves())

	mustMove(e, "E2", "E4")

	assert.Len(t, e.svc.ListMoves(), 1)
}

// ход в игре 1, если не получилось то тест сломан
func mustMove(e *testEnv, from, to string) {
	_, err := e.svc.MakeMove(1, from, to)
	if err != nil {
		panic(err)
	}
}
