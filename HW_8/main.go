package main

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"HW_8/internal/model"
	"HW_8/internal/model/board"
	"HW_8/internal/model/game"
	"HW_8/internal/repository"
	"HW_8/internal/service"
)

const clearScreen = "\033[H\033[2J"

const (
	generateInterval = 3 * time.Second
	redrawInterval   = time.Second

	maxBoards       = 6
	boardsPerRow    = 3
	columnWidth     = 38
	maxMovesPerGame = 60
	minMoveDelay    = 500 * time.Millisecond
	maxMoveDelay    = 3 * time.Second
)

type app struct {
	ctx       context.Context
	cancel    context.CancelFunc
	svc       *service.Service
	entities  chan model.Entity
	producers sync.WaitGroup

	saverDone chan struct{}
	stopOnce  sync.Once
}

func newApp(parent context.Context) *app {
	ctx, cancel := context.WithCancel(parent)
	a := &app{
		ctx:       ctx,
		cancel:    cancel,
		entities:  make(chan model.Entity, 100),
		saverDone: make(chan struct{}),
	}
	a.svc = service.New(a.entities)

	go func() {
		defer close(a.saverDone)
		repository.Run(a.entities)
	}()
	a.producers.Go(func() {
		a.svc.RunGenerator(a.ctx, generateInterval)
	})
	return a
}

func (a *app) Stop() {
	a.stopOnce.Do(func() {
		a.cancel()
		a.producers.Wait()
		close(a.entities)
		<-a.saverDone
	})
}

func readLines(ctx context.Context) <-chan string {
	lines := make(chan string)
	go func() {
		defer close(lines)
		scanner := bufio.NewScanner(os.Stdin)
		for scanner.Scan() {
			select {
			case <-ctx.Done():
				return
			case lines <- strings.TrimSpace(scanner.Text()):
			}
		}
	}()
	return lines
}

func prompt(ctx context.Context, lines <-chan string, question string) (string, bool) {
	fmt.Print(question)
	select {
	case <-ctx.Done():
		return "", false
	case line, ok := <-lines:
		return line, ok
	}
}

func formatDuration(d time.Duration) string {
	return fmt.Sprintf("%.1fс", d.Seconds())
}

func boardLines(s game.Snapshot) []string {
	lines := []string{
		fmt.Sprintf("Игра #%d: %s vs %s", s.ID, s.WhiteName, s.BlackName),
		"   A B C D E F G H",
	}
	for row := 0; row < game.BoardSize; row++ {
		var sb strings.Builder
		fmt.Fprintf(&sb, "%d  ", game.BoardSize-row)
		for col := 0; col < game.BoardSize; col++ {
			switch {
			case s.Cells[row][col] != 0:
				fmt.Fprintf(&sb, "%c ", s.Cells[row][col])
			case (row+col)%2 == 0:
				sb.WriteString("# ")
			default:
				sb.WriteString("  ")
			}
		}
		lines = append(lines, sb.String())
	}
	lines = append(lines, "   A B C D E F G H")

	last := s.LastMove
	if last == "" {
		last = "-"
	}
	lines = append(lines,
		fmt.Sprintf("Ходов: %d, последний: %s", s.MovesCount, last),
		fmt.Sprintf("Белые  %s: %s (всего %s)", s.WhiteName,
			formatDuration(s.WhiteTime.Last), formatDuration(s.WhiteTime.Total)),
		fmt.Sprintf("Чёрные %s: %s (всего %s)", s.BlackName,
			formatDuration(s.BlackTime.Last), formatDuration(s.BlackTime.Total)),
	)

	switch {
	case s.Status == game.Active:
		lines = append(lines, fmt.Sprintf("Ходит %s, думает %s", s.TurnName, formatDuration(s.Thinking)))
	case s.WinnerName != "":
		lines = append(lines, fmt.Sprintf("Победил %s (%s)", s.WinnerName, s.EndReason))
	default:
		lines = append(lines, fmt.Sprintf("Ничья (%s)", s.EndReason))
	}
	return lines
}

