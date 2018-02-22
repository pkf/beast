package protocol

import (
	. "beast/server"
	"beast/util"
	"bytes"
	"encoding/base64"
	_ "log"
	"net/url"
	"regexp"
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

//see https://github.com/walkor/Workerman/blob/master/Protocols/Websocket.php
func dealHandshake(msg []byte, c *ConnInfo) int {
	buffer := string(msg)
	// HTTP protocol.
	if 0 == strings.Index(buffer, "GET") {
		// Find \r\n\r\n.
		heder_end_pos := strings.Index(buffer, "\r\n\r\n")
		if heder_end_pos < 0 {
			return 0
		}
		header_length := heder_end_pos + 4

		// Get Sec-WebSocket-Key.
		Sec_WebSocket_Key := ""
		var match = regexp.MustCompile("Sec-WebSocket-Key: *(.*?)\r\n")
		tmp := match.FindAllStringSubmatch(buffer, -1)
		if nil != tmp {
			Sec_WebSocket_Key = tmp[0][1]

		} else {
			msg := "HTTP/1.1 400 Bad Request\r\n\r\n<b>400 Bad Request</b><br>Sec-WebSocket-Key not found.<br>This is a WebSocket service and can not be accessed via HTTP."
			c.AsynSendMsg([]byte(msg))
			c.AsynClose()
			return 0
		}

		// Calculation websocket key.
		new_key := base64.StdEncoding.EncodeToString([]byte(util.Sha1(Sec_WebSocket_Key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")))
		buf := bytes.Buffer{}
		// Handshake response data.
		buf.WriteString("HTTP/1.1 101 Switching Protocols\r\n")
		buf.WriteString("Upgrade: websocket\r\n")
		buf.WriteString("Sec-WebSocket-Version: 13\r\n")
		buf.WriteString("Connection: Upgrade\r\n")
		buf.WriteString("Server: beast/1.0\r\n")
		buf.WriteString("Sec-WebSocket-Accept: " + new_key + "\r\n\r\n")
		handshake_message := buf.String()
		// Mark handshake complete..
		//connection->websocketHandshake = true;

		// Websocket data buffer.
		//connection->websocketDataBuffer = '';

		// Current websocket frame length.
		//connection->websocketCurrentFrameLength = 0;

		// Current websocket frame data.
		//connection->websocketCurrentFrameBuffer = '';

		// Consume handshake data.
		//connection->consumeRecvBuffer(header_length);

		// Send handshake response.
		c.AsynSendMsg([]byte(handshake_message))

		// There are data waiting to be sent.
		//if (!empty(connection->tmpWebsocketData)) {
		//    connection->send(connection->tmpWebsocketData, true);
		//    connection->tmpWebsocketData = '';
		//}
		// blob or arraybuffer
		//if (empty(connection->websocketType)) {
		//    connection->websocketType = static::BINARY_TYPE_BLOB;
		//}
		// Try to emit onWebSocketConnect callback.
		if len(buffer) > header_length {
			//return input(substr(buffer, header_length), c);
		}
		return 0
	} else if 0 == strings.Index(buffer, "<polic") {
		// Is flash policy-file-request.
		policy_xml := "<?xml version=\"1.0\"?><cross-domain-policy><site-control permitted-cross-domain-policies=\"all\"/><allow-access-from domain=\"*\" to-ports=\"*\"/></cross-domain-policy>\\0"
		c.AsynSendMsg([]byte(policy_xml))
		//connection->consumeRecvBuffer(strlen(buffer));
		return 0
	}
	// Bad websocket handshake request.
	buf := "HTTP/1.1 400 Bad Request\r\n\r\n<b>400 Bad Request</b><br>Invalid handshake data for websocket."
	c.AsynSendMsg([]byte(buf))
	c.AsynClose()
	return 0
}

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
