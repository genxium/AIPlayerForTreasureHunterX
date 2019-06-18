package main

import (
	"AI/astar"
	"AI/constants"
	"AI/login"
	"AI/models"
	"encoding/json"
	"fmt"
	"net/http"
	"syscall"
	"github.com/ByteArena/box2d"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"math"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"
	"context"
	"sync/atomic"
	"runtime/debug"
)

const (
	WAITING                   = 0
	IN_BATTLE                 = 1
	IN_SETTLEMENT             = 2
	IN_DISMISSAL              = 3
	uniformTimeStepSeconds    = 1.0 / 60.0
	uniformVelocityIterations = 0
	uniformPositionIterations = 0
)

const (
	// You can equivalently use the `GroupIndex` approach, but the more complicated and general purpose approach is used deliberately here. Reference http://www.aurelienribon.com/post/2011-07-box2d-tutorial-collision-filtering.
	COLLISION_CATEGORY_CONTROLLED_PLAYER = (1 << 1)
	COLLISION_CATEGORY_TREASURE          = (1 << 2)
	COLLISION_CATEGORY_TRAP              = (1 << 3)
	COLLISION_CATEGORY_TRAP_BULLET       = (1 << 4)
	COLLISION_CATEGORY_BARRIER           = (1 << 5)
	COLLISION_CATEGORY_PUMPKIN           = (1 << 6)
	COLLISION_CATEGORY_SPEED_SHOES       = (1 << 7)

	COLLISION_MASK_FOR_CONTROLLED_PLAYER = (COLLISION_CATEGORY_TREASURE | COLLISION_CATEGORY_TRAP | COLLISION_CATEGORY_TRAP_BULLET | COLLISION_CATEGORY_SPEED_SHOES | COLLISION_CATEGORY_BARRIER)
	COLLISION_MASK_FOR_TREASURE          = (COLLISION_CATEGORY_CONTROLLED_PLAYER)
	COLLISION_MASK_FOR_TRAP              = (COLLISION_CATEGORY_CONTROLLED_PLAYER)
	COLLISION_MASK_FOR_TRAP_BULLET       = (COLLISION_CATEGORY_CONTROLLED_PLAYER)
	//COLLISION_MASK_FOR_BARRIER           = (COLLISION_CATEGORY_PUMPKIN)
	//COLLISION_MASK_FOR_PUMPKIN           = (COLLISION_CATEGORY_BARRIER)
	COLLISION_MASK_FOR_BARRIER     = (COLLISION_CATEGORY_BARRIER)
	COLLISION_MASK_FOR_PUMPKIN     = (COLLISION_CATEGORY_PUMPKIN)
	COLLISION_MASK_FOR_SPEED_SHOES = (COLLISION_CATEGORY_CONTROLLED_PLAYER)
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
	Id                    int //roomId
	LastRoomDownsyncFrame *models.RoomDownsyncFrame
	BattleState           int
	c                     *websocket.Conn
	Player                *models.Player
	CollidableWorld       *box2d.B2World
	Barrier               map[int32]*models.Barrier
	PlayerCollidableBody  *box2d.B2Body `json:"-"`

	Radian float64
	Dir    models.Direction

	TmxIns               *models.TmxMap

  //上一帧时宝物的数量(因为现在每当一个宝物被吃掉时, 后端downFrame.Treasures会带上它的信息,保存该参数用于判断有没有宝物被吃掉)
	RemovedTreasuresNum  int
	LastFrameRemovedTreasureNum int

	//寻路抽象(Incomplete) --kobako
	pathFinding *models.PathFinding
	Started bool
}

