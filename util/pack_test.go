package util

import (
	"fmt"
	"testing"
)

func TestPack(t *testing.T) {
	a := PackxxxxN(97)
	b := PackH("6578616d706c65206865782064617461")
	c, d := Unpacknn("\x04\x00\xa0\x00")
	v, n, m := UnpacknN2c("\x04\x00\x04\x00\x04\x00\x04\x00\xa0\x00")
	fmt.Println(a, b)
	fmt.Println(c, d)
	fmt.Println(v, n, m)
}