func render(games []*game.Game, footer ...string) {
	var sb strings.Builder
	sb.WriteString(clearScreen)

	for start := 0; start < len(games); start += boardsPerRow {
		end := min(start+boardsPerRow, len(games))
		columns := make([][]string, 0, end-start)
		height := 0
		for _, g := range games[start:end] {
			lines := boardLines(g.Snapshot())
			columns = append(columns, lines)
			height = max(height, len(lines))
		}
		for i := 0; i < height; i++ {
			for _, lines := range columns {
				cell := ""
				if i < len(lines) {
					cell = lines[i]
				}

				fmt.Fprintf(&sb, "%-*s", columnWidth, cell)
			}
			sb.WriteString("\n")
		}
		sb.WriteString("\n")
	}

	for _, line := range footer {
		sb.WriteString(line + "\n")
	}
	fmt.Print(sb.String())
}

func makeMove(a *app, g *game.Game, input string) string {
	squares := strings.ToUpper(strings.ReplaceAll(input, " ", ""))
	if len(squares) != 4 {
		return "Неизвестная команда: " + input
	}
	fromRow, fromCol, ok1 := board.ParseSquare(squares[:2])
	toRow, toCol, ok2 := board.ParseSquare(squares[2:])
	if !ok1 || !ok2 {
		return "Неверный формат хода, пример: E2 E4"
	}
	if err := a.svc.MakeMove(g, fromRow, fromCol, toRow, toCol); err != nil {
		return "Ход невозможен: " + err.Error()
	}
	return "Ход выполнен"
}

func autoMoves(a *app, g *game.Game, n int) string {
	for i := 1; i <= n; i++ {
		if g.Status() != game.Active {
			break
		}
		render([]*game.Game{g}, fmt.Sprintf("Автоход %d из %d: %s думает...", i, n, g.CurrentTurn().Name()))
		if !service.Sleep(a.ctx, service.RandomDelay(2*time.Second, 4*time.Second)) {
			return "Прервано"
		}
		a.svc.AutoMove(g)
	}
	return "Автоход завершён"
}

func handleCommand(a *app, g *game.Game, input string) string {
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
		return autoMoves(a, g, count)
	}

	return makeMove(a, g, input)
}

func playManual(a *app, lines <-chan string) {
	player1, ok := prompt(a.ctx, lines, "Введите имя игрока №1 (белые): ")
	if !ok {
		return
	}
	player2, ok := prompt(a.ctx, lines, "Введите имя игрока №2 (чёрные): ")
	if !ok {
		return
	}

	g := a.svc.NewGame(player1, player2)
	message := "Введите ход"

	for {
		render([]*game.Game{g}, "Команды: E2 E4 | сдался | авто N | выход", message)
		if g.Status() == game.Completed {
			return
		}

		input, ok := prompt(a.ctx, lines, "> ")
		if !ok || input == "выход" {
			return
		}
		message = handleCommand(a, g, input)
	}
}

func playSimulation(a *app, lines <-chan string, boards int) {
	games := make([]*game.Game, boards)
	var playing sync.WaitGroup
	for i := range games {
		g := a.svc.NewGame(fmt.Sprintf("white%d", i+1), fmt.Sprintf("black%d", i+1))
		games[i] = g
		playing.Add(1)
		a.producers.Go(func() {
			defer playing.Done()
			a.svc.RunSimulation(a.ctx, g, maxMovesPerGame, minMoveDelay, maxMoveDelay)
		})
	}

	allDone := make(chan struct{})
	go func() {
		playing.Wait()
		close(allDone)
	}()

	ticker := time.NewTicker(redrawInterval)
	defer ticker.Stop()

	for {
		render(games, "Симуляция. Enter — остановить.")
		select {
		case <-allDone:
			render(games, "Все партии завершены.")
			return
		case <-a.ctx.Done():
			render(games, "Симуляция остановлена.")
			return
		case _, ok := <-lines:
			if !ok {

				lines = nil
				continue
			}
			a.cancel()
		case <-ticker.C:
		}
	}
}

func main() {
	ctx, stopSignals := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stopSignals()

	a := newApp(ctx)
	defer a.Stop()

	lines := readLines(a.ctx)
	boards := 0
	for boards < 1 || boards > maxBoards {
		answer, ok := prompt(ctx, lines, fmt.Sprintf("Сколько досок (1-%d)? ", maxBoards))
		if !ok {
			return
		}
		boards, _ = strconv.Atoi(answer)
	}

	if boards == 1 {
		playManual(a, lines)
	} else {
		playSimulation(a, lines, boards)
	}
}
