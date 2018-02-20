package server

import (
	"beast/aio"
	"beast/global"
	"errors"
	"log"
	"sync"
	"syscall"
	"time"
)

func NewIoThread(o *TcpServer, index int) *IoThread {
	if o == nil {
		log.Println("NewIoThread invalid config")
		return nil
	}

	return &IoThread{
		Index:            index,
		Owner:            o,
		EventList:        make([]Event, int(o.MaxSocketNum/o.IoThreadNum)+1),
		NotifyList:       []*NotifyEvent{},
		NotifyMutex:      sync.Mutex{},
		ReadTmpBuffer:    make([]byte, global.READ_BUFFER_LEN),
		NotifyWriteBytes: [8]byte{1, 0, 0, 0, 0, 0, 0, 0},
		NotifyReadBytes:  [8]byte{0, 0, 0, 0, 0, 0, 0, 0},
	}
}

func (s *IoThread) Start() error {
	//s.EpollFd, err = syscall.EpollCreate(s.Owner.MaxSocketNum)
	pollFd, err := aio.NewPoller(s.Owner.MaxSocketNum)
	if err != nil {
		log.Println("IoThread EpollCreate failed")
		return err
	}
	s.EpollFd = int(pollFd)

	r1, _, errn := syscall.Syscall(284, 0, 0, 0)
	if errn != 0 {
		log.Println("SYS_EVENTFD failed,IoThread exit,errn=%d", errn)
		return nil
	}
	s.NotifyFd = int(r1)

	syscall.SetNonblock(s.NotifyFd, true)
	err = EpollAddFd(s.EpollFd, s.NotifyFd, syscall.EPOLLIN|syscall.EPOLLERR)
	if err != nil {
		log.Println("IoThread  EpollAddFd NotifyFd failed,err=%s", err.Error())
		syscall.Close(s.NotifyFd)
		return err
	}

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

			nEvents, err := syscall.EpollWait(s.EpollFd, s.EventList, ts)
			if err != nil {
				log.Println("IoThread EpollWait failed,err=%s", err.Error())
				continue
			}
			//time.Sleep(time.Microsecond * 100)
			//nEvents := 0
			//log.Println("loop2,nEvents=%d,s.EpollFd=%d", nEvents, s.EpollFd)

			if nEvents > 0 {
				lastNotify = time.Now()
			}

			b_handle_notify = false
			for i := 0; i < nEvents; i++ {
				fd := int(s.EventList[i].Fd)
				if fd > s.Owner.MaxSocketNum {
					log.Println("IoThread EpollWait invalid fd:%d", fd)
					continue
				}

				//有自定义事件
				if fd == s.NotifyFd {
					_, err := syscall.Read(s.NotifyFd, s.NotifyReadBytes[:])
					if err != nil {
						log.Println("NotifyFd Read,err:%s", err.Error())
						continue
					}
					b_handle_notify = true
					continue
				}

				if s.EventList[i].Events&syscall.EPOLLERR > 0 {
					log.Println("IoThread EpollWait EPOLLERR")
					s.closeConn(fd)
					continue
				}

				if s.EventList[i].Events&syscall.EPOLLIN > 0 {
					s.handleRead(fd)
				}

				if s.EventList[i].Events&syscall.EPOLLOUT > 0 {
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
		log.Println("IoThread HandleRead already closed,fd=%d", fd)

		return errors.New("HandleRead closed")
	}
	//connInfo.SInfo.LastAccessTime = time.Now().Unix()

	//tmp := make([]byte, ReadBufferLen)
	connInfo.SInfo.UpdateAccessTime()

	n, err := syscall.Read(fd, s.ReadTmpBuffer)
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		log.Println("HandleRead Read error,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return errors.New("Read error")
	}

	//对端关闭，并且发送缓冲区无数据
	if n == 0 {
		log.Println("HandleRead close by peer,fd=%d,socket=%+v", fd, connInfo.SInfo)
		s.closeConn(fd)
		return nil
	}
	if n > READ_BUFFER_LEN {
		log.Println("HandleRead read too much,n:%d", n)
		//panic("HandleRead read too much")
		return errors.New("read too much")
	}

	recv_msg := s.ReadTmpBuffer[0:n]
	//log.Println("HandleRead read msg:%#v", recv_msg)

	connInfo.SInfo.ReadBuffer.Write(recv_msg)

	var whole_msg []byte
	for {
		whole_msg = connInfo.SInfo.ReadBuffer.Bytes()
		ok, packlen := s.Owner.Parser.Unpack(whole_msg, connInfo)
		if !ok {
			log.Println("HandleRead Unpack failed,whole_msg:%#v", whole_msg)
			s.closeConn(fd)
			return errors.New("Read error")
		}
		if packlen == 0 {
			log.Println("HandleRead incomplete pack,,whole_msg:%#v", whole_msg)
			break
		}

		if packlen > MAX_PACK_LEN {
			log.Println("HandleRead invalid packlen:%d", packlen)
			s.closeConn(fd)
			return errors.New("invalid packlen")
		}

		if packlen > connInfo.SInfo.ReadBuffer.Len() {
			log.Println("HandleRead need more,packlen:%d,current len:%d", packlen, connInfo.SInfo.ReadBuffer.Len())
			break
		}
		pack := connInfo.SInfo.ReadBuffer.Next(packlen)
		if !s.Owner.Parser.HandlePack(pack, connInfo) {
			log.Println("HandleRead HandlePack failed,pack:%#v", pack)
			s.closeConn(fd)
			return errors.New("HandlePack failed")
		}
		//log.Println("HandleRead HandlePack ok,pack:%#v", pack)
		if 0 == connInfo.SInfo.ReadBuffer.Len() {
			log.Println("HandleRead HandlePack finished")
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
		log.Println("IoThread HandleWrite already closed,fd=%d", fd)
		return errors.New("HandleWrite closed")
	}
	//connInfo.SInfo.LastAccessTime = time.Now().Unix()

	connInfo.SInfo.UpdateAccessTime()

	writeBuffLen := connInfo.SInfo.WriteBuffer.Len()
	if writeBuffLen == 0 {
		log.Println("HandleWrite no data to write")
		return nil
	}

	connInfo.SInfo.WriteMutex.Lock()
	defer connInfo.SInfo.WriteMutex.Unlock()
	writeBuffLen = connInfo.SInfo.WriteBuffer.Len() //重新取一遍

	n, err := syscall.Write(fd, connInfo.SInfo.WriteBuffer.Bytes())
	if err != nil || n < 0 { //看看EAGAIN会返回啥
		log.Println("HandleWrite Write error,fd=%d,addr=%s", fd, connInfo.SInfo.Addr)
		s.closeConn(fd)
		return errors.New("Write error")
	}

	if n == writeBuffLen {
		//已发完，取消 EPOLLOUT事件
		err := EpollModFd(s.EpollFd, fd, syscall.EPOLLIN|syscall.EPOLLERR)
		if err != nil {
			log.Println("HandleWrite EpollModFd failed,fd=%d,err=%s", fd, err.Error())
			s.closeConn(fd)
			return errors.New("HandleWrite EpollModFd failed")
		}
		connInfo.SInfo.WriteBuffer.Reset()
		//s.Owner.Parser.WriteFinishCb(connInfo)
		//log.Println("HandleWrite Reset WriteBuffer,Cap=%d", connInfo.SInfo.WriteBuffer.Cap())
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
		log.Println("CloseConn already closed,fd=%d", fd)
		return nil
	}

	if c.SInfo.WriteBuffer.Len() != 0 {
		s.tryWrite(c.SInfo.Fd)
	}

	EpollDelFd(s.EpollFd, fd)
	syscall.Close(fd)
	s.Owner.ConnList[fd] = nil
	log.Println("CloseConn ok,SocketInfo:%+v", c.SInfo)
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
			//log.Println("checkTimeoutFds CloseConn already closed,fd=%d", fd)
			continue
		}

		if cur >= c.SInfo.LastAccessTime+int64(s.Owner.TimeoutTs) {
			log.Println("checkTimeoutFds ok,fd:%d,cur:%d,c.SInfo:%+v", index, cur, c.SInfo)
			s.closeConn(index)
		}

	}
	//log.Println("checkTimeoutFds finished")
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
		log.Println("WriteDirect error,fd=%d,msg=%s", fd, string(msg))
		s.closeConn(fd)
		return errors.New("Write error")
	}
	return nil
}
