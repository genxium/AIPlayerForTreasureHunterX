package main

import (
	"AI/models"
	"encoding/json"
	"flag"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	//"go.uber.org/zap"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"
)

const (
	WAITING       = 0
	IN_BATTLE     = 1
	IN_SETTLEMENT = 2
	IN_DISMISSAL  = 3
)

type wsReq struct {
	MsgId int             `json:"msgId"`
	Act   string          `json:"act"`
	Data  json.RawMessage `json:"data"`
}

type wsResp struct {
	Ret         int32           `json:"ret,omitempty"`
	EchoedMsgId int32           `json:"echoedMsgId,omitempty"`
	Act         string          `json:"act,omitempty"`
	Data        json.RawMessage `json:"data,omitempty"`
}

type wsRespPb struct {
	Ret         int32  `json:"ret,omitempty"`
	EchoedMsgId int32  `json:"echoedMsgId,omitempty"`
	Act         string `json:"act,omitempty"`
	Data        []byte `json:"data,omitempty"`
}

type Client struct {
	LastRoomDownsyncFrame *models.RoomDownsyncFrame
	BattleState           int
	c                     *websocket.Conn
	Player                *models.Player
}

var addr = flag.String("addr", "localhost:9992", "http service address")

func main() {
	flag.Parse()
	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: *addr, Path: "/tsrht"}
	q := u.Query()
	q.Set("intAuthToken", "70d8efd491f0ca8dfc20044455da59f6")
	u.RawQuery = q.Encode()
	//ref to the NewClient and DefaultDialer.Dial https://github.com/gorilla/websocket/issues/54
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer close(done)
		client := &Client{
			LastRoomDownsyncFrame: nil,
			BattleState:           -1,
			c:                     c,
			Player:                &models.Player{Id: 9},
		}
		for {
			var resp *wsResp
			resp = new(wsResp)
			err := c.ReadJSON(resp)
			if err != nil {
				//log.Println("marshal wsResp:", err)
			}
			if resp.Act == "RoomDownsyncFrame" {
				var respPb *wsRespPb
				respPb = new(wsRespPb)
				err := c.ReadJSON(respPb)
				if err != nil {
					//log.Println("marshal respPb:", err)
				}
				client.decodeProtoBuf(respPb.Data)
				client.upsyncFrameData()
			} else {
				//handleHbRequirements(resp)
			}
			time.Sleep(time.Duration(int64(20)))
		}
	}()

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("interrupt")

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

func (client *Client) upsyncFrameData() {
	if client.LastRoomDownsyncFrame.Id == 1 {
		client.BattleState = IN_BATTLE
		log.Println("Game Start")
	}
	if client.BattleState == IN_BATTLE {
		newFrame := &struct {
			Id            int64             `json:"id"`
			X             float64           `json:"x"`
			Y             float64           `json:"y"`
			Dir           *models.Direction `json:"dir"`
			AckingFrameId int32             `json:"AckingFrameId"`
		}{client.Player.Id, client.LastRoomDownsyncFrame.Players[int32(client.Player.Id)].X + 1, client.LastRoomDownsyncFrame.Players[int32(client.Player.Id)].Y + 1, client.LastRoomDownsyncFrame.Players[int32(client.Player.Id)].Dir, client.LastRoomDownsyncFrame.Id}
		newFrameByte, err := json.Marshal(newFrame)
		if err != nil {
			log.Println("json Marshal:", err)
			return
		}
		req := &wsReq{
			MsgId: 1,
			Act:   "PlayerUpsyncCmd",
			Data:  newFrameByte,
		}
		reqByte, err := json.Marshal(req)
		err = client.c.WriteMessage(websocket.TextMessage, reqByte)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}

func (client *Client) decodeProtoBuf(message []byte) {
	room_downsync_frame := models.RoomDownsyncFrame{}
	err := proto.Unmarshal(message, &room_downsync_frame)
	if err != nil {
		log.Fatal(err)
	}
	client.LastRoomDownsyncFrame = &room_downsync_frame
}