func spawnBot(botName string, expectedRoomId int, botManager *models.BotManager) {
	defer botManager.ReleaseBot(botName)

	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: constants.SERVER.HOST + constants.SERVER.PORT, Path: "/tsrht"}
	q := u.Query()

	//TODO: Error handle
	intAuthToken, playerId := login.GetIntAuthTokenByBotName(botName)

	//local
	q.Set("intAuthToken", intAuthToken)
	if expectedRoomId > 0 {
		q.Set("expectedRoomId", strconv.Itoa(expectedRoomId))
	}
	u.RawQuery = q.Encode()

  fmt.Println("WS connect to " + u.String())

	//ref to the NewClient and DefaultDialer.Dial https://github.com/gorilla/websocket/issues/54
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	client := &Client{
		LastRoomDownsyncFrame: nil,
		BattleState:           -1,
		c:                     c,
		Player:                &models.Player{Id: int32(playerId)},
		Barrier:               make(map[int32]*models.Barrier),
		Radian:                math.Pi / 2,
		Dir:                   models.Direction{Dx: 0, Dy: 1},
		pathFinding:           new(models.PathFinding),
	}

	//初始化地图资源
	tmx, _ := models.InitMapStaticResource("./map/map/pacman/map.tmx")
	client.TmxIns = &tmx

	gravity := box2d.MakeB2Vec2(0.0, 0.0)
	world := box2d.MakeB2World(gravity)

	client.CollidableWorld = &world

	models.CreateBarrierBodysInWorld(&tmx, &world)

	collideMap := models.InitCollideMap(tmx.World, &tmx)
	client.pathFinding.SetCollideMap(collideMap)

	client.Started = false
	killSignal := int32(0)

		/*defer func() {
			if r := recover(); r != nil {
				log.Println("Recovered from panic", r)
			}

      	log.Println("Exiting lifecycle of bot:", botName, " for room:", expectedRoomId)
			close(done)
		}()*/

		
		upsyncLoopFunc := func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from panic in upsync", r)
					fmt.Printf("panic in upsync: %v \n", string(debug.Stack()))
				}
			}()

			for {
				if swapped := atomic.CompareAndSwapInt32(&killSignal, 1, 1); swapped {
					log.Println("Upsync exit")
					return
				}
				client.controller()
				client.checkReFindPath()
				client.upsyncFrameData()
				time.Sleep(time.Second / 15)
			}
		}

		downSyncLoopFunc := func() {
			defer func() {
				if r := recover(); r != nil {
					log.Println("Recovered from panic in downsync", r)
				}
			}()

			for {
				if swapped := atomic.CompareAndSwapInt32(&killSignal, 1, 1); swapped {
					log.Println("Downsync exit")
					return
				}

				var resp *wsResp
				resp = new(wsResp)
				err := c.ReadJSON(resp)
				if err != nil {
					//log.Println("downsync reading err", err)
				}
	
				if resp.Act == "RoomDownsyncFrame" {
					var respPb *wsRespPb
					respPb = new(wsRespPb)
					err := c.ReadJSON(respPb)
					if err != nil {
						log.Println("Err unmarshalling respPb:", err)
					}
					client.decodeProtoBuf(respPb.Data)
					//client.checkReFindPath() //kobako
				} else {
					//handleHbRequirements(resp)
				}
			}
		}

		go upsyncLoopFunc()
		go downSyncLoopFunc()

		elapsedTime := 0
		for {
		  elapsedTime = elapsedTime + 1
		  if elapsedTime > 65 {
			atomic.CompareAndSwapInt32(&killSignal, 0, 1)
			return
		  }
		  time.Sleep(time.Second)
		}
}

func main() {
	startServer(15351)
}

func startServer(port int) {
	var botManager *models.BotManager
	{
		botManager = new(models.BotManager)
		botManager.SetBots([]string{"bot1", "bot2", "bot3", "bot4"})
	}

	r := gin.Default()
  r.GET("/spawnBot", func(c *gin.Context) {
    expectedRoomId, err := strconv.Atoi(c.Query("expectedRoomId"))
    if err != nil {
      fmt.Println("请求中没有或者转换expectedRoomId出错")
      c.JSON(200, gin.H{
        "ret": 1001,
        "err": "请求中没有或者转换expectedRoomId出错",
      })
    } else {
      botName, err := botManager.GetLeisureBot()
      if err != nil {
        fmt.Println("获取空闲bot出错: " + err.Error())
        c.JSON(200, gin.H{
          "ret":     1001,
          "botName": "获取空闲bot出错: " + err.Error(),
        })
      } else {
        go spawnBot(botName, expectedRoomId, botManager)
        fmt.Printf("Get bot: %s, expectedRoomId: %d \n", botName, expectedRoomId)
        c.JSON(200, gin.H{
          "ret":     1000,
          "botName": botName,
        })
      }
    }
  })

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: r,
	}
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Println("Listening: %s\n", zap.Error(err))
		}
	}()
	var gracefulStop = make(chan os.Signal)
	signal.Notify(gracefulStop, syscall.SIGTERM)
	signal.Notify(gracefulStop, syscall.SIGINT)
	sig := <-gracefulStop
	log.Println("Shutdown Server ...")
	log.Println("caught sig", sig)
	log.Println("Wait for 5 second to finish processing")
	clean()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
	  log.Println("Server Shutdown:", zap.Error(err))
	}
	log.Println("Server exiting")
	os.Exit(0)
}

