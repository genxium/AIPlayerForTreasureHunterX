package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()
	r.GET("/spawnBot", func(c *gin.Context) {
		expectedRoomId := c.Query("expectedRoomId")
		fmt.Println("Need a bot to join room: " + expectedRoomId)
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})
	r.Run() // listen and serve on 0.0.0.0:8080
}
