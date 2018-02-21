package global

const (
	EVENT_BEGIN  = iota //0
	EVENT_ACCEPT        //1
	EVENT_WRITE         //2
	EVENT_CLOSE         //3
)

const (
	ACCEPT_CHAN_LEN = 102400 //backlog
	READ_BUFFER_LEN = 8192   //读缓冲区默认大小
	MAX_PACK_LEN    = 102400 //最大包长
)

type Configuration struct {
	ServerHost     string
	EnableSSL      int
	Mode           string
	MaxSocketNum   int
	CheckTimeoutTs int
	TimeoutTs      int
}
