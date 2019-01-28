package main

import (
	"AI/models"
	"encoding/json"
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type wsReq struct {
	MsgId int             `json:"msgId"`
	Act   string          `json:"act"`
	Data  json.RawMessage `json:"data"`
}

var addr = flag.String("addr", "localhost:9992", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/tsrht"}
	q := u.Query()
	q.Set("intAuthToken", "019eb5719c287d82115bd55cce613d9a")
	u.RawQuery = q.Encode()
	log.Printf("connecting to %s", u.String())
	//ref to the NewClient and DefaultDialer.Dial https://github.com/gorilla/websocket/issues/54
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			var pReq *wsReq
			pReq = new(wsReq)
			err := c.ReadJSON(pReq)
			//_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("pReq: %v", pReq)
			log.Printf("Data: %v", pReq.Data)
			if pReq.Act == "RoomDownsyncFrame" {
				decodeProtoBuf(pReq.Data)
			}
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func decodeProtoBuf(message []byte) models.RoomDownsyncFrame {
	room_downsync_frame := models.RoomDownsyncFrame{}
	err := proto.Unmarshal(message, &room_downsync_frame)
	if err != nil {
		log.Fatal("unmarshaling error: ", err)
	}
	return room_downsync_frame
}

func encodeProtobuf() {
}
