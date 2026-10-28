package main

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"

	"HW_12/internal/model/board"
	"HW_12/internal/model/game"
	"HW_12/internal/model/move"
	"HW_12/internal/model/player"
	"HW_12/internal/repository"
	"HW_12/pkg/chess_api"
)

// чтобы клиенты не мешали друг другу
var mu sync.Mutex

type server struct {
	chess_api.UnimplementedChessServiceServer
}

// перевод в proto

func playerToProto(p *player.Player) *chess_api.Player {
	return &chess_api.Player{Id: int32(p.ID), Name: p.Name, Color: p.Color}
}

func gameToProto(g *game.Game) *chess_api.Game {
	rows := []string{}
	for r := 0; r < board.Size; r++ {
		row := ""
		for c := 0; c < board.Size; c++ {
			if g.Board.Cells[r][c] == nil {
				row += "."
			} else {
				row += g.Board.Cells[r][c].Symbol
			}
		}
		rows = append(rows, row)
	}

	return &chess_api.Game{
		Id:         int32(g.ID),
		White:      playerToProto(g.White),
		Black:      playerToProto(g.Black),
		Board:      rows,
		Turn:       g.Turn,
		Status:     g.Status,
		Winner:     g.Winner,
		MovesCount: int32(len(g.Moves)),
	}
}

func moveToProto(m *move.Move) *chess_api.Move {
	return &chess_api.Move{
		Id:       int32(m.ID),
		GameId:   int32(m.GameID),
		From:     board.SquareName(m.FromRow, m.FromCol),
		To:       board.SquareName(m.ToRow, m.ToCol),
		Notation: m.Notation,
	}
}

// игроки

func (s *server) CreatePlayer(ctx context.Context, req *chess_api.CreatePlayerRequest) (*chess_api.Player, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "не указано имя")
	}
	if req.Color != player.White && req.Color != player.Black {
		return nil, status.Error(codes.InvalidArgument, "цвет должен быть white или black")
	}

	mu.Lock()
	defer mu.Unlock()

	p := player.NewPlayer(repository.Players.NextID(), req.Name, req.Color)
	err := repository.Players.Add(p)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return playerToProto(p), nil
}

func (s *server) GetPlayer(ctx context.Context, req *chess_api.IdRequest) (*chess_api.Player, error) {
	mu.Lock()
	defer mu.Unlock()

	p := repository.Players.Get(int(req.Id))
	if p == nil {
		return nil, status.Error(codes.NotFound, "игрок не найден")
	}
	return playerToProto(p), nil
}

func (s *server) ListPlayers(ctx context.Context, req *chess_api.Empty) (*chess_api.PlayerList, error) {
	mu.Lock()
	defer mu.Unlock()

	list := &chess_api.PlayerList{}
	for _, p := range repository.Players.Items {
		list.Players = append(list.Players, playerToProto(p))
	}
	return list, nil
}

// меняем только имя
func (s *server) UpdatePlayer(ctx context.Context, req *chess_api.UpdatePlayerRequest) (*chess_api.Player, error) {
	if req.Name == "" {
		return nil, status.Error(codes.InvalidArgument, "не указано имя")
	}

	mu.Lock()
	defer mu.Unlock()

	p := repository.Players.Get(int(req.Id))
	if p == nil {
		return nil, status.Error(codes.NotFound, "игрок не найден")
	}
	p.Name = req.Name
	err := repository.Players.Save()
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return playerToProto(p), nil
}

func (s *server) DeletePlayer(ctx context.Context, req *chess_api.IdRequest) (*chess_api.Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	if !repository.Players.Delete(int(req.Id)) {
		return nil, status.Error(codes.NotFound, "игрок не найден")
	}
	err := repository.Players.Save()
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return &chess_api.Empty{}, nil
}

// игры

func (s *server) CreateGame(ctx context.Context, req *chess_api.CreateGameRequest) (*chess_api.Game, error) {
	mu.Lock()
	defer mu.Unlock()

	white := repository.Players.Get(int(req.WhiteId))
	black := repository.Players.Get(int(req.BlackId))
	if white == nil || black == nil {
		return nil, status.Error(codes.InvalidArgument, "нужны white_id и black_id существующих игроков")
	}

	g := game.NewGame(repository.Games.NextID(), white, black)
	err := repository.Games.Add(g)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return gameToProto(g), nil
}

