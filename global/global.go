package global

const (
	EVENT_BEGIN  = iota //0
	EVENT_ACCEPT        //1
	EVENT_WRITE         //2
	EVENT_CLOSE         //3
)

const (
	ACCEPT_CHAN_LEN = 10240
	READ_BUFFER_LEN = 5000   //读缓冲区默认大小
	MAX_PACK_LEN    = 100000 //最大包长
)

type Configuration struct {
	ServerHost     string
	EnableSSL      int
	Mode           string
	MaxSocketNum   int
	CheckTimeoutTs int
	TimeoutTs      int
}
