package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"HW_10/internal/model/board"
	"HW_10/internal/model/game"
	"HW_10/internal/model/move"
	"HW_10/internal/model/player"
	"HW_10/internal/repository"
)

// чтобы клиенты не мешали друг другу
var mu sync.Mutex

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/player", createPlayer)
	mux.HandleFunc("GET /api/players", getPlayers)
	mux.HandleFunc("GET /api/player/{id}", getPlayer)
	mux.HandleFunc("PUT /api/player/{id}", updatePlayer)
	mux.HandleFunc("DELETE /api/player/{id}", deletePlayer)

	mux.HandleFunc("POST /api/game", createGame)
	mux.HandleFunc("GET /api/games", getGames)
	mux.HandleFunc("GET /api/game/{id}", getGame)
	mux.HandleFunc("PUT /api/game/{id}", updateGame)
	mux.HandleFunc("DELETE /api/game/{id}", deleteGame)

	// ходы не меняем и не удаляем
	mux.HandleFunc("POST /api/move", createMove)
	mux.HandleFunc("GET /api/moves", getMoves)
	mux.HandleFunc("GET /api/move/{id}", getMove)

	return mux
}

// игроки

type playerRequest struct {
	Name  string       `json:"name"`
	Color player.Color `json:"color"`
}

func createPlayer(w http.ResponseWriter, r *http.Request) {
	var req playerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "не указано имя", http.StatusBadRequest)
		return
	}
	if req.Color != player.White && req.Color != player.Black {
		http.Error(w, "цвет должен быть white или black", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p := player.NewPlayer(repository.Players.NextID(), req.Name, req.Color)
	err = repository.Players.Add(p)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(p)
}

func getPlayers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Players.Items)
}

func getPlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p := repository.Players.Get(id)
	if p == nil {
		http.Error(w, "игрок не найден", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// меняем только имя
func updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req playerRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}
	if req.Name == "" {
		http.Error(w, "не указано имя", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p := repository.Players.Get(id)
	if p == nil {
		http.Error(w, "игрок не найден", http.StatusNotFound)
		return
	}
	p.SetName(req.Name)
	err = repository.Players.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

func deletePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if !repository.Players.Delete(id) {
		http.Error(w, "игрок не найден", http.StatusNotFound)
		return
	}
	err = repository.Players.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// игры

func createGame(w http.ResponseWriter, r *http.Request) {
	var req struct {
		WhiteID int `json:"white_id"`
		BlackID int `json:"black_id"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	white := repository.Players.Get(req.WhiteID)
	black := repository.Players.Get(req.BlackID)
	if white == nil || black == nil {
		http.Error(w, "нужны white_id и black_id существующих игроков", http.StatusBadRequest)
		return
	}

	g := game.NewGame(repository.Games.NextID(), white, black)
	err = repository.Games.Add(g)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(g)
}

func getGames(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Games.Items)
}

func getGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(id)
	if g == nil {
		http.Error(w, "игра не найдена", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

// игру можно только сдать
// тело запроса: {"action": "surrender"}
func updateGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}
	if req.Action != "surrender" {
		http.Error(w, "action может быть только surrender", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(id)
	if g == nil {
		http.Error(w, "игра не найдена", http.StatusNotFound)
		return
	}
	err = g.Surrender()
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = repository.Games.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(g)
}

func deleteGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	if !repository.Games.Delete(id) {
		http.Error(w, "игра не найдена", http.StatusNotFound)
		return
	}
	err = repository.Games.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// ходы

func createMove(w http.ResponseWriter, r *http.Request) {
	var req struct {
		GameID int    `json:"game_id"`
		From   string `json:"from"`
		To     string `json:"to"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}
	fromRow, fromCol, ok1 := board.ParseSquare(strings.ToUpper(req.From))
	toRow, toCol, ok2 := board.ParseSquare(strings.ToUpper(req.To))
	if !ok1 || !ok2 {
		http.Error(w, "from и to нужно писать так: E2, E4", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(req.GameID)
	if g == nil {
		http.Error(w, "игра не найдена", http.StatusNotFound)
		return
	}
	var m *move.Move
	m, err = g.MakeMove(repository.Moves.NextID(), fromRow, fromCol, toRow, toCol)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	err = repository.Moves.Add(m)
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	// игру тоже сохраняем
	err = repository.Games.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(m)
}

func getMoves(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Moves.Items)
}

func getMove(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	m := repository.Moves.Get(id)
	if m == nil {
		http.Error(w, "ход не найден", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(m)
}
