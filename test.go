package main

import (
	"fmt"
)

func main() {
	fmt.Println("begin")
	a := make([]byte, 100)
	b := []byte{1, 2, 3, 4, 5, 6, 7, 8, 9}

	//length := 5
	//fmt.Println(b[length:])
	copy(a, b)
	//fmt.Println(a)

	//a = append(a, b...)
	//fmt.Println(a)

	length2 := len(a)
	a = a[0:length2]
	copy(a, b)
	fmt.Println(a, length2)

	a = append(a, b...)
	fmt.Println(a)

	fmt.Println("end")
}
