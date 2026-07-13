package main

import "fmt"

func CreateChessboard() {
	size := 8
	for line := 0; line < size; line++{
		for col := 0; col < size; col++{
			if (line+col) % 2 == 0 {
				fmt.Print("#")
			} else {
				fmt.Print(" ")
			}
		}
		fmt.Println()
	}
}

func main(){
	CreateChessboard()
}