package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"HW_12/pkg/chess_api"
)

const help = `команды:
  players                  - список игроков
  player 1                 - игрок по id
  addplayer Аня white      - создать игрока
  renameplayer 1 Анна      - поменять имя игрока
  delplayer 1              - удалить игрока
  games                    - список игр
  game 1                   - игра по id (с доской)
  addgame 1 2              - создать игру (id белого, id чёрного)
  surrender 1              - сдаться в игре
  delgame 1                - удалить игру
  move 1 E2 E4             - сделать ход в игре 1
  moves                    - список ходов
  getmove 1                - ход по id
  help                     - показать команды
  выход`

func printPlayer(p *chess_api.Player) {
	fmt.Printf("игрок #%d: %s (%s)\n", p.Id, p.Name, p.Color)
}

func printGame(g *chess_api.Game) {
	fmt.Printf("игра #%d: %s vs %s, ходов: %d\n", g.Id, g.White.Name, g.Black.Name, g.MovesCount)
	fmt.Println("  A B C D E F G H")
	for i, row := range g.Board {
		fmt.Print(8-i, " ")
		for _, ch := range row {
			fmt.Print(string(ch), " ")
		}
		fmt.Println()
	}
	if g.Status == "active" {
		fmt.Println("ходят:", g.Turn)
	} else {
		fmt.Println("игра окончена, победили:", g.Winner)
	}
}

func printMove(m *chess_api.Move) {
	fmt.Printf("ход #%d в игре %d: %s\n", m.Id, m.GameId, m.Notation)
}

// строка в число
func toID(s string) int32 {
	id, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return int32(id)
}

func main() {
	conn, err := grpc.NewClient("localhost:5001", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		fmt.Println("не получилось подключиться:", err)
		return
	}
	defer conn.Close()

	client := chess_api.NewChessServiceClient(conn)
	ctx := context.Background()

	fmt.Println(help)
	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		parts := strings.Fields(scanner.Text())
		if len(parts) == 0 {
			continue
		}

		switch parts[0] {
		case "выход":
			return

		case "help":
			fmt.Println(help)

		case "players":
			list, err := client.ListPlayers(ctx, &chess_api.Empty{})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			if len(list.Players) == 0 {
				fmt.Println("игроков нет")
			}
			for _, p := range list.Players {
				printPlayer(p)
			}

		case "player":
			if len(parts) != 2 {
				fmt.Println("нужно так: player 1")
				continue
			}
			p, err := client.GetPlayer(ctx, &chess_api.IdRequest{Id: toID(parts[1])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printPlayer(p)

		case "addplayer":
			if len(parts) != 3 {
				fmt.Println("нужно так: addplayer Аня white")
				continue
			}
			p, err := client.CreatePlayer(ctx, &chess_api.CreatePlayerRequest{Name: parts[1], Color: parts[2]})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			fmt.Print("создан ")
			printPlayer(p)

		case "renameplayer":
			if len(parts) != 3 {
				fmt.Println("нужно так: renameplayer 1 Анна")
				continue
			}
			p, err := client.UpdatePlayer(ctx, &chess_api.UpdatePlayerRequest{Id: toID(parts[1]), Name: parts[2]})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			fmt.Print("изменён ")
			printPlayer(p)

		case "delplayer":
			if len(parts) != 2 {
				fmt.Println("нужно так: delplayer 1")
				continue
			}
			_, err := client.DeletePlayer(ctx, &chess_api.IdRequest{Id: toID(parts[1])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			fmt.Println("игрок удалён")

		case "games":
			list, err := client.ListGames(ctx, &chess_api.Empty{})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			if len(list.Games) == 0 {
				fmt.Println("игр нет")
			}
			for _, g := range list.Games {
				fmt.Printf("игра #%d: %s vs %s, %s\n", g.Id, g.White.Name, g.Black.Name, g.Status)
			}

		case "game":
			if len(parts) != 2 {
				fmt.Println("нужно так: game 1")
				continue
			}
			g, err := client.GetGame(ctx, &chess_api.IdRequest{Id: toID(parts[1])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printGame(g)

		case "addgame":
			if len(parts) != 3 {
				fmt.Println("нужно так: addgame 1 2")
				continue
			}
			g, err := client.CreateGame(ctx, &chess_api.CreateGameRequest{WhiteId: toID(parts[1]), BlackId: toID(parts[2])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printGame(g)

		case "surrender":
			if len(parts) != 2 {
				fmt.Println("нужно так: surrender 1")
				continue
			}
			g, err := client.UpdateGame(ctx, &chess_api.UpdateGameRequest{Id: toID(parts[1]), Action: "surrender"})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printGame(g)

		case "delgame":
			if len(parts) != 2 {
				fmt.Println("нужно так: delgame 1")
				continue
			}
			_, err := client.DeleteGame(ctx, &chess_api.IdRequest{Id: toID(parts[1])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			fmt.Println("игра удалена")

		case "move":
			if len(parts) != 4 {
				fmt.Println("нужно так: move 1 E2 E4")
				continue
			}
			m, err := client.CreateMove(ctx, &chess_api.CreateMoveRequest{GameId: toID(parts[1]), From: parts[2], To: parts[3]})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printMove(m)

		case "moves":
			list, err := client.ListMoves(ctx, &chess_api.Empty{})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			if len(list.Moves) == 0 {
				fmt.Println("ходов нет")
			}
			for _, m := range list.Moves {
				printMove(m)
			}

		case "getmove":
			if len(parts) != 2 {
				fmt.Println("нужно так: getmove 1")
				continue
			}
			m, err := client.GetMove(ctx, &chess_api.IdRequest{Id: toID(parts[1])})
			if err != nil {
				fmt.Println("ошибка:", err)
				continue
			}
			printMove(m)

		default:
			fmt.Println("неизвестная команда, напишите help")
		}
	}
}
