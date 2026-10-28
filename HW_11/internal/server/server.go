package server

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"sync"

	httpSwagger "github.com/swaggo/http-swagger/v2"

	"HW_11/internal/model/board"
	"HW_11/internal/model/game"
	"HW_11/internal/model/move"
	"HW_11/internal/model/player"
	"HW_11/internal/repository"
)

// чтобы клиенты не мешали друг другу
var mu sync.Mutex

func NewRouter() *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/login", login)

	// тут нужен токен
	mux.HandleFunc("POST /api/player", authRequired(createPlayer))
	mux.HandleFunc("GET /api/players", getPlayers)
	mux.HandleFunc("GET /api/player/{id}", getPlayer)
	mux.HandleFunc("PUT /api/player/{id}", authRequired(updatePlayer))
	mux.HandleFunc("DELETE /api/player/{id}", authRequired(deletePlayer))

	mux.HandleFunc("POST /api/game", authRequired(createGame))
	mux.HandleFunc("GET /api/games", getGames)
	mux.HandleFunc("GET /api/game/{id}", getGame)
	mux.HandleFunc("PUT /api/game/{id}", authRequired(updateGame))
	mux.HandleFunc("DELETE /api/game/{id}", authRequired(deleteGame))

	// ходы не меняем и не удаляем
	mux.HandleFunc("POST /api/move", authRequired(createMove))
	mux.HandleFunc("GET /api/moves", getMoves)
	mux.HandleFunc("GET /api/move/{id}", getMove)

	// swagger
	mux.HandleFunc("GET /swagger/", httpSwagger.WrapHandler)

	return mux
}

// игроки

type PlayerRequest struct {
	Name  string `json:"name" example:"Аня"`
	Color string `json:"color" example:"white"`
}

// createPlayer godoc
//
//	@Summary		Создать игрока
//	@Tags			players
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			input	body		PlayerRequest	true	"имя и цвет (white или black)"
//	@Success		201		{object}	player.Player
//	@Failure		400		{string}	string	"ошибка в данных"
//	@Failure		401		{string}	string	"нужна авторизация"
//	@Router			/player [post]
func createPlayer(w http.ResponseWriter, r *http.Request) {
	var req PlayerRequest
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

// getPlayers godoc
//
//	@Summary	Список игроков
//	@Tags		players
//	@Produce	json
//	@Success	200	{array}	player.Player
//	@Router		/players [get]
func getPlayers(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Players.Items)
}

// getPlayer godoc
//
//	@Summary	Игрок по id
//	@Tags		players
//	@Produce	json
//	@Param		id	path		int	true	"id игрока"
//	@Success	200	{object}	player.Player
//	@Failure	400	{string}	string	"неправильный id"
//	@Failure	404	{string}	string	"игрок не найден"
//	@Router		/player/{id} [get]
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

