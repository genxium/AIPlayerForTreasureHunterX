package main

import (
	"AI/login"
	"fmt"
)

func main() {
	botName := "bot1"
	intAuthToken, id := login.GetIntAuthTokenByBotName(botName)
	fmt.Printf("intAuthToken: %s, id: %d \n", intAuthToken, id)
}
