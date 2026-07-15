package main

import "fmt"

func CreateChessboard(size int) {
	fmt.Print("   ")
	for col := 0; col < size; col++ {
		fmt.Printf("%c ", 'a'+col)
	}
	fmt.Println()

	for line := 0; line < size; line++ {
		fmt.Printf("%d  ", line+1)
		for col := 0; col < size; col++ {
			if (line+col) % 2 == 0 {
				fmt.Print("# ")
			} else {
				fmt.Print("  ")
			}
		}
		fmt.Println()
	}
}

func main(){
	size := 8
	CreateChessboard(size)
}