// updatePlayer godoc
//
//	@Summary		Изменить имя игрока
//	@Description	Меняется только имя, цвет из запроса не используется
//	@Tags			players
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int				true	"id игрока"
//	@Param			input	body		PlayerRequest	true	"новое имя"
//	@Success		200		{object}	player.Player
//	@Failure		400		{string}	string	"ошибка в данных"
//	@Failure		401		{string}	string	"нужна авторизация"
//	@Failure		404		{string}	string	"игрок не найден"
//	@Router			/player/{id} [put]
func updatePlayer(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req PlayerRequest
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
	p.Name = req.Name
	err = repository.Players.Save()
	if err != nil {
		http.Error(w, "ошибка сохранения", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(p)
}

// deletePlayer godoc
//
//	@Summary	Удалить игрока
//	@Tags		players
//	@Security	BearerAuth
//	@Param		id	path	int	true	"id игрока"
//	@Success	204
//	@Failure	400	{string}	string	"неправильный id"
//	@Failure	401	{string}	string	"нужна авторизация"
//	@Failure	404	{string}	string	"игрок не найден"
//	@Router		/player/{id} [delete]
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

type GameRequest struct {
	WhiteID int `json:"white_id" example:"1"`
	BlackID int `json:"black_id" example:"2"`
}

type GameUpdateRequest struct {
	Action string `json:"action" example:"surrender"`
}

// createGame godoc
//
//	@Summary	Создать игру
//	@Tags		games
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		input	body		GameRequest	true	"id белого и чёрного игрока"
//	@Success	201		{object}	game.Game
//	@Failure	400		{string}	string	"ошибка в данных"
//	@Failure	401		{string}	string	"нужна авторизация"
//	@Router		/game [post]
func createGame(w http.ResponseWriter, r *http.Request) {
	var req GameRequest
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

// getGames godoc
//
//	@Summary	Список игр
//	@Tags		games
//	@Produce	json
//	@Success	200	{array}	game.Game
//	@Router		/games [get]
func getGames(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Games.Items)
}

// getGame godoc
//
//	@Summary	Игра по id
//	@Tags		games
//	@Produce	json
//	@Param		id	path		int	true	"id игры"
//	@Success	200	{object}	game.Game
//	@Failure	400	{string}	string	"неправильный id"
//	@Failure	404	{string}	string	"игра не найдена"
//	@Router		/game/{id} [get]
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

// updateGame godoc
//
//	@Summary		Сдаться
//	@Description	Изменить игру можно только одним способом - сдаться. Сдаётся тот, чей сейчас ход
//	@Tags			games
//	@Accept			json
//	@Produce		json
//	@Security		BearerAuth
//	@Param			id		path		int					true	"id игры"
//	@Param			input	body		GameUpdateRequest	true	"action = surrender"
//	@Success		200		{object}	game.Game
//	@Failure		400		{string}	string	"ошибка в данных"
//	@Failure		401		{string}	string	"нужна авторизация"
//	@Failure		404		{string}	string	"игра не найдена"
//	@Router			/game/{id} [put]
func updateGame(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil {
		http.Error(w, "неправильный id", http.StatusBadRequest)
		return
	}
	var req GameUpdateRequest
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

// deleteGame godoc
//
//	@Summary	Удалить игру
//	@Tags		games
//	@Security	BearerAuth
//	@Param		id	path	int	true	"id игры"
//	@Success	204
//	@Failure	400	{string}	string	"неправильный id"
//	@Failure	401	{string}	string	"нужна авторизация"
//	@Failure	404	{string}	string	"игра не найдена"
//	@Router		/game/{id} [delete]
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

type MoveRequest struct {
	GameID int    `json:"game_id" example:"1"`
	From   string `json:"from" example:"E2"`
	To     string `json:"to" example:"E4"`
}

// createMove godoc
//
//	@Summary	Сделать ход
//	@Tags		moves
//	@Accept		json
//	@Produce	json
//	@Security	BearerAuth
//	@Param		input	body		MoveRequest	true	"номер игры и ход"
//	@Success	201		{object}	move.Move
//	@Failure	400		{string}	string	"ошибка в данных или ход невозможен"
//	@Failure	401		{string}	string	"нужна авторизация"
//	@Failure	404		{string}	string	"игра не найдена"
//	@Router		/move [post]
func createMove(w http.ResponseWriter, r *http.Request) {
	var req MoveRequest
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

// getMoves godoc
//
//	@Summary	Список ходов
//	@Tags		moves
//	@Produce	json
//	@Success	200	{array}	move.Move
//	@Router		/moves [get]
func getMoves(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(repository.Moves.Items)
}

// getMove godoc
//
//	@Summary	Ход по id
//	@Tags		moves
//	@Produce	json
//	@Param		id	path		int	true	"id хода"
//	@Success	200	{object}	move.Move
//	@Failure	400	{string}	string	"неправильный id"
//	@Failure	404	{string}	string	"ход не найден"
//	@Router		/move/{id} [get]
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
