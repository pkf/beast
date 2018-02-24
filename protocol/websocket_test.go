package protocol

import (
	"beast/util"
	"fmt"
	"testing"
)

func TestWs(t *testing.T) {
	a := util.FileGetContents("../config/ws_test.txt")
	//fmt.Printf("packet: %s", a)

	server, cookie, get := parseHttpHeader(string(a))
	fmt.Printf("server: %v\n", server)
	fmt.Printf("cookie: %v\n", cookie)
	fmt.Printf("get: %v\n", get)

	//dealHandshake(a, nil)

	fmt.Println(string(rune(126)))

}
