package repository_test

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/v2/mongo"

	"HW_14/internal/model/game"
	"HW_14/internal/model/player"
	"HW_14/internal/repository"
)

// интеграционные тесты, нужны запущенные базы: docker compose up -d
// если баз нет, тесты пропускаются

func getEnv(name string, def string) string {
	val := os.Getenv(name)
	if val == "" {
		return def
	}
	return val
}

// отдельная база на каждый тест, после теста удаляем
func testDB(t *testing.T) *mongo.Database {
	client, err := repository.ConnectMongo(getEnv("MONGO_URI", "mongodb://localhost:27017"))
	if err != nil {
		t.Skip("mongodb не запущена:", err)
	}
	db := client.Database(fmt.Sprintf("chess_test_%d", time.Now().UnixNano()))
	t.Cleanup(func() {
		db.Drop(context.Background())
		client.Disconnect(context.Background())
	})
	return db
}

func TestPlayers(t *testing.T) {
	players, err := repository.NewPlayers(testDB(t))
	require.NoError(t, err)

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
	all := players.All()
	require.Len(t, all, 2)
	assert.Equal(t, 1, all[0].ID)
	assert.Equal(t, 2, all[1].ID)

	ok, err := players.Delete(1)
	require.NoError(t, err)
	assert.True(t, ok)
	ok, err = players.Delete(1)
	require.NoError(t, err)
	assert.False(t, ok)
	assert.Nil(t, players.Get(1))
}

func TestGamesAndMoves(t *testing.T) {
	db := testDB(t)
	games, err := repository.NewGames(db)
	require.NoError(t, err)
	moves, err := repository.NewMoves(db)
	require.NoError(t, err)

	white := player.NewPlayer(1, "Аня", player.White)
	black := player.NewPlayer(2, "Боря", player.Black)
	g := game.NewGame(games.NextID(), white, black)
	require.NoError(t, games.Add(g))

	// E2-E4
	m, err := g.MakeMove(moves.NextID(), 6, 4, 4, 4)
	require.NoError(t, err)
	require.NoError(t, moves.Add(m))
	require.NoError(t, games.Update(g))

	// доска и ходы должны вернуться из монги такими же
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

	ok, err := games.Delete(g.ID)
	require.NoError(t, err)
	assert.True(t, ok)
	assert.Empty(t, games.All())
}

func TestRedisLogger(t *testing.T) {
	client, err := repository.ConnectRedis(getEnv("REDIS_ADDR", "localhost:6379"))
	if err != nil {
		t.Skip("redis не запущен:", err)
	}
	defer client.Close()

	ctx := context.Background()
	// чистим старые логи, чтобы проверить только свои
	clearLogs := func() {
		keys, _ := client.Keys(ctx, "log:*").Result()
		if len(keys) > 0 {
			client.Del(ctx, keys...)
		}
	}
	clearLogs()
	t.Cleanup(clearLogs)

	logger := repository.NewRedisLogger(client, time.Minute)
	logger.Log("create player", map[string]int{"id": 1})
	logger.Log("delete player", map[string]int{"id": 1})

	list := logger.All()
	require.Len(t, list, 2)

	var first struct {
		Action string         `json:"action"`
		Data   map[string]int `json:"data"`
	}
	require.NoError(t, json.Unmarshal([]byte(list[0]), &first))
	assert.Equal(t, "create player", first.Action)
	assert.Equal(t, 1, first.Data["id"])
	assert.Contains(t, list[1], "delete player")

	// у каждой записи должен стоять срок жизни
	keys, err := client.Keys(ctx, "log:*").Result()
	require.NoError(t, err)
	for _, key := range keys {
		ttl, err := client.TTL(ctx, key).Result()
		require.NoError(t, err)
		assert.Greater(t, ttl, time.Duration(0))
		assert.LessOrEqual(t, ttl, time.Minute)
	}
}

func TestRedisLoggerExpire(t *testing.T) {
	client, err := repository.ConnectRedis(getEnv("REDIS_ADDR", "localhost:6379"))
	if err != nil {
		t.Skip("redis не запущен:", err)
	}
	defer client.Close()

	logger := repository.NewRedisLogger(client, time.Second)
	logger.Log("expire test", nil)
	time.Sleep(1500 * time.Millisecond)

	for _, l := range logger.All() {
		assert.NotContains(t, l, "expire test")
	}
}
