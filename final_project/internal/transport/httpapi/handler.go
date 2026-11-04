// Package httpapi - транспортный слой: принимает HTTP-запросы и вызывает сервис.
package httpapi

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"tictactoe/internal/api"
	"tictactoe/internal/game"
	"tictactoe/internal/service"
)

// maxBodySize - максимальный размер тела запроса. Нам хватит и меньшего.
const maxBodySize = 1024

// Handler обрабатывает запросы к API игры.
type Handler struct {
	svc *service.Service
}

// New создаёт обработчик.
func New(svc *service.Service) *Handler {
	return &Handler{svc: svc}
}

// Routes возвращает роутер со всеми ручками API:
//
//	POST /games              - создать игру
//	POST /games/{id}/join    - присоединиться к игре
//	GET  /games/{id}         - состояние игры
//	POST /games/{id}/moves   - сделать ход (нужен заголовок X-Player-Token)
//
// Подробное описание есть в README.md.
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /games", h.createGame)
	mux.HandleFunc("POST /games/{id}/join", h.joinGame)
	mux.HandleFunc("GET /games/{id}", h.getState)
	mux.HandleFunc("POST /games/{id}/moves", h.move)
	return mux
}

func (h *Handler) createGame(w http.ResponseWriter, r *http.Request) {
	var req api.CreateGameRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	resp, err := h.svc.CreateGame(req.Size)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, resp)
}

func (h *Handler) joinGame(w http.ResponseWriter, r *http.Request) {
	resp, err := h.svc.JoinGame(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, resp)
}

func (h *Handler) getState(w http.ResponseWriter, r *http.Request) {
	state, err := h.svc.State(r.PathValue("id"))
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func (h *Handler) move(w http.ResponseWriter, r *http.Request) {
	var req api.MoveRequest
	if err := readJSON(w, r, &req); err != nil {
		writeError(w, http.StatusBadRequest, "некорректный JSON")
		return
	}

	token := r.Header.Get(api.TokenHeader)
	state, err := h.svc.Move(r.PathValue("id"), token, req.Row, req.Col)
	if err != nil {
		writeServiceError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, state)
}

func readJSON(w http.ResponseWriter, r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}

// writeServiceError переводит ошибку сервиса или игры в HTTP-код.
func writeServiceError(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, service.ErrGameNotFound):
		writeError(w, http.StatusNotFound, err.Error())
	case errors.Is(err, service.ErrBadToken):
		writeError(w, http.StatusForbidden, err.Error())
	case errors.Is(err, service.ErrGameFull),
		errors.Is(err, service.ErrWaiting),
		errors.Is(err, game.ErrNotYourTurn),
		errors.Is(err, game.ErrCellBusy),
		errors.Is(err, game.ErrGameOver):
		writeError(w, http.StatusConflict, err.Error())
	case errors.Is(err, game.ErrBadSize),
		errors.Is(err, game.ErrOutOfBoard),
		errors.Is(err, game.ErrBadPlayer):
		writeError(w, http.StatusBadRequest, err.Error())
	case errors.Is(err, service.ErrTooManyGames):
		writeError(w, http.StatusServiceUnavailable, err.Error())
	default:
		log.Printf("внутренняя ошибка: %v", err)
		writeError(w, http.StatusInternalServerError, "внутренняя ошибка сервера")
	}
}

func writeError(w http.ResponseWriter, code int, msg string) {
	writeJSON(w, code, api.ErrorResponse{Error: msg})
}

func writeJSON(w http.ResponseWriter, code int, v interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("не удалось записать ответ: %v", err)
	}
}
