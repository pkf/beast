package global

type Configuration struct {
	ServerHost     string
	EnableSSL      int
	Mode           string
	MaxSocketNum   int
	CheckTimeoutTs int
	TimeoutTs      int
}
