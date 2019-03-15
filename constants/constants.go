package constants

import (
  "os"
)

var DOMAIN string
var INT_AUTH_TOKEN string
var PLAYER_ID int64

func Init(){
  if os.Getenv("ENV") == "SERVER" {
    DOMAIN = "tsrht.lokcol.com:9992"
    INT_AUTH_TOKEN = "1da05d70c52a57d1379737bd537cd415"
    PLAYER_ID = 93
  }else{
    DOMAIN = "localhost:9992"
    INT_AUTH_TOKEN = "64a343adc704f73a14f7f291dc8c30fd"
    PLAYER_ID = 8
  }
}
