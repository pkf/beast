package main

import (
	"beast/util"
	"errors"
	"fmt"
	"strings"
	"syscall"
)

var (
	Server *TcpServer
)

type ConnInfo interface {
}

func InitServer(maxSocketNum, checkTimeoutTs int, timeoutTs int, addr string) {
	Server = NewTcpServer(maxSocketNum, checkTimeoutTs, timeoutTs, addr)
	Server.Start()
}

type TcpServer struct {
	ConnList       []*ConnInfo
	MaxSocketNum   int
	Addr           string
	UniqueId       uint64
	CheckTimeoutTs int //多久检查一次
	TimeoutTs      int //多少秒超时
}

func NewTcpServer(maxSocketNum, checkTimeoutTs, timeoutTs int, addr string) *TcpServer {
	if maxSocketNum == 0 {
		fmt.Println("NewTcpServer invalid config")
		return nil
	}

	return &TcpServer{
		ConnList:       make([]*ConnInfo, maxSocketNum),
		MaxSocketNum:   maxSocketNum,
		Addr:           addr,
		UniqueId:       0,
		CheckTimeoutTs: checkTimeoutTs,
		TimeoutTs:      timeoutTs,
	}
}

func (s *TcpServer) Start() error {
	listenfd, err := s.CreateListenSocket(s.Addr)
	if err != nil {
		fmt.Println("TcpServer CreateListenSocket failed")
		return err
	}

	for {
		fd, addr, err := syscall.Accept(listenfd)
		if err != nil {
			fmt.Println("TcpServer Accept failed")
			continue
		}

		if fd > s.MaxSocketNum {
			fmt.Println("IoThread Accept invalid fd:%d", fd)
			continue
		}
		id := s.CreateUniqueId()

		fmt.Println(fd, addr, err, id)
	}

	return nil
}

func (s *TcpServer) CreateUniqueId() uint64 {
	s.UniqueId += 1
	return s.UniqueId
}

/*
func (s *TcpServer) CheckSocketId(fd int, id uint64) bool {
	if fd > s.MaxSocketNum {
		fmt.Println("TcpServer CheckSocketId failed,fd:%d,id:%d", fd, id)
		return false
	}
	c := s.ConnList[fd]
	if c == nil {
		fmt.Println("TcpServer CheckSocketId failed, already closed,fd:%d,id:%d", fd, id)
		return false
	}
	return c.SInfo.Id == id
}
*/

func (s *TcpServer) CreateListenSocket(ipport string) (int, error) {
	socket, _ := syscall.Socket(syscall.AF_INET, syscall.SOCK_STREAM, syscall.IPPROTO_TCP)
	var flag = int(1)
	err := syscall.SetsockoptInt(socket, syscall.SOL_SOCKET, syscall.SO_REUSEADDR, flag)
	if err != nil {
		fmt.Println("TcpServer Setsockopt failed")
		return 0, err
	}

	ipinfo := strings.Split(ipport, ":")
	if len(ipinfo) != 2 {
		fmt.Println("TcpServer invalid ipport:%s", ipport)
		return 0, errors.New("invalid ipport")
	}

	ip, err := util.ParseIPv4(ipinfo[0])
	if err != nil {
		fmt.Println("TcpServer parseIPv4 failed")
		return 0, errors.New("invalid ip")
	}

	port, err := util.ParsePort(ipinfo[1])
	if err != nil {
		fmt.Println("TcpServer parsePort failed")
		return 0, err
	}

	addr := &syscall.SockaddrInet4{
		Port: port,
		Addr: ip,
	}

	err = syscall.Bind(socket, addr)
	if err != nil {
		fmt.Println("TcpServer Bind failed")
		return 0, err
	}

	err = syscall.Listen(socket, util.ACCEPT_CHAN_LEN)
	if err != nil {
		fmt.Println("TcpServer listen failed")
		return 0, err
	}
	return socket, nil
}