//通过当前玩家的坐标, 和treasureMap来计算start, end point, 用于寻路, 重新初始化walkInfo
func reFindPath(tmx *models.TmxMap, client *Client) {
	var startPoint astar.Point
	{
		playerVec := models.Vec2D{
			X: client.Player.X,
			Y: client.Player.Y,
		}
		temp := tmx.CoordToPoint(playerVec)
		startPoint = astar.Point{
			temp.X,
			temp.Y,
		}
	}

	var endPoint astar.Point
  {
		var min float64 = 999999
		playerVec := models.Vec2D{
			X: client.Player.X,
			Y: client.Player.Y,
		}
    playerPoint := tmx.CoordToPoint(playerVec)
    for id, v := range client.pathFinding.TreasureMap {
      treasurePoint := astar.Point{
			  X: v.X,
			  Y: v.Y,
		  }
      dist := astar.DistBetween(astar.Point{
        X: playerPoint.X,
        Y: playerPoint.Y,
      }, treasurePoint)
      if dist < min {
				min = dist
				endPoint = treasurePoint
        client.pathFinding.UpdateTargetTreasureId(id)
        //client.pathFinding.TargetTreasureId = id
      }
	}
	}

	//fmt.Printf("NEW END POINT %v , NEW TID %d \n", endPoint, client.pathFinding.TargetTreasureId)

	fmt.Printf("++++++ start point %v, end point %v\n", startPoint, endPoint)
  pointPath := client.pathFinding.FindPointPath(startPoint, endPoint)
	fmt.Printf("The point path: %v", pointPath)

	//将离散的路径转为连续坐标, 初始化walkInfo, 每次controller的时候调用
	var path []models.Vec2D
	for _, pt := range pointPath {
		gid := pt.Y*tmx.Width + pt.X
		x, y := tmx.GetCoordByGid(gid)
		path = append(path, models.Vec2D{
			X: x,
			Y: y,
		})
	}
  client.pathFinding.SetNewCoordPath(path)
}

