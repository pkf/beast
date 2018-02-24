package protocol

import (
	"beast/global"
	. "beast/server"
	"beast/util"
	"bytes"
	"encoding/base64"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const BINARY_TYPE_BLOB = 0x81
const BINARY_TYPE_ARRAYBUFFER = 0x82

type WebSocketParser struct {
}

func (p *WebSocketParser) Name() string {
	return "websocket"
}

func (p *WebSocketParser) Unpack(msg []byte, c *ConnInfo) (ok bool, packlen int) {
	index := input(msg, c)
	return true, index
}

func (p *WebSocketParser) HandlePack(msg []byte, c *ConnInfo) (ok bool) {
	//logging.Debug("HttpParser HandlePack,msg:%s",string(msg))
	r := "HTTP/1.1 200 OK\r\nDate: Tue, 18 Jul 2017 09:49:30 GMT\r\nContent-Length: 4\r\nContent-Type: text/plain; charset=utf-8\r\n\r\nabcd"
	c.SynSendMsg([]byte(r))
	c.SynClose()
	//c.Close()
	return true
}

func input(msg []byte, c *ConnInfo) int {
	buffer := string(msg)
	recv_len := len(buffer)

	// We need more data.
	if recv_len < 2 {
		return 0
	}

	ext, _ := c.Ext.(*WebSocket)
	// Has not yet completed the handshake.
	if ext.WebsocketHandshake != true {
		return dealHandshake(msg, c)
	}

	// Buffer websocket frame data.
	if true {
		/*
		   if (connection->websocketCurrentFrameLength) {
		       // We need more frame data.
		       if (connection->websocketCurrentFrameLength > recv_len) {
		           // Return 0, because it is not clear the full packet length, waiting for the frame of fin=1.
		           return 0;
		       }
		*/
	} else {
		var head_len int
		firstbyte := util.Ord(buffer[0])
		secondbyte := util.Ord(buffer[1])
		data_len := secondbyte & 127
		is_fin_frame := firstbyte >> 7
		masked := secondbyte >> 7
		opcode := firstbyte & 0xf
		switch opcode {
		case 0x0:
			break
		// Blob type.
		case 0x1:
			break
		// Arraybuffer type.
		case 0x2:
			break
		// Close package.
		case 0x8:
			// Try to emit onWebSocketClose callback.
			/*
			   if (isset(connection->onWebSocketClose) || isset(connection->worker->onWebSocketClose)) {
			       try {
			           call_user_func(isset(connection->onWebSocketClose)?connection->onWebSocketClose:connection->worker->onWebSocketClose, connection);
			       } catch (\Exception e) {
			           Worker::log(e);
			           exit(250);
			       } catch (\Error e) {
			           Worker::log(e);
			           exit(250);
			       }
			   } // Close connection.
			   else {
			       connection->close();
			   }
			*/
			return 0
		// Ping package.
		case 0x9:
			// Try to emit onWebSocketPing callback.
			if true {
				/*
				   if (isset(connection->onWebSocketPing) || isset(connection->worker->onWebSocketPing)) {
				       try {
				           call_user_func(isset(connection->onWebSocketPing)?connection->onWebSocketPing:connection->worker->onWebSocketPing, connection);
				       } catch (\Exception e) {
				           Worker::log(e);
				           exit(250);
				       } catch (\Error e) {
				           Worker::log(e);
				           exit(250);
				       }
				   } // Send pong package to client.
				*/
			} else {
				//buf:= (pack('H*', '8a00'), true)
				buf := ""
				c.AsynSendMsg([]byte(buf))
			}

			// Consume data from receive buffer.
			if int(data_len) < 0 {

				if masked > 0 {
					head_len = 6
				} else {
					head_len = 6
				}

				/*
				   connection->consumeRecvBuffer(head_len);
				   if (recv_len > head_len) {
				       return static::input(substr(buffer, head_len), connection);
				   }
				*/
				return 0
			}
			break
		// Pong package.
		case 0xa:
			// Try to emit onWebSocketPong callback.
			/*
			   if (isset(connection->onWebSocketPong) || isset(connection->worker->onWebSocketPong)) {
			       try {
			           call_user_func(isset(connection->onWebSocketPong)?connection->onWebSocketPong:connection->worker->onWebSocketPong, connection);
			       } catch (\Exception e) {
			           Worker::log(e);
			           exit(250);
			       } catch (\Error e) {
			           Worker::log(e);
			           exit(250);
			       }
			   }
			*/
			//  Consume data from receive buffer.
			if int(data_len) < 0 {
				if masked > 0 {
					head_len = 6
				} else {
					head_len = 6
				}
				/*
				   connection->consumeRecvBuffer(head_len);
				   if (recv_len > head_len) {
				       return static::input(substr(buffer, head_len), connection);
				   }*/
				return 0
			}
			break
		// Wrong opcode.
		default:
			//echo "error opcode opcode and close websocket connection. Buffer:" . bin2hex(buffer) . "\n";
			c.AsynClose()
			return 0
		}

		// Calculate packet length.
		head_len = 6
		if data_len == 126 {
			head_len = 8
			if head_len > recv_len {
				return 0
			}
			//pack     = unpack('nn/ntotal_len', buffer);
			//data_len = pack['total_len'];
			_, v := util.Unpacknn(buffer)
			data_len = int32(v)
		} else {
			if data_len == 127 {
				head_len = 14
				if head_len > recv_len {
					return 0
				}
				//arr      = unpack('n/N2c', buffer);
				//data_len = arr['c1']*4294967296 + arr['c2'];
				_, c1, c2 := util.UnpacknN2c(buffer)
				data_len = int32(int(c1)*4294967296 + int(c2))
			}
		}
		current_frame_length := head_len + int(data_len)

		//total_package_size = strlen(connection->websocketDataBuffer) + current_frame_length;
		total_package_size := 0
		if total_package_size > global.MAX_PACK_LEN {
			//echo "error package. package_length=total_package_size\n";
			c.AsynClose()
			return 0
		}

		if int(is_fin_frame) > 0 {
			return current_frame_length
		} else {
			//connection->websocketCurrentFrameLength = current_frame_length;
		}
	}

	// Received just a frame length data.
	/*
	   if (connection->websocketCurrentFrameLength === recv_len) {
	       static::decode(buffer, connection);
	       connection->consumeRecvBuffer(connection->websocketCurrentFrameLength);
	       connection->websocketCurrentFrameLength = 0;
	       return 0;
	   } // The length of the received data is greater than the length of a frame.
	   elseif (connection->websocketCurrentFrameLength < recv_len) {
	       static::decode(substr(buffer, 0, connection->websocketCurrentFrameLength), connection);
	       connection->consumeRecvBuffer(connection->websocketCurrentFrameLength);
	       current_frame_length                    = connection->websocketCurrentFrameLength;
	       connection->websocketCurrentFrameLength = 0;
	       // Continue to read next frame.
	       return static::input(substr(buffer, current_frame_length), connection);
	   } // The length of the received data is less than the length of a frame.
	   else {
	*/
	return 0
	/*
	   }
	*/
}

func encode(msg []byte, c *ConnInfo) string {
	buffer := string(msg)
	length := len(buffer)
	ext, _ := c.Ext.(*WebSocket)
	if ext.WebsocketType == 0 {
		ext.WebsocketType = BINARY_TYPE_BLOB
	}

	first_byte := BINARY_TYPE_BLOB
	var encode_buffer string
	if length <= 125 {
		encode_buffer = string(first_byte) + string(rune(length)) + buffer
	} else {
		if length <= 65535 {
			encode_buffer = string(first_byte) + string(rune(126)) + util.Packn(length) + buffer
		} else {
			encode_buffer = string(first_byte) + string(rune(127)) + util.PackxxxxN(length) + buffer
		}
	}

	//Handshake not completed so temporary buffer websocket data waiting for send.
	if ext.WebsocketHandshake != true {
		if ext.TmpWebsocketData.Len() == 0 {
			ext.TmpWebsocketData.Reset()
		}

		ext.TmpWebsocketData.WriteString(encode_buffer)

		//Return empty string.
		return ""
	}

	return encode_buffer
}

func decode(msg []byte, c *ConnInfo) string {
	buffer := string(msg)
	var masks string
	var data string
	b := []rune(string(buffer[1]))
	len := b[0] & 127
	if len == 126 {
		masks = buffer[4:8]
		data = buffer[8:]
	} else {
		if len == 127 {
			masks = buffer[10:14]
			data = buffer[14:]
		} else {
			masks = buffer[2:6]
			data = buffer[6:]
		}
	}
	buf := bytes.Buffer{}
	for index, v := range data {
		m := []rune(string(masks[index%4]))
		tmp := v ^ m[0]
		buf.WriteString(string(tmp))

	}
	decoded := buf.String()

	ext, _ := c.Ext.(*WebSocket)
	if ext.WebsocketCurrentFrameLength > 0 {
		ext.WebsocketDataBuffer.WriteString(decoded)
		return ext.WebsocketDataBuffer.String()
	} else {
		if ext.WebsocketDataBuffer.Len() > 0 {
			decoded = ext.WebsocketDataBuffer.String() + decoded
			ext.WebsocketDataBuffer.Reset()
		}
		return decoded
	}

	return decoded
}

//see https://github.com/walkor/Workerman/blob/master/Protocols/Websocket.php
func dealHandshake(msg []byte, c *ConnInfo) int {
	buffer := string(msg)
	if 0 == strings.Index(buffer, "GET") {
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
		new_key := base64.URLEncoding.EncodeToString([]byte(util.Sha1(Sec_WebSocket_Key + "258EAFA5-E914-47DA-95CA-C5AB0DC85B11")))
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
		ext, _ := c.Ext.(*WebSocket)

		//util.Log.Infof("websocket  dealHandshake=%v, %v", ext, ext.WebsocketDataBuffer)
		ext.WebsocketHandshake = true

		// Websocket data buffer.
		ext.WebsocketDataBuffer.Reset()

		// Current websocket frame length.
		ext.WebsocketCurrentFrameLength = 0

		// Current websocket frame data.
		ext.WebsocketCurrentFrameBuffer.Reset()

		// Consume handshake data.
		c.ConsumeRecvBuffer(header_length)

		// Send handshake response.
		//fmt.Println(handshake_message)
		c.AsynSendMsg([]byte(handshake_message))

		// There are data waiting to be sent.
		if ext.TmpWebsocketData.Len() > 0 {
			c.AsynSendMsg(ext.TmpWebsocketData.Bytes())
			ext.TmpWebsocketData.Reset()
		}
		// blob or arraybuffer
		if ext.WebsocketType == 0 {
			ext.WebsocketType = BINARY_TYPE_BLOB
		}
		// Try to emit onWebSocketConnect callback.
		if len(buffer) > header_length {
			//return input(substr(buffer, header_length), c);
		}
		return 0
	} else if 0 == strings.Index(buffer, "<polic") {
		// Is flash policy-file-request.
		policy_xml := "<?xml version=\"1.0\"?><cross-domain-policy><site-control permitted-cross-domain-policies=\"all\"/><allow-access-from domain=\"*\" to-ports=\"*\"/></cross-domain-policy>\\0"
		c.AsynSendMsg([]byte(policy_xml))
		c.ConsumeRecvBuffer(len(buffer))
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
	fmt.Printf("%v \n", headerData)
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
			if len(tmp) > 1 {
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
