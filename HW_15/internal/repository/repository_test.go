package repository_test

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"HW_15/internal/model/game"
	"HW_15/internal/model/move"
	"HW_15/internal/model/player"
	"HW_15/internal/repository"
)

// интеграционные тесты, нужен запущенный postgres: docker compose up -d
// и переменная POSTGRES_PASSWORD. если базы нет, тесты пропускаются

func dbURL(dbName string) string {
	host := os.Getenv("POSTGRES_HOST")
	if host == "" {
		host = "localhost:5432"
	}
	u := url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword("chess", os.Getenv("POSTGRES_PASSWORD")),
		Host:     host,
		Path:     dbName,
		RawQuery: "sslmode=disable",
	}
	return u.String()
}

// на каждый тест своя пустая база, после теста удаляем
func testPool(t *testing.T) *pgxpool.Pool {
	if os.Getenv("POSTGRES_PASSWORD") == "" {
		t.Skip("не задан POSTGRES_PASSWORD")
	}
	admin, err := repository.Connect(dbURL("chess"))
	if err != nil {
		t.Skip("postgres не запущен:", err)
	}

	// имя из времени, пользовательского ввода тут нет
	name := fmt.Sprintf("chess_test_%d", time.Now().UnixNano())
	_, err = admin.Exec(context.Background(), "CREATE DATABASE "+name)
	require.NoError(t, err)

	pool, err := repository.Connect(dbURL(name))
	require.NoError(t, err)
	require.NoError(t, repository.Migrate(pool))

	t.Cleanup(func() {
		pool.Close()
		admin.Exec(context.Background(), "DROP DATABASE "+name)
		admin.Close()
	})
	return pool
}

func TestPlayers(t *testing.T) {
	players := repository.NewPlayers(testPool(t))

	assert.Equal(t, 1, players.NextID())
	assert.Empty(t, players.All())

	p := player.NewPlayer(players.NextID(), "Аня", player.White)
	require.NoError(t, players.Add(p))
	assert.Equal(t, 2, players.NextID())

	// тот же id второй раз нельзя
	assert.Error(t, players.Add(player.NewPlayer(1, "Боря", player.Black)))

	got := players.Get(1)
	require.NotNil(t, got)
	assert.Equal(t, *p, *got)
	assert.Nil(t, players.Get(100))

	p.Name = "Анна"
	require.NoError(t, players.Update(p))
	assert.Equal(t, "Анна", players.Get(1).Name)

	require.NoError(t, players.Add(player.NewPlayer(2, "Боря", player.Black)))
	assert.Len(t, players.All(), 2)

	ok, err := players.Delete(1)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = players.Delete(1)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, players.Get(1))
}

func TestGamesAndMoves(t *testing.T) {
	pool := testPool(t)
	players := repository.NewPlayers(pool)
	games := repository.NewGames(pool)
	moves := repository.NewMoves(pool)

	white := player.NewPlayer(1, "Аня", player.White)
	black := player.NewPlayer(2, "Боря", player.Black)
	require.NoError(t, players.Add(white))
	require.NoError(t, players.Add(black))

	g := game.NewGame(games.NextID(), white, black)
	require.NoError(t, games.Add(g))

	// E2-E4: ход и игра сохраняются в одной транзакции
	m, err := g.MakeMove(moves.NextID(), 6, 4, 4, 4)
	require.NoError(t, err)
	require.NoError(t, games.AddMove(g, m))

	// доска и ходы должны вернуться из базы такими же
	got := games.Get(g.ID)
	require.NotNil(t, got)
	assert.Equal(t, g, got)
	assert.Nil(t, got.Board.Cells[6][4])
	assert.Equal(t, player.Black, got.Turn)

	gotMove := moves.Get(m.ID)
	require.NotNil(t, gotMove)
	assert.Equal(t, *m, *gotMove)
	assert.Len(t, moves.All(), 1)
	assert.Nil(t, moves.Get(100))

	require.NoError(t, got.Surrender())
	require.NoError(t, games.Update(got))
	assert.Equal(t, "finished", games.Get(g.ID).Status)
	assert.Len(t, games.All(), 1)

	// игрока, у которого есть игра, удалить нельзя
	_, err = players.Delete(1)
	assert.Error(t, err)

	// вместе с игрой удаляются её ходы
	ok, err := games.Delete(g.ID)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Empty(t, games.All())
	assert.Empty(t, moves.All())
}

// если второй запрос в транзакции упал, первый тоже отменяется
func TestAddMoveRollback(t *testing.T) {
	pool := testPool(t)
	players := repository.NewPlayers(pool)
	games := repository.NewGames(pool)
	moves := repository.NewMoves(pool)

	white := player.NewPlayer(1, "Аня", player.White)
	black := player.NewPlayer(2, "Боря", player.Black)
	require.NoError(t, players.Add(white))
	require.NoError(t, players.Add(black))
	g := game.NewGame(1, white, black)
	require.NoError(t, games.Add(g))

	m, err := g.MakeMove(1, 6, 4, 4, 4)
	require.NoError(t, err)
	require.NoError(t, games.AddMove(g, m))

	// ход с тем же id: вставка упадёт, доска в базе меняться не должна
	m2, err := g.MakeMove(1, 1, 4, 3, 4)
	require.NoError(t, err)
	assert.Error(t, games.AddMove(g, m2))

	got := games.Get(1)
	require.NotNil(t, got)
	assert.Equal(t, player.Black, got.Turn)
	assert.NotNil(t, got.Board.Cells[1][4], "пешка E7 должна остаться на месте")
	assert.Equal(t, []*move.Move{m}, got.Moves)
	assert.Len(t, moves.All(), 1)
}
