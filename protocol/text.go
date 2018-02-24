package protocol

import (
	. "beast/server"
	"log"
	"strings"
)

type TextParser struct {
}

func (p *TextParser) Name() string {
	return "text"
}

func (p *TextParser) Unpack(msg []byte, c *ConnInfo) (ok bool, packlen int) {
	s := string(msg)
	index := strings.Index(s, "\n")
	if index == -1 {
		log.Println("HttpParser not find end")
		return true, 0
	} else {
		log.Println("HttpParser find end")
		return true, index + 1
	}
}

func (p *TextParser) HandlePack(msg []byte, c *ConnInfo) (ok bool) {
	//c.SynSendMsg(msg)
	c.AsynSendMsg(msg)
	//c.AsynClose()
	//c.SynClose()
	return true
}
