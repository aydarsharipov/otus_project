package server

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"sync"

	"HW_13/internal/service"
)

// чтобы клиенты не мешали друг другу
var mu sync.Mutex

var svc *service.Service

func NewRouter(s *service.Service) *http.ServeMux {
	svc = s
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

// отправляем ошибку
func sendError(w http.ResponseWriter, err error) {
	code := http.StatusBadRequest
	if errors.Is(err, service.ErrPlayerNotFound) || errors.Is(err, service.ErrGameNotFound) || errors.Is(err, service.ErrMoveNotFound) {
		code = http.StatusNotFound
	}
	if errors.Is(err, service.ErrSave) {
		code = http.StatusInternalServerError
	}
	http.Error(w, err.Error(), code)
}

func sendJSON(w http.ResponseWriter, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(data)
}

func getID(r *http.Request) (int, bool) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		return 0, false
	}
	return id, true
}

// игроки

type playerRequest struct {
	Name  string `json:"name"`
	Color string `json:"color"`
}

func createPlayer(w http.ResponseWriter, r *http.Request) {
	var req playerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p, err := svc.CreatePlayer(req.Name, req.Color)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, p)
}

func getPlayers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sendJSON(w, http.StatusOK, svc.ListPlayers())
}

func getPlayer(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p, err := svc.GetPlayer(id)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, p)
}

func updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req playerRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	p, err := svc.RenamePlayer(id, req.Name)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, p)
}

func deletePlayer(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	err := svc.DeletePlayer(id)
	if err != nil {
		sendError(w, err)
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

	g, err := svc.CreateGame(req.WhiteID, req.BlackID)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, g)
}

func getGames(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sendJSON(w, http.StatusOK, svc.ListGames())
}

func getGame(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	g, err := svc.GetGame(id)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, g)
}

// тело запроса: {"action": "surrender"}
func updateGame(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req struct {
		Action string `json:"action"`
	}
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "неправильный json", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	g, err := svc.UpdateGame(id, req.Action)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, g)
}

func deleteGame(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	err := svc.DeleteGame(id)
	if err != nil {
		sendError(w, err)
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

	mu.Lock()
	defer mu.Unlock()

	m, err := svc.MakeMove(req.GameID, req.From, req.To)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusCreated, m)
}

func getMoves(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	sendJSON(w, http.StatusOK, svc.ListMoves())
}

func getMove(w http.ResponseWriter, r *http.Request) {
	id, ok := getID(r)
	if !ok {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	m, err := svc.GetMove(id)
	if err != nil {
		sendError(w, err)
		return
	}
	sendJSON(w, http.StatusOK, m)
}
