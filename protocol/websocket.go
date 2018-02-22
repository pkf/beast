package protocol

import (
	_ "beast/server"
	_ "log"
	"net/url"
	"strings"
)

const BINARY_TYPE_BLOB = 0x81
const BINARY_TYPE_ARRAYBUFFER = 0x82

type WebSocketParser struct {
}

/*
func (p *WebSocketParser) Unpack(msg []byte, c *ConnInfo) (ok bool, packlen int) {
	//logging.Debug("HttpParser Unpack,msg:%#v",msg)
	s := string(msg)
	index := strings.Index(s, "\r\n\r\n")
	if index == -1 {
		log.Println("HttpParser not find end")
		return true, 0
	} else {
		log.Println("HttpParser find end")
		return true, index + 4
	}
}

func (p *WebSocketParser) HandlePack(msg []byte, c *ConnInfo) (ok bool) {
	//logging.Debug("HttpParser HandlePack,msg:%s",string(msg))
	r := "HTTP/1.1 200 OK\r\nDate: Tue, 18 Jul 2017 09:49:30 GMT\r\nContent-Length: 4\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nabcd"
	c.SynSendMsg([]byte(r))
	c.SynClose()
	//c.Close()
	return true
}
*/

func parseHttpHeader(content string) (server map[string]string, cookie, get map[string][]string) {
	server = make(map[string]string)
	cookie = make(map[string][]string)
	get = make(map[string][]string)

	lines := strings.Split(content, "\r\n\r\n")
	httpHeader := lines[0]
	headerData := strings.Split(httpHeader, "\r\n")
	tmp := strings.Split(headerData[0], " ")
	server["REQUEST_METHOD"] = strings.TrimSpace(tmp[0])
	server["REQUEST_URI"] = strings.TrimSpace(tmp[1])
	server["SERVER_PROTOCOL"] = strings.TrimSpace(tmp[2])

	for i, c := range headerData {
		if i == 0 {
			continue
		}
		if len(c) == 0 {
			continue
		}
		tmp = strings.Split(c, ":")
		key := strings.Replace(strings.ToUpper(tmp[0]), "-", "_", -1)
		value := strings.TrimSpace(tmp[1])
		server["HTTP_"+key] = value

		switch key {
		case "HOST":
			tmp = strings.Split(value, ":")
			server["SERVER_NAME"] = tmp[0]
			if len(tmp[0]) > 0 {
				server["SERVER_PORT"] = tmp[1]
			}
		case "COOKIE":
			cookie, _ = url.ParseQuery(strings.Replace(server["HTTP_COOKIE"], ";", "&", -1))

		}

		uri, e := url.ParseRequestURI(server["REQUEST_URI"])

		if e == nil && len(uri.RawQuery) > 0 {
			get, _ = url.ParseQuery(uri.RawQuery)
			server["QUERY_STRING"] = strings.TrimSpace(uri.RawQuery)
		} else {
			server["QUERY_STRING"] = ""
		}

	}
	return server, cookie, get

}
