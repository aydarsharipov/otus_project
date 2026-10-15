package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"time"

	"HW_6/internal/model/game"
	"HW_6/internal/repository"
	"HW_6/internal/service"
)

// Escape-последовательность терминала: переместить курсор в начало экрана
// и очистить экран. Новое состояние рисуется поверх старого.
const clearScreen = "\033[H\033[2J"

// Интервал, с которым main() вызывает service.GenerateEntities.
const generateInterval = 3 * time.Second

var lastGenerate = time.Now()

// generateIfNeeded вызывает сервис, если с прошлого вызова прошёл интервал.
func generateIfNeeded() {
	if time.Since(lastGenerate) < generateInterval {
		return
	}
	if err := service.GenerateEntities(); err != nil {
		fmt.Println("Ошибка:", err)
	}
	lastGenerate = time.Now()
}

// squareName переводит индексы доски в название клетки, например (6, 4) -> "E2".
func squareName(row, col int) string {
	return fmt.Sprintf("%c%d", 'A'+col, game.BoardSize-row)
}

// parseSquare переводит название клетки, например "E2", в индексы доски.
func parseSquare(s string) (row, col int, ok bool) {
	if len(s) != 2 {
		return 0, 0, false
	}
	col = int(s[0] - 'A')
	rank := int(s[1] - '0')
	if col < 0 || col >= game.BoardSize || rank < 1 || rank > game.BoardSize {
		return 0, 0, false
	}
	return game.BoardSize - rank, col, true
}

// render очищает экран и рисует доску и состояние игры.
func render(g *game.Game, message string) {
	fmt.Print(clearScreen)

	fmt.Printf("Игра #%d: %s (белые) vs %s (чёрные)\n\n",
		g.ID(), g.WhitePlayer().Name(), g.BlackPlayer().Name())

	b := g.Board()
	fmt.Println("   A B C D E F G H")
	for row := 0; row < b.Size(); row++ {
		fmt.Printf("%d  ", b.Size()-row)
		for col := 0; col < b.Size(); col++ {
			cell := b.GetCell(row, col)
			if cell.Piece() != nil {
				fmt.Printf("%c ", cell.Piece().Symbol())
			} else if (row+col)%2 == 0 {
				fmt.Print("# ")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
	fmt.Println("   A B C D E F G H")
	fmt.Println()

	moves := g.Moves()
	fmt.Println("Всего ходов:", len(moves))
	if len(moves) > 0 {
		last := moves[len(moves)-1]
		fmt.Printf("Последний ход: %c %s-%s\n", last.Piece().Symbol(),
			squareName(last.FromRow(), last.FromCol()), squareName(last.ToRow(), last.ToCol()))
	}

	if g.Status() == game.Completed {
		fmt.Printf("Игра окончена (%s). Победитель: %s!\n", g.EndReason(), g.Winner().Name())
	} else {
		fmt.Printf("Ходит: %s (%s)\n", g.CurrentTurn().Name(), g.CurrentTurn().Color())
	}

	stats := repository.Stats()
	fmt.Printf("Репозиторий: players=%d pieces=%d moves=%d boards=%d games=%d\n",
		stats["player"], stats["piece"], stats["move"], stats["board"], stats["game"])

	fmt.Println("Команды: E2 E4 | сдался | авто N | выход")
	fmt.Println(message)
}

// makeMove разбирает ход вида "E2 E4" и выполняет его.
func makeMove(g *game.Game, input string) string {
	squares := strings.ToUpper(strings.ReplaceAll(input, " ", ""))
	if len(squares) != 4 {
		return "Неизвестная команда: " + input
	}
	fromRow, fromCol, ok1 := parseSquare(squares[:2])
	toRow, toCol, ok2 := parseSquare(squares[2:])
	if !ok1 || !ok2 {
		return "Неверный формат хода, пример: E2 E4"
	}
	if _, err := g.MakeMove(fromRow, fromCol, toRow, toCol); err != nil {
		return "Ход невозможен: " + err.Error()
	}
	return "Ход выполнен"
}

// autoMove делает один случайный ход за текущего игрока.
func autoMove(g *game.Game) {
	// Собираем в слайс все возможные ходы: [fromRow, fromCol, toRow, toCol].
	possible := make([][4]int, 0)
	for fromRow := 0; fromRow < game.BoardSize; fromRow++ {
		for fromCol := 0; fromCol < game.BoardSize; fromCol++ {
			for toRow := 0; toRow < game.BoardSize; toRow++ {
				for toCol := 0; toCol < game.BoardSize; toCol++ {
					if g.CanMove(fromRow, fromCol, toRow, toCol) == nil {
						possible = append(possible, [4]int{fromRow, fromCol, toRow, toCol})
					}
				}
			}
		}
	}
	if len(possible) == 0 {
		return
	}

	m := possible[rand.Intn(len(possible))]
	g.MakeMove(m[0], m[1], m[2], m[3])
}

// autoMoves делает n случайных ходов со случайной задержкой 2-4 секунды.
func autoMoves(g *game.Game, n int) {
	for i := 1; i <= n; i++ {
		if g.Status() != game.Active {
			return
		}
		render(g, fmt.Sprintf("Автоход %d из %d: %s думает...", i, n, g.CurrentTurn().Name()))

		delay := 2 + rand.Intn(3) // 2, 3 или 4 секунды
		time.Sleep(time.Duration(delay) * time.Second)

		autoMove(g)
		generateIfNeeded()
	}
}

// handleCommand выполняет команду игрока и возвращает сообщение для экрана.
func handleCommand(g *game.Game, input string) string {
	command := strings.ToUpper(input)

	if command == "СДАЛСЯ" {
		g.Surrender()
		return ""
	}

	if strings.HasPrefix(command, "АВТО") {
		countText := strings.TrimSpace(strings.TrimPrefix(command, "АВТО"))
		count, err := strconv.Atoi(countText)
		if err != nil || count < 1 || count > 100 {
			return "Укажите число ходов от 1 до 100, например: авто 3"
		}
		autoMoves(g, count)
		return "Автоход завершён"
	}

	return makeMove(g, input)
}

func main() {
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Print("Введите имя игрока №1 (белые): ")
	scanner.Scan()
	player1 := strings.TrimSpace(scanner.Text())

	fmt.Print("Введите имя игрока №2 (чёрные): ")
	scanner.Scan()
	player2 := strings.TrimSpace(scanner.Text())

	g := service.NewGame(player1, player2)
	message := "Введите ход"

	for {
		generateIfNeeded()
		render(g, message)
		if g.Status() == game.Completed {
			return
		}

		fmt.Print("> ")
		if !scanner.Scan() {
			return
		}
		input := strings.TrimSpace(scanner.Text())
		if input == "выход" {
			return
		}
		message = handleCommand(g, input)
	}
}
