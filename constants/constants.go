package constants

// TODO: Read from config file!

type NetConf struct{
  PROTOCOL string
  HOST string
  PORT string
}

var (
  SERVER = NetConf{
	  PROTOCOL    : "http",
	  HOST        : "localhost",
    PORT        : ":9992",
  }
	API         = "/api"
	VERSION     = "/v1"
	PLAYER      = "/player"
	SMS_CAPTCHA = "/SmsCaptcha"
	GET         = "/get"
	LOGIN       = "/login"
)
