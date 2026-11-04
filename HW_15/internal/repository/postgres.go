package repository

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"

	"HW_15/internal/db"
	"HW_15/internal/model/board"
	"HW_15/internal/model/game"
	"HW_15/internal/model/move"
	"HW_15/internal/model/piece"
	"HW_15/internal/model/player"
	"HW_15/migrations"
)

// чтобы запрос не висел вечно
func newCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// подключаемся к постгресу и проверяем что он отвечает
func Connect(url string) (*pgxpool.Pool, error) {
	ctx, cancel := newCtx()
	defer cancel()
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	err = pool.Ping(ctx)
	if err != nil {
		pool.Close()
		return nil, err
	}
	return pool, nil
}

// создаём таблицы и добавляем поля, goose сам помнит какие миграции уже были
func Migrate(pool *pgxpool.Pool) error {
	// goose работает через database/sql, поэтому делаем обёртку над пулом
	sqlDB := stdlib.OpenDBFromPool(pool)
	defer sqlDB.Close()

	goose.SetBaseFS(migrations.FS)
	err := goose.SetDialect("postgres")
	if err != nil {
		return err
	}
	return goose.Up(sqlDB, ".")
}

// игроки

type Players struct {
	q *db.Queries
}

func NewPlayers(pool *pgxpool.Pool) *Players {
	return &Players{q: db.New(pool)}
}

func toPlayer(row db.Player) *player.Player {
	return player.NewPlayer(int(row.ID), row.Name, row.Color)
}

func (s *Players) NextID() int {
	ctx, cancel := newCtx()
	defer cancel()
	id, err := s.q.NextPlayerID(ctx)
	if err != nil {
		fmt.Println("ошибка получения id:", err)
		return 0
	}
	return int(id)
}

func (s *Players) Add(p *player.Player) error {
	ctx, cancel := newCtx()
	defer cancel()
	return s.q.CreatePlayer(ctx, db.CreatePlayerParams{
		ID:    int32(p.ID),
		Name:  p.Name,
		Color: p.Color,
	})
}

func (s *Players) Get(id int) *player.Player {
	ctx, cancel := newCtx()
	defer cancel()
	row, err := s.q.GetPlayer(ctx, int32(id))
	if err != nil {
		// не нашли
		return nil
	}
	return toPlayer(row)
}

func (s *Players) All() []*player.Player {
	ctx, cancel := newCtx()
	defer cancel()
	list := []*player.Player{}
	rows, err := s.q.ListPlayers(ctx)
	if err != nil {
		fmt.Println("ошибка чтения игроков:", err)
		return list
	}
	for _, row := range rows {
		list = append(list, toPlayer(row))
	}
	return list
}

func (s *Players) Update(p *player.Player) error {
	ctx, cancel := newCtx()
	defer cancel()
	return s.q.UpdatePlayer(ctx, db.UpdatePlayerParams{
		ID:    int32(p.ID),
		Name:  p.Name,
		Color: p.Color,
	})
}

