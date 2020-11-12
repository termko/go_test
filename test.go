package main

import (
	"fmt"
	"os"
	"strconv"
)

func firstTask() {
	var n int
	var res int = 0

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s n\n", os.Args[0])
		return
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	for i := 0; i < 64; i++ {
		if (n>>i)&1 == 1 {
			res++
		}
	}
	fmt.Println(res)
}

func digitSum(n int) int {
	var ret int = 0
	for n > 0 {
		ret += n % 10
		n /= 10
	}
	return ret
}

func secondTask() {
	var n int
	var n1 int = 0
	var n2 int = 1
	var tmp int

	if len(os.Args) != 2 {
		fmt.Printf("Usage: %s n\n", os.Args[0])
		return
	}
	n, err := strconv.Atoi(os.Args[1])
	if err != nil {
		fmt.Println(err)
		return
	}
	if n < 0 {
		fmt.Println("-1")
		return
	}
	for n > 0 {
		tmp = n2
		n2 = digitSum(n1) + digitSum(n2)
		n1 = tmp
		fmt.Println(n2)
		n--
	}
}

func main() {
	firstTask()
	secondTask()
}
