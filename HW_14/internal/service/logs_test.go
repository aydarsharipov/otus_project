package service_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// каждое изменение попадает в лог по порядку
func TestListLogs(t *testing.T) {
	e := newEnv()
	assert.Empty(t, e.svc.ListLogs())

	_, _ = e.svc.CreatePlayer("Анна", "white")
	_, _ = e.svc.RenamePlayer(3, "Аня")
	_, _ = e.svc.CreateGame(1, 2)
	mustMove(e, "E2", "E4")
	_, _ = e.svc.UpdateGame(1, "surrender")
	_ = e.svc.DeleteGame(2)
	_ = e.svc.DeletePlayer(3)

	// чтение лог не пишет
	_, _ = e.svc.GetPlayer(1)
	e.svc.ListGames()

	want := []string{
		"create player",
		"rename player",
		"create game",
		"move",
		"surrender",
		"delete game",
		"delete player",
	}
	assert.Equal(t, want, e.svc.ListLogs())
}
