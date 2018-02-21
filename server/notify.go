package server

import (
	"beast/aio"
	"beast/global"
	"errors"
	"log"
	"syscall"
)

func (s *IoThread) handleAcceptEvent(socketInfo *SocketInfo) error {
	log.Println("HandleAcceptEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	err := syscall.SetNonblock(socketInfo.Fd, true)
	if err != nil {
		log.Println("IoThread HandleAccept SetNonblock failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	var flag = int(1)
	err = syscall.SetsockoptInt(socketInfo.Fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, flag)
	if err != nil {
		log.Println("TcpServer Set TCP_NODELAY failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	err = aio.Poller(s.EpollFd).Add(socketInfo.Fd, aio.In|aio.Err)
	if err != nil {
		log.Println("IoThread HandleAccept EpollAddFd failed,err=%s", err.Error())
		syscall.Close(socketInfo.Fd)
		return err
	}

	connInfo := &ConnInfo{
		SInfo: socketInfo,
		T:     s,
	}
	s.Owner.ConnList[socketInfo.Fd] = connInfo
	return nil
}

func (s *IoThread) handleWriteEvent(socketInfo *SocketInfo) error {
	log.Println("HandleWriteEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !s.CheckSocketInfo(socketInfo) {
		log.Println("HandleWriteEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}
	err := aio.Poller(s.EpollFd).Add(socketInfo.Fd, aio.In|aio.Out|aio.Err)
	if err != nil {
		log.Println("HandleWriteEvent EpollModFd failed,fd=%d,err=%s", socketInfo.Fd, err.Error())
		s.closeConn(socketInfo.Fd)
		return errors.New("sendMsg EpollModFd failed")
	}
	return nil
}

func (s *IoThread) handleCloseEvent(socketInfo *SocketInfo) error {
	log.Println("HandleCloseEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !s.CheckSocketInfo(socketInfo) {
		log.Println("HandleCloseEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}

	s.closeConn(socketInfo.Fd)
	return nil
}

func (s *IoThread) handleNotify() {
	log.Println("NotifyList len:%d", len(s.NotifyList))
	if len(s.NotifyList) != 0 {
		s.NotifyMutex.Lock()
		var nl = s.NotifyList
		s.NotifyList = []*NotifyEvent{} //置空
		s.NotifyMutex.Unlock()
		for _, v := range nl {
			if v.Type == global.EVENT_ACCEPT {
				s.handleAcceptEvent(v.Info)
			} else if v.Type == global.EVENT_WRITE {
				s.handleWriteEvent(v.Info)
			} else if v.Type == global.EVENT_CLOSE {
				s.handleCloseEvent(v.Info)
			}
		}
	}
}

//外部通知io线程，chan性能不强，这里用 mutex+slice 来代替
func (s *IoThread) Notify(_type int, info *SocketInfo) error {
	e := &NotifyEvent{
		Type: _type,
		Info: info,
	}
	s.NotifyMutex.Lock()
	s.NotifyList = append(s.NotifyList, e)
	s.NotifyMutex.Unlock()
	_, err := syscall.Write(s.NotifyFdW, s.NotifyWriteBytes[:])
	if err != nil {
		log.Println("IoThread Notify failed,err:%s", err.Error())
		return err
	}
	return nil
}
