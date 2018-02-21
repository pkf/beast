package server

import (
	"beast/global"
	"beast/util"
	"errors"
	"strings"
	"syscall"
)

var (
	Server *TcpServer
)

//The number of IO threads should not exceed the number of CPU physical core (the number of non logical processors)
//cat /proc/cpuinfo| grep "cpu cores"| uniq
func InitServer(ioNum, maxSocketNum, checkTimeoutTs, timeoutTs int, addr string, parser TcpParser) {
	Server = NewTcpServer(ioNum, maxSocketNum, checkTimeoutTs, timeoutTs, addr, parser)
	Server.Start()
}

func NewTcpServer(ioNum, maxSocketNum, checkTimeoutTs, timeoutTs int, addr string, parser TcpParser) *TcpServer {
	if maxSocketNum == 0 || ioNum == 0 {
		log.Infof("NewTcpServer invalid config")
		return nil
	}

	return &TcpServer{
		ConnList:       make([]*ConnInfo, maxSocketNum),
		MaxSocketNum:   maxSocketNum,
		IoThreadNum:    ioNum,
		Addr:           addr,
		UniqueId:       0,
		Parser:         parser,
		CheckTimeoutTs: checkTimeoutTs,
		TimeoutTs:      timeoutTs,
	}
}

func (s *TcpServer) Start() error {

	for i := 0; i < s.IoThreadNum; i++ {
		iothread := NewIoThread(s, i)
		s.IoThreadList = append(s.IoThreadList, iothread)
		iothread.Start()
	}

	listenfd, err := s.CreateListenSocket(s.Addr)
	if err != nil {
		log.Infof("TcpServer CreateListenSocket failed")
		return err
	}

	for {
		//logging.Debug("loop1")
		fd, addr, err := syscall.Accept(listenfd)
		if err != nil {
			log.Infof("TcpServer Accept failed")
			continue
		}

		if fd > s.MaxSocketNum {
			log.Infof("IoThread Accept invalid fd:%d", fd)
			continue
		}
		id := s.CreateUniqueId()

		socketInfo := NewSocketInfo(fd, id, addr)
		ioIndex := fd % s.IoThreadNum
		log.Infof("Accept fd=%d,id:%d,ioIndex=%d,addr=%+v,err=%+v", fd, id, ioIndex, addr, err)
		s.IoThreadList[ioIndex].Notify(global.EVENT_ACCEPT, socketInfo)
	}

	return nil
}

func (s *TcpServer) CreateUniqueId() uint64 {
	s.UniqueId += 1
	return s.UniqueId
}

//异步场景下检查socket唯一id是否匹配
func (s *TcpServer) CheckSocketId(fd int, id uint64) bool {
	if fd > s.MaxSocketNum {
		log.Infof("TcpServer CheckSocketId failed,fd:%d,id:%d", fd, id)
		return false
	}
	c := s.ConnList[fd]
	if c == nil {
		log.Infof("TcpServer CheckSocketId failed, already closed,fd:%d,id:%d", fd, id)
		return false
	}
	return c.SInfo.Id == id
}

func (s *TcpServer) SendMsg(fd int, id uint64, msg []byte) error {
	if len(msg) == 0 {
		log.Infof("TcpServer SendMsg empty,fd=%d", fd)
		return nil
	}
	if !s.CheckSocketId(fd, id) {
		log.Infof("TcpServer SendMsg CheckSocketId failed,fd:%d,id:%d", fd, id)
		return errors.New("CheckSocketId failed")
	}
	c := s.ConnList[fd]
	if c != nil {
		c.AsynSendMsg(msg)
		log.Infof("TcpServer SendMsg ok,msg:%#v,fd:%d,id:%d", msg, fd, id)
	} else {
		log.Infof("TcpServer SendMsg failed,socket closed,msg:%#v,fd:%d,id:%d", msg, fd, id)
	}
	return nil
}

func (s *TcpServer) CreateListenSocket(ipport string) (int, error) {
	socket, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	var flag = int(1)
	err := syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, flag)
	if err != nil {
		log.Infof("TcpServer Setsockopt failed")
		return 0, err
	}

	ipinfo := strings.Split(ipport, ":")
	if len(ipinfo) != 2 {
		log.Infof("TcpServer invalid ipport:%s", ipport)
		return 0, errors.New("invalid ipport")
	}

	ip, err := util.ParseIPv4(ipinfo[0])
	if err != nil {
		log.Infof("TcpServer parseIPv4 failed")
		return 0, errors.New("invalid ip")
	}

	port, err := util.ParsePort(ipinfo[1])
	if err != nil {
		log.Infof("TcpServer parsePort failed")
		return 0, err
	}

	addr := &syscall.SockaddrInet4{
		//Family: syscall.AF_INET,
		Port: port,
		Addr: ip,
	}

	err = syscall.Bind(socket, addr)
	if err != nil {
		log.Infof("TcpServer Bind failed")
		return 0, err
	}

	err = syscall.Listen(socket, global.ACCEPT_CHAN_LEN)
	if err != nil {
		log.Infof("TcpServer listen failed")
		return 0, err
	}
	return socket, nil
}
