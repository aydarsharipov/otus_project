package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HW_13/internal/model/game"
	"HW_13/internal/model/player"
	"HW_13/internal/service"
)

func TestCreateGame(t *testing.T) {
	tests := []struct {
		name      string
		whiteID   int
		blackID   int
		prepare   func(e *testEnv)
		wantErr   error
		wantCount int
	}{
		{name: "успешно", whiteID: 1, blackID: 2, wantCount: 2},
		{name: "нет белого", whiteID: 100, blackID: 2, wantErr: service.ErrPlayerNotFound, wantCount: 1},
		{name: "нет чёрного", whiteID: 1, blackID: 100, wantErr: service.ErrPlayerNotFound, wantCount: 1},
		{name: "нет обоих", whiteID: 50, blackID: 60, wantErr: service.ErrPlayerNotFound, wantCount: 1},
		{
			name: "ошибка сохранения", whiteID: 1, blackID: 2,
			prepare: func(e *testEnv) { e.games.saveErr = errDisk },
			wantErr: service.ErrSave, wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			g, err := e.svc.CreateGame(tt.whiteID, tt.blackID)

			assert.Len(t, e.svc.ListGames(), tt.wantCount)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, g)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 2, g.ID)
			assert.Equal(t, "Иван", g.White.Name)
			assert.Equal(t, "Пётр", g.Black.Name)
			assert.Equal(t, player.White, g.Turn)
			assert.Equal(t, "active", g.Status)
			assert.Empty(t, g.Moves)
		})
	}
}

func TestGetGame(t *testing.T) {
	tests := []struct {
		name    string
		id      int
		wantErr error
	}{
		{name: "есть такая", id: 1},
		{name: "нет такой", id: 100, wantErr: service.ErrGameNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()

			g, err := e.svc.GetGame(tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, g)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.id, g.ID)
		})
	}
}

func TestListGames(t *testing.T) {
	e := newEnv()

	list := e.svc.ListGames()

	require.Len(t, list, 1)
	assert.Equal(t, 1, list[0].ID)
}

func TestUpdateGame(t *testing.T) {
	tests := []struct {
		name       string
		id         int
		action     string
		prepare    func(e *testEnv)
		wantErr    error
		wantWinner string
	}{
		{name: "сдаются белые", id: 1, action: "surrender", wantWinner: player.Black},
		{
			name: "сдаются чёрные", id: 1, action: "surrender",
			prepare:    func(e *testEnv) { mustMove(e, "E2", "E4") },
			wantWinner: player.White,
		},
		{name: "плохое действие", id: 1, action: "draw", wantErr: service.ErrBadAction},
		{name: "пустое действие", id: 1, action: "", wantErr: service.ErrBadAction},
		{name: "нет такой игры", id: 100, action: "surrender", wantErr: service.ErrGameNotFound},
		{
			name: "игра уже закончена", id: 1, action: "surrender",
			prepare: func(e *testEnv) {
				_, err := e.svc.UpdateGame(1, "surrender")
				if err != nil {
					panic(err)
				}
			},
			wantErr: game.ErrGameFinished,
		},
		{
			name: "ошибка сохранения", id: 1, action: "surrender",
			prepare: func(e *testEnv) { e.games.saveErr = errDisk },
			wantErr: service.ErrSave,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			g, err := e.svc.UpdateGame(tt.id, tt.action)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, g)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, "finished", g.Status)
			assert.Equal(t, tt.wantWinner, g.Winner)
		})
	}
}

func TestDeleteGame(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		prepare   func(e *testEnv)
		wantErr   error
		wantCount int
	}{
		{name: "успешно", id: 1, wantCount: 0},
		{name: "нет такой", id: 100, wantErr: service.ErrGameNotFound, wantCount: 1},
		{
			name: "ошибка сохранения", id: 1,
			prepare: func(e *testEnv) { e.games.saveErr = errDisk },
			wantErr: service.ErrSave, wantCount: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			err := e.svc.DeleteGame(tt.id)

			assert.Len(t, e.svc.ListGames(), tt.wantCount)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
