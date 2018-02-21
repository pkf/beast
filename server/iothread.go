package server

import (
	"beast/aio"
	"beast/global"
	"errors"
	"sync"
	"syscall"
	"time"
)

func NewIoThread(server *TcpServer, index int) *IoThread {
	if server == nil {
		log.Infof("NewIoThread invalid config")
		return nil
	}

	return &IoThread{
		Index:            index,
		Server:           server,
		NotifyList:       []*NotifyEvent{},
		NotifyMutex:      sync.Mutex{},
		ReadTmpBuffer:    make([]byte, global.READ_BUFFER_LEN),
		NotifyWriteBytes: [8]byte{1, 0, 0, 0, 0, 0, 0, 0},
		NotifyReadBytes:  [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (this *IoThread) Start() error {
	//this.EpollFd, err = syscall.EpollCreate(this.Server.MaxSocketNum)
	pollFd, err := aio.NewPoller()
	if err != nil {
		log.Infof("IoThread EpollCreate failed")
		return err
	}
	this.EpollFd = int(pollFd)

	//r1, _, errn := syscall.Syscall(284, 0, 0, 0)
	r, w, err := aio.Pipe()
	if err != nil {
		log.Infof("SYS_EVENTFD failed,IoThread exit,errn=%d", err)
		return nil
	}
	this.NotifyFdR = r
	this.NotifyFdW = w

	err = pollFd.Add(this.NotifyFdR, aio.In|aio.Err)
	if err != nil {
		log.Infof("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
		syscall.Close(this.NotifyFdR)
		return err
	}
	/*
		err = pollFd.Add(this.NotifyFdW, aio.Out|aio.Err)
		if err != nil {
			log.Infof("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
			syscall.Close(this.NotifyFdW)
			return err
		}
	*/

	go func() {
		lastNotify := time.Now()
		lastCheck := time.Now()
		var ts int
		var b_handle_notify bool

		for {
			if time.Since(lastNotify) < time.Second {
				ts = 50
			} else {
				ts = 1000
			}

			if time.Since(lastCheck) > time.Duration(this.Server.CheckTimeoutTs)*time.Second {
				this.checkTimeoutFds()
				lastCheck = time.Now()
			}

			nEvents, err := pollFd.Wait(time.Duration(ts) * time.Millisecond)
			if err != nil {
				log.Infof("IoThread EpollWait failed,err=%s", err.Error())
				continue
			}

			if len(nEvents) > 0 {
				lastNotify = time.Now()
			}

			b_handle_notify = false
			for i := 0; i < len(nEvents); i++ {
				fd := int(nEvents[i].Fd)
				if fd > this.Server.MaxSocketNum {
					log.Infof("IoThread EpollWait invalid fd:%d", fd)
					continue
				}

				//Have custom events
				if fd == this.NotifyFdR {
					_, err := syscall.Read(this.NotifyFdR, this.NotifyReadBytes[:])
					if err != nil {
						log.Infof("NotifyFd Read,err:%s", err.Error())
						continue
					}
					b_handle_notify = true
					log.Infof("NotifyFdR fd")
					continue
				}

				if nEvents[i].Flags&aio.Err > 0 {
					log.Infof("IoThread EpollWait Err")
					this.closeConn(fd)
					continue
				}

				if nEvents[i].Flags&aio.In > 0 {
					log.Infof("IoThread EpollWait In")
					this.handleRead(fd)
				}

				if nEvents[i].Flags&aio.Out > 0 {
					log.Infof("IoThread EpollWait Out")
					this.handleWrite(fd)
				}
			}
			if b_handle_notify {
				this.handleNotify()
			}
		}

	}()
	return nil
}

func (this *IoThread) handleRead(fd int) error {
	connInfo := this.Server.ConnList[fd]
	if connInfo == nil {
		log.Infof("IoThread HandleRead already closed,fd=%d", fd)

		return errors.New("HandleRead closed")
	}

	connInfo.SocketInfo.UpdateAccessTime()

	n, err := syscall.Read(fd, this.ReadTmpBuffer)
	//what EAGAIN is return
	if err != nil || n < 0 {
		log.Infof("HandleRead Read error,fd=%d,socket=%+v", fd, connInfo.SocketInfo)
		this.closeConn(fd)
		return errors.New("Read error")
	}

	//socket closes and sends no data in the buffer.
	if n == 0 {
		log.Infof("HandleRead close by peer,fd=%d,socket=%+v", fd, connInfo.SocketInfo)
		this.closeConn(fd)
		return nil
	}
	if n > global.READ_BUFFER_LEN {
		log.Infof("HandleRead read too much,n:%d", n)
		//panic("HandleRead read too much")
		return errors.New("read too much")
	}

	recv_msg := this.ReadTmpBuffer[0:n]
	//log.Infof("HandleRead read msg:%#v", recv_msg)

	connInfo.SocketInfo.ReadBuffer.Write(recv_msg)
	var whole_msg []byte
	for {
		whole_msg = connInfo.SocketInfo.ReadBuffer.Bytes()
		ok, packlen := this.Server.Parser.Unpack(whole_msg, connInfo)
		if !ok {
			log.Infof("HandleRead Unpack failed,whole_msg:%#v", whole_msg)
			this.closeConn(fd)
			return errors.New("Read error")
		}
		if packlen == 0 {
			log.Infof("HandleRead incomplete pack,,whole_msg:%#v", whole_msg)
			break
		}

		if packlen > global.MAX_PACK_LEN {
			log.Infof("HandleRead invalid packlen:%d", packlen)
			this.closeConn(fd)
			return errors.New("invalid packlen")
		}

		if packlen > connInfo.SocketInfo.ReadBuffer.Len() {
			log.Infof("HandleRead need more,packlen:%d,current len:%d", packlen, connInfo.SocketInfo.ReadBuffer.Len())
			break
		}
		pack := connInfo.SocketInfo.ReadBuffer.Next(packlen)
		if !this.Server.Parser.HandlePack(pack, connInfo) {
			log.Infof("HandleRead HandlePack failed,pack:%#v", pack)
			this.closeConn(fd)
			return errors.New("HandlePack failed")
		}
		//log.Infof("HandleRead HandlePack ok,pack:%#v", pack)
		if 0 == connInfo.SocketInfo.ReadBuffer.Len() {
			log.Infof("HandleRead HandlePack finished")
			break
		}
	}

	//this.sendMsg(fd, s.ReadTmpBuffer[0:n])
	return nil
}

func (this *IoThread) handleWrite(fd int) error {
	connInfo := this.Server.ConnList[fd]
	if connInfo == nil {
		log.Infof("IoThread HandleWrite already closed,fd=%d", fd)
		return errors.New("HandleWrite closed")
	}

	connInfo.SocketInfo.UpdateAccessTime()

	writeBuffLen := connInfo.SocketInfo.WriteBuffer.Len()
	if writeBuffLen == 0 {
		log.Infof("HandleWrite no data to write")
		return nil
	}

	connInfo.SocketInfo.WriteMutex.Lock()
	defer connInfo.SocketInfo.WriteMutex.Unlock()
	writeBuffLen = connInfo.SocketInfo.WriteBuffer.Len()

	n, err := syscall.Write(fd, connInfo.SocketInfo.WriteBuffer.Bytes())
	//Look at what EAGAIN will return
	if err != nil || n < 0 {
		log.Infof("HandleWrite Write error,fd=%d,addr=%s", fd, connInfo.SocketInfo.Addr)
		this.closeConn(fd)
		return errors.New("Write error")
	}

	if n == writeBuffLen {
		//has finished, cancel the EPOLLOUT event
		err := aio.Poller(this.EpollFd).Add(fd, aio.In|aio.Err)
		if err != nil {
			log.Infof("HandleWrite EpollModFd failed,fd=%d,err=%s", fd, err.Error())
			this.closeConn(fd)
			return errors.New("HandleWrite EpollModFd failed")
		}
		connInfo.SocketInfo.WriteBuffer.Reset()
		//this.Server.Parser.WriteFinishCb(connInfo)
		//log.Infof("HandleWrite Reset WriteBuffer,Cap=%d", connInfo.SocketInfo.WriteBuffer.Cap())
	} else {
		//modifying the offset of WriteBuffer
		connInfo.SocketInfo.WriteBuffer.Next(n)
	}

	return nil
}

func (this *IoThread) tryWrite(fd int) error {
	return this.handleWrite(fd)
}

//Close connections within the IO thread
func (this *IoThread) closeConn(fd int) error {
	c := this.Server.ConnList[fd]
	if c == nil {
		log.Infof("CloseConn already closed,fd=%d", fd)
		return nil
	}

	if c.SocketInfo.WriteBuffer.Len() != 0 {
		this.tryWrite(c.SocketInfo.Fd)
	}

	aio.Poller(this.EpollFd).Delete(fd)
	syscall.Close(fd)
	this.Server.ConnList[fd] = nil
	log.Infof("CloseConn ok,SocketInfo:%+v", c.SocketInfo)
	return nil
}

//checkTimeoutFds
func (this *IoThread) checkTimeoutFds() {
	var index int
	cur := time.Now().Unix()
	for i := 0; ; i++ {
		index = i*this.Server.IoThreadNum + this.Index
		if index >= this.Server.MaxSocketNum {
			break
		}

		c := this.Server.ConnList[index]
		if c == nil {
			//log.Infof("checkTimeoutFds CloseConn already closed,fd=%d", fd)
			continue
		}

		if cur >= c.SocketInfo.LastAccessTime+int64(this.Server.TimeoutTs) {
			log.Infof("checkTimeoutFds ok,fd:%d,cur:%d,c.SocketInfo:%+v", index, cur, c.SocketInfo)
			this.closeConn(index)
		}

	}
	//log.Infof("checkTimeoutFds finished")
}

//Check whether the socket only ID matches in asynchronous mode
func (this *IoThread) CheckSocketInfo(socketInfo *SocketInfo) bool {
	return this.Server.CheckSocketId(socketInfo.Fd, socketInfo.Id)
}

//called in the IO thread
func (this *IoThread) CloseDirect(fd int) error {
	return this.closeConn(fd)
}

//called in the IO thread
func (this *IoThread) WriteDirect(fd int, msg []byte) error {
	n, err := syscall.Write(fd, msg)
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		log.Infof("WriteDirect error,fd=%d,msg=%s", fd, string(msg))
		this.closeConn(fd)
		return errors.New("Write error")
	}
	return nil
}