func (s *server) GetGame(ctx context.Context, req *chess_api.IdRequest) (*chess_api.Game, error) {
	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(int(req.Id))
	if g == nil {
		return nil, status.Error(codes.NotFound, "игра не найдена")
	}
	return gameToProto(g), nil
}

func (s *server) ListGames(ctx context.Context, req *chess_api.Empty) (*chess_api.GameList, error) {
	mu.Lock()
	defer mu.Unlock()

	list := &chess_api.GameList{}
	for _, g := range repository.Games.Items {
		list.Games = append(list.Games, gameToProto(g))
	}
	return list, nil
}

// игру можно только сдать
func (s *server) UpdateGame(ctx context.Context, req *chess_api.UpdateGameRequest) (*chess_api.Game, error) {
	if req.Action != "surrender" {
		return nil, status.Error(codes.InvalidArgument, "action может быть только surrender")
	}

	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(int(req.Id))
	if g == nil {
		return nil, status.Error(codes.NotFound, "игра не найдена")
	}
	err := g.Surrender()
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	err = repository.Games.Save()
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return gameToProto(g), nil
}

func (s *server) DeleteGame(ctx context.Context, req *chess_api.IdRequest) (*chess_api.Empty, error) {
	mu.Lock()
	defer mu.Unlock()

	if !repository.Games.Delete(int(req.Id)) {
		return nil, status.Error(codes.NotFound, "игра не найдена")
	}
	err := repository.Games.Save()
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return &chess_api.Empty{}, nil
}

// ходы

func (s *server) CreateMove(ctx context.Context, req *chess_api.CreateMoveRequest) (*chess_api.Move, error) {
	fromRow, fromCol, ok1 := board.ParseSquare(strings.ToUpper(req.From))
	toRow, toCol, ok2 := board.ParseSquare(strings.ToUpper(req.To))
	if !ok1 || !ok2 {
		return nil, status.Error(codes.InvalidArgument, "from и to нужно писать так: E2, E4")
	}

	mu.Lock()
	defer mu.Unlock()

	g := repository.Games.Get(int(req.GameId))
	if g == nil {
		return nil, status.Error(codes.NotFound, "игра не найдена")
	}
	m, err := g.MakeMove(repository.Moves.NextID(), fromRow, fromCol, toRow, toCol)
	if err != nil {
		return nil, status.Error(codes.FailedPrecondition, err.Error())
	}
	err = repository.Moves.Add(m)
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	// игру тоже сохраняем
	err = repository.Games.Save()
	if err != nil {
		return nil, status.Error(codes.Internal, "ошибка сохранения")
	}
	return moveToProto(m), nil
}

func (s *server) GetMove(ctx context.Context, req *chess_api.IdRequest) (*chess_api.Move, error) {
	mu.Lock()
	defer mu.Unlock()

	m := repository.Moves.Get(int(req.Id))
	if m == nil {
		return nil, status.Error(codes.NotFound, "ход не найден")
	}
	return moveToProto(m), nil
}

func (s *server) ListMoves(ctx context.Context, req *chess_api.Empty) (*chess_api.MoveList, error) {
	mu.Lock()
	defer mu.Unlock()

	list := &chess_api.MoveList{}
	for _, m := range repository.Moves.Items {
		list.Moves = append(list.Moves, moveToProto(m))
	}
	return list, nil
}

func main() {
	err := repository.LoadAll()
	if err != nil {
		fmt.Println("ошибка загрузки данных:", err)
		return
	}
	fmt.Println("загружено игроков:", len(repository.Players.Items))
	fmt.Println("загружено игр:", len(repository.Games.Items))
	fmt.Println("загружено ходов:", len(repository.Moves.Items))

	lis, err := net.Listen("tcp", "localhost:5001")
	if err != nil {
		fmt.Println("не получилось открыть порт:", err)
		return
	}

	s := grpc.NewServer()
	chess_api.RegisterChessServiceServer(s, &server{})
	// для Postman
	reflection.Register(s)

	go func() {
		fmt.Println("grpc сервер запущен на localhost:5001")
		err := s.Serve(lis)
		if err != nil {
			fmt.Println("ошибка сервера:", err)
			os.Exit(1)
		}
	}()

	// ждём Ctrl+C
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()

	fmt.Println("останавливаем сервер")
	s.GracefulStop()
}
