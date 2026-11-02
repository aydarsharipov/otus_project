package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HW_13/internal/service"
)

func TestCreatePlayer(t *testing.T) {
	tests := []struct {
		name      string
		inName    string
		inColor   string
		prepare   func(e *testEnv)
		wantName  string
		wantErr   error
		wantCount int // сколько игроков должно стать
	}{
		{name: "белый игрок", inName: "Анна", inColor: "white", wantName: "Анна", wantCount: 3},
		{name: "чёрный игрок", inName: "Олег", inColor: "black", wantName: "Олег", wantCount: 3},
		{name: "пробелы обрезаются", inName: "  Анна  ", inColor: "white", wantName: "Анна", wantCount: 3},
		{name: "пустое имя", inName: "", inColor: "white", wantErr: service.ErrEmptyName, wantCount: 2},
		{name: "имя из пробелов", inName: "   ", inColor: "white", wantErr: service.ErrEmptyName, wantCount: 2},
		{name: "плохой цвет", inName: "Анна", inColor: "red", wantErr: service.ErrBadColor, wantCount: 2},
		{name: "цвет большими буквами", inName: "Анна", inColor: "WHITE", wantErr: service.ErrBadColor, wantCount: 2},
		{
			name: "ошибка сохранения", inName: "Анна", inColor: "white",
			prepare: func(e *testEnv) { e.players.saveErr = errDisk },
			wantErr: service.ErrSave, wantCount: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			p, err := e.svc.CreatePlayer(tt.inName, tt.inColor)

			assert.Len(t, e.svc.ListPlayers(), tt.wantCount)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, p)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, 3, p.ID)
			assert.Equal(t, tt.wantName, p.Name)
			assert.Equal(t, tt.inColor, p.Color)
		})
	}
}

func TestGetPlayer(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		wantName string
		wantErr  error
	}{
		{name: "первый игрок", id: 1, wantName: "Иван"},
		{name: "второй игрок", id: 2, wantName: "Пётр"},
		{name: "нет такого", id: 100, wantErr: service.ErrPlayerNotFound},
		{name: "ноль", id: 0, wantErr: service.ErrPlayerNotFound},
		{name: "минус", id: -1, wantErr: service.ErrPlayerNotFound},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()

			p, err := e.svc.GetPlayer(tt.id)

			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, p)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, p.Name)
		})
	}
}

func TestListPlayers(t *testing.T) {
	e := newEnv()

	list := e.svc.ListPlayers()

	require.Len(t, list, 2)
	assert.Equal(t, "Иван", list[0].Name)
	assert.Equal(t, "Пётр", list[1].Name)
}

func TestRenamePlayer(t *testing.T) {
	tests := []struct {
		name     string
		id       int
		newName  string
		prepare  func(e *testEnv)
		wantName string // какое имя должно быть у игрока 1 после вызова
		wantErr  error
	}{
		{name: "успешно", id: 1, newName: "Ваня", wantName: "Ваня"},
		{name: "пробелы обрезаются", id: 1, newName: " Ваня ", wantName: "Ваня"},
		{name: "пустое имя", id: 1, newName: "", wantName: "Иван", wantErr: service.ErrEmptyName},
		{name: "нет такого игрока", id: 100, newName: "Ваня", wantName: "Иван", wantErr: service.ErrPlayerNotFound},
		{
			name: "ошибка сохранения", id: 1, newName: "Ваня",
			prepare:  func(e *testEnv) { e.players.saveErr = errDisk },
			wantName: "Ваня", wantErr: service.ErrSave,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			p, err := e.svc.RenamePlayer(tt.id, tt.newName)

			got, _ := e.svc.GetPlayer(1)
			assert.Equal(t, tt.wantName, got.Name)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				assert.Nil(t, p)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.wantName, p.Name)
		})
	}
}

func TestDeletePlayer(t *testing.T) {
	tests := []struct {
		name      string
		id        int
		prepare   func(e *testEnv)
		wantErr   error
		wantCount int
	}{
		{name: "успешно", id: 1, wantCount: 1},
		{name: "нет такого", id: 100, wantErr: service.ErrPlayerNotFound, wantCount: 2},
		{
			name: "ошибка сохранения", id: 2,
			prepare: func(e *testEnv) { e.players.saveErr = errDisk },
			wantErr: service.ErrSave, wantCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := newEnv()
			if tt.prepare != nil {
				tt.prepare(e)
			}

			err := e.svc.DeletePlayer(tt.id)

			assert.Len(t, e.svc.ListPlayers(), tt.wantCount)
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
				return
			}
			require.NoError(t, err)
			_, err = e.svc.GetPlayer(tt.id)
			assert.ErrorIs(t, err, service.ErrPlayerNotFound)
		})
	}
}