func (client *Client) initTreasureAndPlayers() {

	//根据第一帧的数据来设置好玩家的位置, 以及宝物的位置,以服务器为准
	initFullFrame := client.LastRoomDownsyncFrame

	//Init treasures
	client.LastFrameRemovedTreasureNum = 0

	//Sign on map
	tmx := client.TmxIns

	{ //Init ContinuousPosMap
		//将离散的点转换成连续的点, 用于确认道具的位置(遍历每个点判断相对距离最短)
		var continuousPosMap [][]models.Vec2D
		continuousPosMap = make([][]models.Vec2D, tmx.Height)
		for i := 0; i < tmx.Height; i++ {
			continuousPosMap[i] = make([]models.Vec2D, tmx.Width)
		}
		for i := 0; i < tmx.Height; i++ {
			for j := 0; j < tmx.Width; j++ {
				gid := i*tmx.Width + j
				x, y := tmx.GetCoordByGid(gid)
				continuousPosMap[i][j].X = x
				continuousPosMap[i][j].Y = y
			}
		}
		tmx.ContinuousPosMap = continuousPosMap
	}

	var treasureDiscreteMap map[int32]models.Point
	{
		treasureDiscreteMap = make(map[int32]models.Point)
		//对每一个宝物, 遍历地图找到距离最近的离散点, 标记为宝物
		for id, treasure := range initFullFrame.Treasures {
			coord := models.Vec2D{
				X: treasure.X,
				Y: treasure.Y,
			}
			discretePoint := tmx.CoordToPoint(coord)
			treasureDiscreteMap[id] = discretePoint
		}
	}
  client.pathFinding.SetTreasureMap(treasureDiscreteMap)
	//client.pathFinding.TreasureMap = treasureDiscreteMap

	fmt.Printf("INIT Treasure: %v \n", client.pathFinding.TreasureMap)

	//mark
	var playerPoint astar.Point
	{
		temp := tmx.CoordToPoint(models.Vec2D{
			X: client.Player.X,
			Y: client.Player.Y,
		})
		//client.pathFinding.CurrentPoint = temp
		//类型转换
		playerPoint = astar.Point{
			X: temp.X,
			Y: temp.Y,
		}
	}

	fmt.Printf("++++++ player point: %v \n", playerPoint)

	//找出最近的一个宝物, 标记为client.pathFinding.TargetTreasureId
	minDistance := 99999.0
	for k, v := range treasureDiscreteMap {
		vPoint := astar.Point{
			X: v.X,
			Y: v.Y,
		}
		dist := astar.DistBetween(vPoint, playerPoint)
		if dist < minDistance {
			minDistance = dist
      client.pathFinding.UpdateTargetTreasureId(k)
			//client.pathFinding.TargetTreasureId = k
		}
	}

	fmt.Printf("++++++ minDistance %f, %v \n", minDistance, playerPoint)

	reFindPath(tmx, client)
}

func (client *Client) checkReFindPath() {
	// 仅当 (当前帧的宝物数量比上一帧少 && 目标宝物id被吃掉)  的时候重新寻路
	if client.LastRoomDownsyncFrame == nil {
		return
	}
	if client.LastRoomDownsyncFrame.RefFrameId != 0 && len(client.LastRoomDownsyncFrame.Treasures) != client.LastFrameRemovedTreasureNum {
		//fmt.Printf("last number %d, now number %d \n", client.LastFrameRemovedTreasureNum, len(client.LastRoomDownsyncFrame.Treasures))
		client.LastFrameRemovedTreasureNum = len(client.LastRoomDownsyncFrame.Treasures)

		var needReFindPath bool
		for id, _ := range client.LastRoomDownsyncFrame.Treasures {
      //删除以减轻后续最短距离计算量
			delete(client.pathFinding.TreasureMap, id)
			if id == client.pathFinding.TargetTreasureId {
				needReFindPath = true
			}
		}

		if needReFindPath {
			reFindPath(client.TmxIns, client)
		}
	} else {
		//Do nothing
	}
}

func (client *Client) controller() {
	/*if client.Player.Speed == 0 {
		return
	}*/
	if client.LastRoomDownsyncFrame == nil {
		return
	}
	if !client.Started && client.LastRoomDownsyncFrame.Id > 0 { // 初始帧
		client.Started = true
		log.Println("Game Start")
		client.BattleState = IN_BATTLE
		client.Player.X = client.LastRoomDownsyncFrame.Players[client.Player.Id].X
		client.Player.Y = client.LastRoomDownsyncFrame.Players[client.Player.Id].Y
		//初始化需要寻找的宝物和玩家位置
		client.InitPlayerCollider()
		client.initTreasureAndPlayers()
		fmt.Printf("Init coord: %.2f, %.2f\n", client.Player.X, client.Player.Y);  
		fmt.Printf("Frame coord: %.2f, %.2f\n", client.LastRoomDownsyncFrame.Players[client.Player.Id].X, client.LastRoomDownsyncFrame.Players[client.Player.Id].Y);  
		client.pathFinding.SetCurrentCoord(client.Player.X, client.Player.Y)
		fmt.Printf("Receive id: %d, treasure length %d, refId: %d \n", client.LastRoomDownsyncFrame.Id, len(client.LastRoomDownsyncFrame.Treasures), client.LastRoomDownsyncFrame.RefFrameId)
	} else {
		step := 12.0

		pathFindingMove(client, step)

		//client :q

		//foolMove(client, step);

		//time.Sleep(time.Duration(int64(40)))
	}

}

