package main

import (
	"fmt"
	"HW_5/internal/model/board"
	"HW_5/internal/model/game"
	"HW_5/internal/model/move"
	"HW_5/internal/model/piece"
	"HW_5/internal/model/player"
)

func CreateChessboard(b *board.Board, player1, player2 string) {
	fmt.Printf("   %s\n", player1)
	fmt.Print("   ")
	for col := 0; col < b.Size(); col++ {
		fmt.Printf("%c ", 'A'+col)
	}
	fmt.Println()

	for row := 0; row < b.Size(); row++ {
		fmt.Printf("%d  ", row+1)
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
	fmt.Printf("   %s\n", player2)
}

func main() {
	var size int
	var player1 string
	var player2 string

	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)
	fmt.Print("Введите имя игрока №1: ")
	fmt.Scan(&player1)
	fmt.Print("Введите имя игрока №2: ")
	fmt.Scan(&player2)

	whitePlayer := player.NewPlayer(1, player1, player.White)
	blackPlayer := player.NewPlayer(2, player2, player.Black)

	fmt.Printf("Игрок 1: %s (цвет: %s)\n", whitePlayer.Name(), whitePlayer.Color())
	fmt.Printf("Игрок 2: %s (цвет: %s)\n", blackPlayer.Name(), blackPlayer.Color())

	b := board.NewBoard(size)

	pawnWhite := piece.NewPiece(piece.Pawn, "white")
	b.SetPiece(6, 4, pawnWhite)

	pawnBlack := piece.NewPiece(piece.Pawn, "black")
	b.SetPiece(1, 4, pawnBlack)

	rookWhite := piece.NewPiece(piece.Rook, "white")
	b.SetPiece(7, 0, rookWhite)

	rookBlack := piece.NewPiece(piece.Rook, "black")
	b.SetPiece(0, 0, rookBlack)

	gameInstance := game.NewGame(1, whitePlayer, blackPlayer)

	fmt.Printf("\nИгра #%d: %s vs %s\n", gameInstance.ID(), whitePlayer.Name(), blackPlayer.Name())
	fmt.Printf("Статус игры: %s\n", gameInstance.Status())
	fmt.Printf("Ход игрока: %s\n", gameInstance.CurrentTurn().Name())

	fmt.Println("\nИгровая доска:")
	CreateChessboard(b, whitePlayer.Name(), blackPlayer.Name())

	move1 := move.NewMove(1, 6, 4, 4, 4, pawnWhite)
	gameInstance.MakeMove(move1)

	fmt.Printf("Статус игры: %s\n", gameInstance.Status())
	fmt.Printf("Ход игрока: %s\n", gameInstance.CurrentTurn().Name())
	fmt.Printf("Всего ходов: %d\n", len(gameInstance.Moves()))
}
