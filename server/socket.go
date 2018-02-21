package server

import (
	"bytes"
	"sync"
	"syscall"
	"time"
)

func NewSocketInfo(fd int, id uint64, addr syscall.Sockaddr) *SocketInfo {
	return &SocketInfo{
		Fd:             fd,
		Id:             id,
		Addr:           addr,
		ReadBuffer:     bytes.NewBuffer(make([]byte, 0)),
		WriteBuffer:    bytes.NewBuffer(make([]byte, 0)),
		WriteMutex:     sync.Mutex{},
		LastAccessTime: time.Now().Unix(),
	}
}

func (this *SocketInfo) AddMsgToWriteBuffer(msg []byte) bool {
	this.WriteMutex.Lock()
	nWrite, err := this.WriteBuffer.Write(msg)
	this.WriteMutex.Unlock()
	if nWrite != len(msg) || err != nil {
		log.Infof("SocketInfo AddMsgToWriteBuffer Write Buffer failed,s=%+v", this)
		//this.closeConn(fd)
		return false
	}
	//log.InfoF("AddMsgToWriteBuffer msg:%#v",msg)
	return true
}

func (this *SocketInfo) UpdateAccessTime() bool {
	this.LastAccessTime = time.Now().Unix()
	return true
}
