package server

import (
	"beast/global"
)

//Send messages synchronously, only used within the IO threads
func (this *ConnInfo) SynSendMsg(msg []byte) bool {
	return this.T.WriteDirect(this.SocketInfo.Fd, msg) == nil
}

//Asynchronous send message
func (this *ConnInfo) AsynSendMsg(msg []byte) bool {
	if !this.SocketInfo.AddMsgToWriteBuffer(msg) {
		log.Infof("ConnInfo SendMsg failed,c:%+v", this)
		this.AsynClose()
		return false
	}
	this.T.Notify(global.EVENT_WRITE, this.SocketInfo)
	return true
}

//Shut down the connection synchronously and can only be used in the IO thread
func (this *ConnInfo) SynClose() bool {
	return this.T.CloseDirect(this.SocketInfo.Fd) == nil
}

//Asynchronous close connection
func (this *ConnInfo) AsynClose() {
	log.Infof("ConnInfo Close,c:%+v", this)
	this.T.Notify(global.EVENT_CLOSE, this.SocketInfo)
}
