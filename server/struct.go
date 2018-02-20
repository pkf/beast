package server

import (
	"bytes"
	"sync"
	"syscall"
)

type SocketInfo struct {
	Fd             int
	Addr           syscall.Sockaddr
	LastAccessTime int64
	ReadBuffer     *bytes.Buffer
	WriteBuffer    *bytes.Buffer
	EpollFlag      int        //EpollModFd比较耗时，可以保存上次epoll flag，暂时不优化
	Id             uint64     //唯一标识
	WriteMutex     sync.Mutex //保护写缓冲区
}

type NotifyEvent struct {
	Type int //1.accept  2.write 3.close
	Info *SocketInfo
}

type IoThread struct {
	Index int
	//EventList        []syscall.EpollEvent
	EventList        []interface{}
	NotifyList       []*NotifyEvent
	NotifyMutex      sync.Mutex
	NotifyWriteBytes [8]byte
	NotifyReadBytes  [8]byte
	Owner            *TcpServer
	EpollFd          int
	NotifyFd         int
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
	CheckTimeoutTs int //多久检查一次
	TimeoutTs      int //多少秒超时
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
	Unpack(msg []byte, c *ConnInfo) (ok bool, packlen int) //返回成功失败，包长,包长为0表示包长未知
	HandlePack(msg []byte, c *ConnInfo) (ok bool)
	//WriteFinishCb(c *ConnInfo)
}
