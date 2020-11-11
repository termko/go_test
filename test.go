package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	var n int
	var res int = 0

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s n\n", os.Args[0])
		os.Exit(1)
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for i := 0; i < 64; i++ {
		if (n>>i)&1 == 1 {
			res++
		}
	}
	fmt.Println(res)
}