//撞墙转向
func foolMove(client *Client, step float64) {
	nowRadian := client.Radian

	for nowRadian-client.Radian < math.Pi*2 {
		xStep := step * math.Cos(nowRadian)
		yStep := step * math.Sin(nowRadian)
		//fmt.Println(xStep, yStep);

		//移动collideBody
		newB2Vec2Pos := box2d.MakeB2Vec2(client.Player.X+xStep, client.Player.Y-yStep)
		//newB2Vec2Pos := box2d.MakeB2Vec2(client.Player.X, client.Player.Y - yStep);
		models.MoveDynamicBody(client.PlayerCollidableBody, &newB2Vec2Pos, 0)

		//world.Step
		client.CollidableWorld.Step(uniformTimeStepSeconds, uniformVelocityIterations, uniformPositionIterations)

		//碰撞检测
		collided := false
		for edge := client.PlayerCollidableBody.GetContactList(); edge != nil; edge = edge.Next {
			if edge.Contact.IsTouching() {
				collided = true
				break
				//log.Println("player conteact")
				if _, ok := edge.Other.GetUserData().(*models.Barrier); ok {
					//log.Println("player conteact to the barrier")
				}
			}
		}

		if !collided { //一直走
			client.Player.X = client.Player.X + xStep
			client.Player.Y = client.Player.Y - yStep

			//kobako
			//TODO: set correct direction
			dx, dy := func() (dx float64, dy float64) {
				floorRadian := nowRadian - math.Pi*2*math.Floor(nowRadian/(2*math.Pi))
				//fmt.Println(floorRadian);
				if floorRadian < math.Pi/2 {
					return 2, -1
				} else if floorRadian < math.Pi {
					return -2, -1
				} else if floorRadian < math.Pi*3/2 {
					return -2, 1
				} else {
					return 2, 1
				}
			}()
			client.Dir = models.Direction{
				Dx: dx,
				Dy: dy,
			}
			//fmt.Println(dx, dy)
			//kobako
			break
		} else { //转向
			log.Println("player collided with barriers & change direction: ", nowRadian)
			nowRadian = nowRadian + math.Pi/16

		}
	}

	client.Radian = nowRadian
}

func pathFindingMove(client *Client, step float64) {
	//通过服务器位置进行修正
	//client.pathFinding.SetCurrentCoord(client.Player.X, client.Player.Y)
	client.pathFinding.Move(step)
	//fmt.Println("Before: ", client.Player.X, client.Player.Y);
	client.Player.X = client.pathFinding.CurrentCoord.X
	client.Player.Y = client.pathFinding.CurrentCoord.Y
	client.pathFinding.SetCurrentCoord(client.Player.X, client.Player.Y)
	//fmt.Println("After: ", client.Player.X, client.Player.Y);
}

//lastPos := Position{};

