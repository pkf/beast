package util

import (
	"fmt"
	"testing"
)

func TestPack(t *testing.T) {
	a := PackxxxxN(97)
	b := PackH("6578616d706c65206865782064617461")
	fmt.Println(a, b)
}
