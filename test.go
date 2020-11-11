package main

import (
	"fmt"
	"os"
	"strconv"
)

func main() {
	var n int
	var flag bool = false
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
	// fmt.Println(n, reflect.TypeOf(n))
	if n < 0 {
		flag = true
		n = -n
	}
	for n > 0 {
		if n&1 == 1 {
			res++
		}
		n >>= 1
	}
	if flag == true {
		res = 64 - res + 1
	}
	fmt.Println(res)
}
