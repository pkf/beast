package server

import (
	"beast/util"
	"bytes"
	"sync"
	"syscall"
)

var log = util.Log

type SocketInfo struct {
	Fd             int
	Addr           syscall.Sockaddr
	LastAccessTime int64
	ReadBuffer     *bytes.Buffer
	WriteBuffer    *bytes.Buffer
	EpollFlag      int        //EpollModFd is more time-consuming and can save the last epoll flag, temporarily not optimizes
	Id             uint64     //Unique identifier
	WriteMutex     sync.Mutex //Protected write buffer
}

type Event struct {
}

type NotifyEvent struct {
	Type int //1.accept  2.write 3.close
	Info *SocketInfo
}

type IoThread struct {
	Index            int
	NotifyList       []*NotifyEvent
	NotifyMutex      sync.Mutex
	NotifyWriteBytes [8]byte
	NotifyReadBytes  [8]byte
	Owner            *TcpServer
	EpollFd          int
	NotifyFdW        int
	NotifyFdR        int
	ReadTmpBuffer    []byte
}

type TcpServer struct {
	ConnList       []*ConnInfo
	MaxSocketNum   int
	IoThreadNum    int
	IoThreadList   []*IoThread
	Addr           string
	UniqueId       uint64
	Parser         TcpParser
	CheckTimeoutTs int //How long will it be checked
	TimeoutTs      int //How many seconds of timeout
}

type BusiInfo struct {
}

type ConnInfo struct {
	SInfo *SocketInfo
	BInfo *BusiInfo
	T     *IoThread
}

/*
下面接口中的两个方法都是在io线程中调用的，不能有任何阻塞操作，业务中有阻塞场景请起协程，然后使用ConnInfo中的异步接口。
Unpack 用来计算包长；遇到非法包才返回false；未判断出包长时packlen返回0；成功时packlen返回包长
HandlePack 用来处理业务包，参数中的msg是完整包
*/
type TcpParser interface {
	Unpack(msg []byte, c *ConnInfo) (ok bool, packlen int)
	HandlePack(msg []byte, c *ConnInfo) (ok bool)
	//WriteFinishCb(c *ConnInfo)
}
