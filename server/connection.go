package server

import (
	"beast/global"
)

//Send messages synchronously, only used within the IO threads
func (c *ConnInfo) SynSendMsg(msg []byte) bool {
	return c.T.WriteDirect(c.SInfo.Fd, msg) == nil
}

//Asynchronous send message
func (c *ConnInfo) AsynSendMsg(msg []byte) bool {
	if !c.SInfo.AddMsgToWriteBuffer(msg) {
		log.Infof("ConnInfo SendMsg failed,c:%+v", c)
		c.AsynClose()
		return false
	}
	c.T.Notify(global.EVENT_WRITE, c.SInfo)
	//logging.Debug("SendMsg msg:%#v,c:%+v",msg,c)
	return true
}

//Shut down the connection synchronously and can only be used in the IO thread
func (c *ConnInfo) SynClose() bool {
	return c.T.CloseDirect(c.SInfo.Fd) == nil
}

//Asynchronous close connection
func (c *ConnInfo) AsynClose() {
	log.Infof("ConnInfo Close,c:%+v", c)
	c.T.Notify(global.EVENT_CLOSE, c.SInfo)
}
