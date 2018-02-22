package test

import (
	"fmt"
	"regexp"
	"testing"
)

func TestRegexp(t *testing.T) {
	var s = "Sec-WebSocket-Key: asdfghjkl\r\n/i"
	var valid = regexp.MustCompile("Sec-WebSocket-Key: *(.*?)\r\n")
	tmp := valid.FindAllStringSubmatch(s, -1)
	fmt.Printf("%q\n", tmp[0][1])
	//t.Error(tmp[0])
}