func (s *Players) Delete(id int) (bool, error) {
	ctx, cancel := newCtx()
	defer cancel()
	count, err := s.q.DeletePlayer(ctx, int32(id))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// игры

type Games struct {
	pool *pgxpool.Pool // нужен для транзакций
	q    *db.Queries
}

func NewGames(pool *pgxpool.Pool) *Games {
	return &Games{pool: pool, q: db.New(pool)}
}

// собираем игру из строки таблицы: игроки и ходы лежат в своих таблицах
func (s *Games) toGame(ctx context.Context, row db.Game) (*game.Game, error) {
	white, err := s.q.GetPlayer(ctx, row.WhiteID)
	if err != nil {
		return nil, err
	}
	black, err := s.q.GetPlayer(ctx, row.BlackID)
	if err != nil {
		return nil, err
	}
	moveRows, err := s.q.ListGameMoves(ctx, row.ID)
	if err != nil {
		return nil, err
	}

	g := &game.Game{
		ID:     int(row.ID),
		White:  toPlayer(white),
		Black:  toPlayer(black),
		Board:  &board.Board{},
		Moves:  []*move.Move{},
		Turn:   row.Turn,
		Status: row.Status,
		Winner: row.Winner,
	}
	err = json.Unmarshal(row.Board, g.Board)
	if err != nil {
		return nil, err
	}
	for _, mr := range moveRows {
		m, err := toMove(mr)
		if err != nil {
			return nil, err
		}
		g.Moves = append(g.Moves, m)
	}
	return g, nil
}

func (s *Games) NextID() int {
	ctx, cancel := newCtx()
	defer cancel()
	id, err := s.q.NextGameID(ctx)
	if err != nil {
		fmt.Println("ошибка получения id:", err)
		return 0
	}
	return int(id)
}

func (s *Games) Add(g *game.Game) error {
	boardJSON, err := json.Marshal(g.Board)
	if err != nil {
		return err
	}
	ctx, cancel := newCtx()
	defer cancel()
	return s.q.CreateGame(ctx, db.CreateGameParams{
		ID:      int32(g.ID),
		WhiteID: int32(g.White.ID),
		BlackID: int32(g.Black.ID),
		Board:   boardJSON,
		Turn:    g.Turn,
		Status:  g.Status,
		Winner:  g.Winner,
	})
}

func (s *Games) Get(id int) *game.Game {
	ctx, cancel := newCtx()
	defer cancel()
	row, err := s.q.GetGame(ctx, int32(id))
	if err != nil {
		// не нашли
		return nil
	}
	g, err := s.toGame(ctx, row)
	if err != nil {
		fmt.Println("ошибка чтения игры:", err)
		return nil
	}
	return g
}

func (s *Games) All() []*game.Game {
	ctx, cancel := newCtx()
	defer cancel()
	list := []*game.Game{}
	rows, err := s.q.ListGames(ctx)
	if err != nil {
		fmt.Println("ошибка чтения игр:", err)
		return list
	}
	for _, row := range rows {
		g, err := s.toGame(ctx, row)
		if err != nil {
			fmt.Println("ошибка чтения игры:", err)
			continue
		}
		list = append(list, g)
	}
	return list
}

func updateGameParams(g *game.Game) (db.UpdateGameParams, error) {
	boardJSON, err := json.Marshal(g.Board)
	if err != nil {
		return db.UpdateGameParams{}, err
	}
	return db.UpdateGameParams{
		ID:     int32(g.ID),
		Board:  boardJSON,
		Turn:   g.Turn,
		Status: g.Status,
		Winner: g.Winner,
	}, nil
}

func (s *Games) Update(g *game.Game) error {
	params, err := updateGameParams(g)
	if err != nil {
		return err
	}
	ctx, cancel := newCtx()
	defer cancel()
	return s.q.UpdateGame(ctx, params)
}

func (s *Games) Delete(id int) (bool, error) {
	ctx, cancel := newCtx()
	defer cancel()
	count, err := s.q.DeleteGame(ctx, int32(id))
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// ход меняет данные в двух таблицах: добавляем строку в moves и обновляем доску в games.
// делаем это в одной транзакции, чтобы не было хода без изменения доски и наоборот
func (s *Games) AddMove(g *game.Game, m *move.Move) error {
	pieceJSON, err := json.Marshal(m.Piece)
	if err != nil {
		return err
	}
	gameParams, err := updateGameParams(g)
	if err != nil {
		return err
	}

	ctx, cancel := newCtx()
	defer cancel()

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	// если вышли с ошибкой, всё отменится. после Commit откат уже ничего не делает
	defer tx.Rollback(ctx)

	// те же запросы, только внутри транзакции
	qtx := s.q.WithTx(tx)

	err = qtx.CreateMove(ctx, db.CreateMoveParams{
		ID:       int32(m.ID),
		GameID:   int32(m.GameID),
		FromRow:  int32(m.FromRow),
		FromCol:  int32(m.FromCol),
		ToRow:    int32(m.ToRow),
		ToCol:    int32(m.ToCol),
		Piece:    pieceJSON,
		Notation: m.Notation,
	})
	if err != nil {
		return err
	}

	err = qtx.UpdateGame(ctx, gameParams)
	if err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// ходы

type Moves struct {
	q *db.Queries
}

func NewMoves(pool *pgxpool.Pool) *Moves {
	return &Moves{q: db.New(pool)}
}

func toMove(row db.Move) (*move.Move, error) {
	m := &move.Move{
		ID:       int(row.ID),
		GameID:   int(row.GameID),
		FromRow:  int(row.FromRow),
		FromCol:  int(row.FromCol),
		ToRow:    int(row.ToRow),
		ToCol:    int(row.ToCol),
		Piece:    &piece.Piece{},
		Notation: row.Notation,
	}
	err := json.Unmarshal(row.Piece, m.Piece)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *Moves) NextID() int {
	ctx, cancel := newCtx()
	defer cancel()
	id, err := s.q.NextMoveID(ctx)
	if err != nil {
		fmt.Println("ошибка получения id:", err)
		return 0
	}
	return int(id)
}

func (s *Moves) Get(id int) *move.Move {
	ctx, cancel := newCtx()
	defer cancel()
	row, err := s.q.GetMove(ctx, int32(id))
	if err != nil {
		// не нашли
		return nil
	}
	m, err := toMove(row)
	if err != nil {
		fmt.Println("ошибка чтения хода:", err)
		return nil
	}
	return m
}

func (s *Moves) All() []*move.Move {
	ctx, cancel := newCtx()
	defer cancel()
	list := []*move.Move{}
	rows, err := s.q.ListMoves(ctx)
	if err != nil {
		fmt.Println("ошибка чтения ходов:", err)
		return list
	}
	for _, row := range rows {
		m, err := toMove(row)
		if err != nil {
			fmt.Println("ошибка чтения хода:", err)
			continue
		}
		list = append(list, m)
	}
	return list
}
