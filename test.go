package main

import (
	"fmt"
)

func main() {
	fmt.Println("begin")
	a := make([]byte, 1024)
	b := []byte{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}

	length := 5
	fmt.Println(b[length:])
	copy(a, b)
	fmt.Println(a)

	fmt.Println("end")

}
