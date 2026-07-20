package main

import "fmt"

func letterLabel(n int) string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	if n < 26 {
		return string(letters[n])
	}
	first := (n) / 26
	second := n % 26
	return string(letters[first-1]) + string(letters[second])
}

func CreateChessboard(size int, player1, player2 string) {
	fmt.Printf("   %s\n", player1)
	fmt.Print("   ")
	for col := 0; col < size; col++ {
		fmt.Printf("%s ", letterLabel(col))
	}
	fmt.Println()

	firstRow := []rune{'♜', '♞', '♝', '♛', '♚', '♝', '♞', '♜'}
	pawnRow := []rune{'♟', '♟', '♟', '♟', '♟', '♟', '♟', '♟'}
	lastPawnRow := []rune{'♙', '♙', '♙', '♙', '♙', '♙', '♙', '♙'}
	lastRow := []rune{'♖', '♘', '♗', '♕', '♔', '♗', '♘', '♖'}

	for line := 0; line < size; line++ {
		fmt.Printf("%s  ", letterLabel(line))
		for col := 0; col < size; col++ {
			piece := rune(' ')
			if line == 0 {
				piece = firstRow[col%len(firstRow)]
			} else if line == 1 {
				piece = pawnRow[col%len(pawnRow)]
			} else if line == size-2 {
				piece = lastPawnRow[col%len(lastPawnRow)]
			} else if line == size-1 {
				piece = lastRow[col%len(lastRow)]
			}

			if piece != ' ' {
				fmt.Printf("%c ", piece)
			} else if (line+col)%2 == 0 {
				fmt.Print("# ")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
	fmt.Printf("   %s\n", player2)
}

func main(){
	var size int
	var player1 string
	var player2 string
	fmt.Print("Введите размер доски: ")
	fmt.Scan(&size)
	fmt.Print("Введите имя игрока №1: ")
	fmt.Scan(&player1)
	fmt.Print("Введите имя игрока №2: ")
	fmt.Scan(&player2)
	CreateChessboard(size, player1, player2)
}