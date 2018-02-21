package server

import (
	"beast/aio"
	"beast/global"
	"errors"
	"sync"
	"syscall"
	"time"
)

func NewIoThread(o *TcpServer, index int) *IoThread {
	if o == nil {
		log.Infof("NewIoThread invalid config")
		return nil
	}

	return &IoThread{
		Index:            index,
		Owner:            o,
		NotifyList:       []*NotifyEvent{},
		NotifyMutex:      sync.Mutex{},
		ReadTmpBuffer:    make([]byte, global.READ_BUFFER_LEN),
		NotifyWriteBytes: [8]byte{1, 0, 0, 0, 0, 0, 0, 0},
		NotifyReadBytes:  [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (s *IoThread) Start() error {
	//s.EpollFd, err = syscall.EpollCreate(s.Owner.MaxSocketNum)
	pollFd, err := aio.NewPoller()
	if err != nil {
		log.Infof("IoThread EpollCreate failed")
		return err
	}
	s.EpollFd = int(pollFd)

	//r1, _, errn := syscall.Syscall(284, 0, 0, 0)
	r, w, err := aio.Pipe()
	if err != nil {
		log.Infof("SYS_EVENTFD failed,IoThread exit,errn=%d", err)
		return nil
	}
	s.NotifyFdR = r
	s.NotifyFdW = w

	err = pollFd.Add(s.NotifyFdR, aio.In|aio.Err)
	if err != nil {
		log.Infof("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
		syscall.Close(s.NotifyFdR)
		return err
	}
	/*
		err = pollFd.Add(s.NotifyFdW, aio.Out|aio.Err)
		if err != nil {
			log.Infof("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
			syscall.Close(s.NotifyFdW)
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

			if time.Since(lastCheck) > time.Duration(s.Owner.CheckTimeoutTs)*time.Second {
				s.checkTimeoutFds()
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
				if fd > s.Owner.MaxSocketNum {
					log.Infof("IoThread EpollWait invalid fd:%d", fd)
					continue
				}

				//有自定义事件
				if fd == s.NotifyFdR {
					_, err := syscall.Read(s.NotifyFdR, s.NotifyReadBytes[:])
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
					s.closeConn(fd)
					continue
				}

				if nEvents[i].Flags&aio.In > 0 {
					log.Infof("IoThread EpollWait In")
					s.handleRead(fd)
				}

				if nEvents[i].Flags&aio.Out > 0 {
					log.Infof("IoThread EpollWait Out")
					s.handleWrite(fd)
				}
			}
			if b_handle_notify {
				s.handleNotify()
			}
		}

	}()
	return nil
}

func (s *IoThread) handleRead(fd int) error {
	connInfo := s.Owner.ConnList[fd]
	if connInfo == nil {
		log.Infof("IoThread HandleRead already closed,fd=%d", fd)

		return errors.New("HandleRead closed")
	}

	connInfo.SInfo.UpdateAccessTime()

	n, err := syscall.Read(fd, s.ReadTmpBuffer)
	//看看EAGAIN会返回什么
	if err != nil || n < 0 {
		log.Infof("HandleRead Read error,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return errors.New("Read error")
	}

	//对端关闭，并且发送缓冲区无数据
	if n == 0 {
		log.Infof("HandleRead close by peer,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return nil
	}
	if n > global.READ_BUFFER_LEN {
		log.Infof("HandleRead read too much,n:%d", n)
		//panic("HandleRead read too much")
		return errors.New("read too much")
	}

	recv_msg := s.ReadTmpBuffer[0:n]
	//log.Infof("HandleRead read msg:%#v", recv_msg)

	connInfo.SInfo.ReadBuffer.Write(recv_msg)

	var whole_msg []byte
	for {
		whole_msg = connInfo.SInfo.ReadBuffer.Bytes()
		ok, packlen := s.Owner.Parser.Unpack(whole_msg, connInfo)
		if !ok {
			log.Infof("HandleRead Unpack failed,whole_msg:%#v", whole_msg)
			s.closeConn(fd)
			return errors.New("Read error")
		}
		if packlen == 0 {
			log.Infof("HandleRead incomplete pack,,whole_msg:%#v", whole_msg)
			break
		}

		if packlen > global.MAX_PACK_LEN {
			log.Infof("HandleRead invalid packlen:%d", packlen)
			s.closeConn(fd)
			return errors.New("invalid packlen")
		}

		if packlen > connInfo.SInfo.ReadBuffer.Len() {
			log.Infof("HandleRead need more,packlen:%d,current len:%d", packlen, connInfo.SInfo.ReadBuffer.Len())
			break
		}
		pack := connInfo.SInfo.ReadBuffer.Next(packlen)
		if !s.Owner.Parser.HandlePack(pack, connInfo) {
			log.Infof("HandleRead HandlePack failed,pack:%#v", pack)
			s.closeConn(fd)
			return errors.New("HandlePack failed")
		}
		//log.Infof("HandleRead HandlePack ok,pack:%#v", pack)
		if 0 == connInfo.SInfo.ReadBuffer.Len() {
			log.Infof("HandleRead HandlePack finished")
			break
		}
	}
	//测试
	//s.sendMsg(fd, s.ReadTmpBuffer[0:n])

	return nil
}

func (s *IoThread) handleWrite(fd int) error {
	connInfo := s.Owner.ConnList[fd]
	if connInfo == nil {
		log.Infof("IoThread HandleWrite already closed,fd=%d", fd)
		return errors.New("HandleWrite closed")
	}

	connInfo.SInfo.UpdateAccessTime()

	writeBuffLen := connInfo.SInfo.WriteBuffer.Len()
	if writeBuffLen == 0 {
		log.Infof("HandleWrite no data to write")
		return nil
	}

	connInfo.SInfo.WriteMutex.Lock()
	defer connInfo.SInfo.WriteMutex.Unlock()
	writeBuffLen = connInfo.SInfo.WriteBuffer.Len()

	n, err := syscall.Write(fd, connInfo.SInfo.WriteBuffer.Bytes())
	//看看EAGAIN会返回啥
	if err != nil || n < 0 {
		log.Infof("HandleWrite Write error,fd=%d,addr=%s", fd, connInfo.SInfo.Addr)
		s.closeConn(fd)
		return errors.New("Write error")
	}

	if n == writeBuffLen {
		//已发完，取消 EPOLLOUT事件
		err := aio.Poller(s.EpollFd).Add(fd, aio.In|aio.Err)
		if err != nil {
			log.Infof("HandleWrite EpollModFd failed,fd=%d,err=%s", fd, err.Error())
			s.closeConn(fd)
			return errors.New("HandleWrite EpollModFd failed")
		}
		connInfo.SInfo.WriteBuffer.Reset()
		//s.Owner.Parser.WriteFinishCb(connInfo)
		//log.Infof("HandleWrite Reset WriteBuffer,Cap=%d", connInfo.SInfo.WriteBuffer.Cap())
	} else {
		//修改WriteBuffer的偏移
		connInfo.SInfo.WriteBuffer.Next(n)
	}

	return nil
}

func (s *IoThread) tryWrite(fd int) error {
	return s.handleWrite(fd)
}

//io线程内关闭连接
func (s *IoThread) closeConn(fd int) error {
	c := s.Owner.ConnList[fd]
	if c == nil {
		log.Infof("CloseConn already closed,fd=%d", fd)
		return nil
	}

	if c.SInfo.WriteBuffer.Len() != 0 {
		s.tryWrite(c.SInfo.Fd)
	}

	aio.Poller(s.EpollFd).Delete(fd)
	syscall.Close(fd)
	s.Owner.ConnList[fd] = nil
	log.Infof("CloseConn ok,SocketInfo:%+v", c.SInfo)
	return nil
}

//checkTimeoutFds
func (s *IoThread) checkTimeoutFds() {
	var index int
	cur := time.Now().Unix()
	for i := 0; ; i++ {
		index = i*s.Owner.IoThreadNum + s.Index
		if index >= s.Owner.MaxSocketNum {
			break
		}

		c := s.Owner.ConnList[index]
		if c == nil {
			//log.Infof("checkTimeoutFds CloseConn already closed,fd=%d", fd)
			continue
		}

		if cur >= c.SInfo.LastAccessTime+int64(s.Owner.TimeoutTs) {
			log.Infof("checkTimeoutFds ok,fd:%d,cur:%d,c.SInfo:%+v", index, cur, c.SInfo)
			s.closeConn(index)
		}

	}
	//log.Infof("checkTimeoutFds finished")
}

//异步场景下检查socket唯一id是否匹配
func (s *IoThread) CheckSocketInfo(socketInfo *SocketInfo) bool {
	return s.Owner.CheckSocketId(socketInfo.Fd, socketInfo.Id)
}

//在io线程中才可以调用
func (s *IoThread) CloseDirect(fd int) error {
	return s.closeConn(fd)
}

//在io线程中才可以调用
func (s *IoThread) WriteDirect(fd int, msg []byte) error {
	n, err := syscall.Write(fd, msg)
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		log.Infof("WriteDirect error,fd=%d,msg=%s", fd, string(msg))
		s.closeConn(fd)
		return errors.New("Write error")
	}
	return nil
}
