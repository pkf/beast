package protocol

import (
	"beast/util"
	"fmt"
	"testing"
)

func TestWs(t *testing.T) {
	a := util.FileGetContents("../config/ws_test.txt")
	//fmt.Printf("packet: %s", a)

	server, cookie, get := ParseHttpHeader(string(a))
	fmt.Printf("server: %v\n", server)
	fmt.Printf("cookie: %v\n", cookie)
	fmt.Printf("get: %v\n", get)

	dealHandshake(a, nil)

}
