// +build  ignore

package main

import (
	_ "beast/global"
	_ "fmt"
)

const READ_BUFFER_SIZE = 65535
const STATUS_INITIAL = 0
const STATUS_CONNECTING = 1
const STATUS_ESTABLISHED = 2
const STATUS_CLOSING = 4
const STATUS_CLOSED = 8

type ConnInfo struct {
	SocketFd          int
	MaxSendBufferSize int
	RemoteAddress     string
	Status            int
	SendBuffer        []byte
	ReadBuffer        []byte
	Parser            interface{}
	BytesWritten      int
}

func GetInstance(socketFd int, remoteAddress string) *ConnInfo {

	var ret = &ConnInfo{
		SocketFd:          socketFd,
		RemoteAddress:     remoteAddress,
		MaxSendBufferSize: 1024000,
		Status:            STATUS_ESTABLISHED,
		SendBuffer:        make([]byte, 1024),
		ReadBuffer:        make([]byte, 1024),
		Parser:            nil,
		BytesWritten:      0,
	}
	return ret
}

func (this *ConnInfo) Send(sendBuffer []byte, raw bool) bool {
	if this.Status == STATUS_CLOSING || this.Status == STATUS_CLOSED {
		return false
	}

	// Try to call protocol::encode($send_buffer) before sending.
	if false == raw && this.Parser != nil {
		//sendBuffer = Parser::encode(sendBuffer);
		if len(sendBuffer) == 0 {
			return false
		}
	}

	// Attempt to send data directly.
	if len(this.SendBuffer) == 0 {

		//$length = @fwrite($this->_socket, $send_buffer, 8192);
		length := 0
		// send successful.
		if length == len(sendBuffer) {
			this.BytesWritten += length
			return true
		}

		// Send only part of the data.
		if length > 0 {
			tmp := sendBuffer[length:]
			copy(this.SendBuffer, tmp)
			this.BytesWritten += length
		} else {
			//if (!is_resource($this->_socket) || feof($this->_socket)) {
			//self::$statistics['send_fail']++;

			//$this->destroy();
			//return false
			//}
			copy(this.SendBuffer, sendBuffer)
		}
		//Worker::$globalEvent->add($this->_socket, EventInterface::EV_WRITE, array($this, 'baseWrite'));

		return true
	} else {
		this.SendBuffer = append(this.SendBuffer, sendBuffer...)
	}

	return true
}

func (this *ConnInfo) BaseWrite(sendBuffer []byte, raw bool) bool {
	//$len = @fwrite($this->_socket, $this->_sendBuffer, 8192);
	length := 0
	if length == len(this.SendBuffer) {
		this.BytesWritten += length
		//Worker::$globalEvent->del($this->_socket, EventInterface::EV_WRITE);
		this.SendBuffer = this.SendBuffer[0:0]

		if this.Status == STATUS_CLOSING {
			this.Destroy()
		}
		return true
	}
	if length > 0 {
		this.BytesWritten += length
		//$this->_sendBuffer = substr($this->_sendBuffer, $len);
		tmp := sendBuffer[length:]
		copy(this.SendBuffer, tmp)
	} else {
		//self::$statistics['send_fail']++;
		this.Destroy()
	}
	return true
}

func (this *ConnInfo) baseRead(sendBuffer []byte, raw bool) bool {
	return true
}

func (this *ConnInfo) Destroy() bool {
	return true
}
