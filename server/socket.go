package server

import (
	"bytes"
	"log"
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

func (s *SocketInfo) AddMsgToWriteBuffer(msg []byte) bool {
	s.WriteMutex.Lock()
	nWrite, err := s.WriteBuffer.Write(msg)
	s.WriteMutex.Unlock()
	if nWrite != len(msg) || err != nil {
		log.Println("SocketInfo AddMsgToWriteBuffer Write Buffer failed,s=%+v", s)
		//s.closeConn(fd)
		return false
	}
	//logging.Debug("AddMsgToWriteBuffer msg:%#v",msg)
	return true
}

//修改socket的上次活跃时间
func (s *SocketInfo) UpdateAccessTime() bool {
	s.LastAccessTime = time.Now().Unix()
	return true
}