func (client *Client) upsyncFrameData() {
	//if(lastPos)
	/*
	if client.TmxIns.ContinuousPosMap != nil {
		var startPoint astar.Point
		{
			playerVec := models.Vec2D{
				X: client.Player.X,
				Y: client.Player.Y,
			}
			temp := client.TmxIns.CoordToPoint(playerVec)
			startPoint = astar.Point{
				temp.X,
				temp.Y,
			}
		}
		fmt.Printf("(%.2f, %2.f), %v\n", client.Player.X, client.Player.Y, startPoint)
	}*/
	if client.BattleState == IN_BATTLE {
		newFrame := &struct {
			Id int32   `json:"id"`
			X  float64 `json:"x"`
			Y  float64 `json:"y"`
			Dir           models.Direction `json:"dir"`
			AckingFrameId int32     `json:"AckingFrameId"`
		}{client.Player.Id, client.Player.X, client.Player.Y, models.Direction{}, client.LastRoomDownsyncFrame.Id}

		//fmt.Println(newFrame.AckingFrameId)

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

//kobako: 从下行帧解析宝物信息是否减少
func (client *Client) decodeProtoBuf(message []byte) {
  //log.Println("About to decode message into models.RoomDownsyncFrame:", message)
	room_downsync_frame := models.RoomDownsyncFrame{}
	err := proto.Unmarshal(message, &room_downsync_frame)
	if err != nil {
    fmt.Println("解析room_downsync_frame出错了!");
		log.Fatal(err)
	}else{
    //解析room_downsync_frame成功
    //fmt.Println("解析room_downsync_frame成功");
  }

  //fmt.Println(room_downsync_frame.Players);
  //fmt.Println(client.Player.Id);

	//fmt.Printf("Receive id: %d, treasure length %d, refId: %d \n", room_downsync_frame.Id, len(room_downsync_frame.Treasures), room_downsync_frame.RefFrameId)

	//根据最新一帧的信息设置bot玩家的新位置及方向等
	client.LastRoomDownsyncFrame = &room_downsync_frame
	//client.Player.Speed = room_downsync_frame.Players[int32(client.Player.Id)].Speed
	//client.Player.Dir = room_downsync_frame.Players[int32(client.Player.Id)].Dir
	//client.Player.X = room_downsync_frame.Players[int32(client.Player.Id)].X
	//client.Player.Y = room_downsync_frame.Players[int32(client.Player.Id)].Y

	//fmt.Printf("Treasures length: %d \n", len(room_downsync_frame.Treasures))
	//fmt.Printf("room_downsync_frame: Id: %d, RefFrameId: %d, Treasures: %v \n", room_downsync_frame.Id, room_downsync_frame.RefFrameId, room_downsync_frame.Treasures)
	/*
	  for k, v := range room_downsync_frame.Treasures{
	    //fmt.Printf("ID: %d, X: %d, Y: %d || ", v.Id, v.Removed, v.X, v.Y)
	    fmt.Printf("k: %d, v: %v || ", k, v)
	  }
	*/

}

//kobako: Hacked in and stored some info for path finding in the tmxIns
func (client *Client) initMapStaticResource() models.TmxMap {

	relativePath := "./map/map/treasurehunter.tmx"
	execPath, err := os.Executable()
	ErrFatal(err)

	pwd, err := os.Getwd()
	ErrFatal(err)

	fmt.Printf("execPath = %v, pwd = %s, returning...\n", execPath, pwd)

	tmxMapIns := models.TmxMap{}
	pTmxMapIns := &tmxMapIns
	fp := filepath.Join(pwd, relativePath)
	fmt.Printf("fp == %v\n", fp)
	if !filepath.IsAbs(fp) {
		panic("Tmx filepath must be absolute!")
	}

	byteArr, err := ioutil.ReadFile(fp)
	ErrFatal(err)
	models.DeserializeToTmxMapIns(byteArr, pTmxMapIns)

	tsxIns := models.Tsx{}
	pTsxIns := &tsxIns
	relativePath = "./map/map/tile_1.tsx"
	fp = filepath.Join(pwd, relativePath)
	fmt.Printf("fp == %v\n", fp)
	if !filepath.IsAbs(fp) {
		panic("Filepath must be absolute!")
	}

	byteArr, err = ioutil.ReadFile(fp)
	ErrFatal(err)
	models.DeserializeToTsxIns(byteArr, pTsxIns)


	client.InitBarrier(pTmxMapIns, pTsxIns)

	//kobako

	fmt.Println("Barrier")
	fmt.Println(client.Barrier)

	//kobako

	return tmxMapIns
}

func (client *Client) InitPlayerCollider() {
	log.Println("InitPlayerCollider for client.Players:", zap.Any("roomId", client.Id))
	player := client.Player
	var bdDef box2d.B2BodyDef
	colliderOffset := box2d.MakeB2Vec2(0, 0) // Matching that of client-side setting.
	bdDef = box2d.MakeB2BodyDef()
	bdDef.Type = box2d.B2BodyType.B2_dynamicBody
	bdDef.Position.Set(player.X+colliderOffset.X, player.Y+colliderOffset.Y)

	b2PlayerBody := client.CollidableWorld.CreateBody(&bdDef)

	b2CircleShape := box2d.MakeB2CircleShape()
	b2CircleShape.M_radius = 32 // Matching that of client-side setting.

	fd := box2d.MakeB2FixtureDef()
	fd.Shape = &b2CircleShape

	//fd.Filter.CategoryBits = COLLISION_CATEGORY_CONTROLLED_PLAYER
	//fd.Filter.MaskBits = COLLISION_MASK_FOR_CONTROLLED_PLAYER
	//mark
	fd.Filter.CategoryBits = 1
	fd.Filter.MaskBits = 2

	fd.Density = 0.0
	b2PlayerBody.CreateFixtureFromDef(&fd)

	client.PlayerCollidableBody = b2PlayerBody

	log.Println("Player:")
	log.Println(b2PlayerBody)

	b2PlayerBody.SetUserData(player)
	models.PrettyPrintBody(client.PlayerCollidableBody)
}

func (client *Client) InitBarrier(pTmxMapIns *models.TmxMap, pTsxIns *models.Tsx) {
	gravity := box2d.MakeB2Vec2(0.0, 0.0)
	world := box2d.MakeB2World(gravity)
	world.SetContactFilter(&box2d.B2ContactFilter{})
	client.CollidableWorld = &world
	for _, lay := range pTmxMapIns.Layers {
		if lay.Name != "tile_1 human skeleton" && lay.Name != "tile_1 board" && lay.Name != "tile_1 stone" {
			continue
		}
		fmt.Println("InitBarrier:")
		fmt.Println(lay.Name, len(lay.Tile))
		counter := 0
		for index, tile := range lay.Tile {
			counter = counter + 1
			if counter > 20 {
				break
			}

			fmt.Printf("tile: %v \n", tile)
			if tile == nil || tile.Tileset == nil {
				continue
			}
			if tile.Tileset.Source != "tile_1.tsx" {
				continue
			}

			barrier := &models.Barrier{}
			barrier.X, barrier.Y = pTmxMapIns.GetCoordByGid(index)
			barrier.Type = tile.Id
			if v, ok := pTsxIns.BarrierPolyLineList[int(tile.Id)]; ok {
				thePoints := make([]*models.Vec2D, 0)
				for _, p := range v.Points {
					thePoints = append(thePoints, &models.Vec2D{
						X: p.X,
						Y: p.Y,
					})
				}
				barrier.Boundary = &models.Polygon2D{Points: thePoints}
			}

			var bdDef box2d.B2BodyDef
			bdDef = box2d.MakeB2BodyDef()
			bdDef.Type = box2d.B2BodyType.B2_staticBody
			bdDef.Position.Set(barrier.X, barrier.Y) // todo ？？？？？
			b2EmelementBody := client.CollidableWorld.CreateBody(&bdDef)

			fd := box2d.MakeB2FixtureDef()
			if barrier.Boundary != nil {
				b2Vertices := make([]box2d.B2Vec2, len(barrier.Boundary.Points))
				for vIndex, v2 := range barrier.Boundary.Points {
					b2Vertices[vIndex] = v2.ToB2Vec2()
				}
				b2PolygonShape := box2d.MakeB2PolygonShape()
				b2PolygonShape.Set(b2Vertices, len(barrier.Boundary.Points))
				fd.Shape = &b2PolygonShape
				fd.Filter.CategoryBits = COLLISION_CATEGORY_BARRIER
				fd.Filter.MaskBits = COLLISION_MASK_FOR_BARRIER
				fd.Density = 0.0
			} else {
				b2CircleShape := box2d.MakeB2CircleShape()
				b2CircleShape.M_radius = 32
				fd.Shape = &b2CircleShape
				fd.Filter.CategoryBits = COLLISION_CATEGORY_CONTROLLED_PLAYER
				fd.Filter.MaskBits = COLLISION_MASK_FOR_CONTROLLED_PLAYER
				fd.Density = 0.0
			}

			//mark
			fd.Filter.CategoryBits = 2
			fd.Filter.MaskBits = 1

			b2EmelementBody.CreateFixtureFromDef(&fd)

			barrier.CollidableBody = b2EmelementBody
			b2EmelementBody.SetUserData(barrier)
			client.Barrier[int32(index)] = barrier
		}
		fmt.Println(client.Barrier)
	}
}

func ErrFatal(err error) {
	if err != nil {
		log.Fatal("ErrFatal", zap.NamedError("err", err))
	}
}

func clean() {
	log.Println("About to clean up the resources occupied by this server-process.")
}
