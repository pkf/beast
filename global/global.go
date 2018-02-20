package global

import ()

const (
	EVENT_BEGIN  = iota //0
	EVENT_ACCEPT        //1
	EVENT_WRITE         //2
	EVENT_CLOSE         //3
)

type Configuration struct {
	ServerHost     string
	EnableSSL      int
	Mode           string
	MaxSocketNum   int
	CheckTimeoutTs int
	TimeoutTs      int
}
