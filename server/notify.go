package server

import (
	"beast/aio"
	"beast/global"
	"errors"
	"syscall"
)

func (this *IoThread) handleAcceptEvent(socketInfo *SocketInfo) error {
	log.Infof("HandleAcceptEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	err := syscall.SetNonblock(socketInfo.Fd, true)
	if err != nil {
		log.Infof("IoThread HandleAccept SetNonblock failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	var flag = int(1)
	err = syscall.SetsockoptInt(socketInfo.Fd, syscall.IPPROTO_TCP, syscall.TCP_NODELAY, flag)
	if err != nil {
		log.Infof("TcpServer Set TCP_NODELAY failed")
		syscall.Close(socketInfo.Fd)
		return err
	}

	err = aio.Poller(this.EpollFd).Add(socketInfo.Fd, aio.In|aio.Err)
	if err != nil {
		log.Infof("IoThread HandleAccept EpollAddFd failed,err=%s", err.Error())
		syscall.Close(socketInfo.Fd)
		return err
	}

	connInfo := &ConnInfo{
		SocketInfo: socketInfo,
		T:          this,
	}
	this.Server.ConnList[socketInfo.Fd] = connInfo
	return nil
}

func (this *IoThread) handleWriteEvent(socketInfo *SocketInfo) error {
	log.Infof("HandleWriteEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !this.CheckSocketInfo(socketInfo) {
		log.Infof("HandleWriteEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}
	err := aio.Poller(this.EpollFd).Add(socketInfo.Fd, aio.In|aio.Out|aio.Err)
	if err != nil {
		log.Infof("HandleWriteEvent EpollModFd failed,fd=%d,err=%s", socketInfo.Fd, err.Error())
		this.closeConn(socketInfo.Fd)
		return errors.New("sendMsg EpollModFd failed")
	}
	return nil
}

func (this *IoThread) handleCloseEvent(socketInfo *SocketInfo) error {
	log.Infof("HandleCloseEvent,fd:%d,id:%d", socketInfo.Fd, socketInfo.Id)
	if !this.CheckSocketInfo(socketInfo) {
		log.Infof("HandleCloseEvent CheckSocketInfo failed")
		return errors.New("CheckSocketInfo failed")
	}

	this.closeConn(socketInfo.Fd)
	return nil
}

func (this *IoThread) handleNotify() {
	log.Infof("NotifyList len:%d", len(this.NotifyList))
	if len(this.NotifyList) != 0 {
		this.NotifyMutex.Lock()
		var nl = this.NotifyList
		this.NotifyList = []*NotifyEvent{}
		this.NotifyMutex.Unlock()
		for _, v := range nl {
			if v.Type == global.EVENT_ACCEPT {
				this.handleAcceptEvent(v.Info)
			} else if v.Type == global.EVENT_WRITE {
				this.handleWriteEvent(v.Info)
			} else if v.Type == global.EVENT_CLOSE {
				this.handleCloseEvent(v.Info)
			}
		}
	}
}

func (this *IoThread) Notify(_type int, info *SocketInfo) error {
	e := &NotifyEvent{
		Type: _type,
		Info: info,
	}
	this.NotifyMutex.Lock()
	this.NotifyList = append(this.NotifyList, e)
	this.NotifyMutex.Unlock()
	_, err := syscall.Write(this.NotifyFdW, this.NotifyWriteBytes[:])
	if err != nil {
		log.Infof("IoThread Notify failed,err:%s", err.Error())
		return err
	}
	return nil
}
