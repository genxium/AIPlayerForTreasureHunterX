package main

import (
	"AI/astar"
	"AI/constants"
	"AI/login"
	"AI/models"
	pb "AI/pb_output"
	"context"
	"encoding/json"
	"fmt"
	"github.com/ByteArena/box2d"
	"github.com/gin-gonic/gin"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"log"
	"math"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime/debug"
	"strconv"
	"sync/atomic"
	"syscall"
	"time"
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

type HeartbeatRequirementsData struct {
	IntervalToPing        int    `json:"intervalToPing"`
	WillKickIfInactiveFor int    `json:"willKickIfInactiveFor"`
	BoundRoomId           int    `json:"boundRoomId"`
	BattleColliderInfo    []byte `json:"battleColliderInfo"`
}

type wsRespPb struct {
	Ret         int32  `json:"ret,omitempty"`
	EchoedMsgId int32  `json:"echoedMsgId,omitempty"`
	Act         string `json:"act,omitempty"`
	Data        []byte `json:"data,omitempty"`
}

type Client struct {
	Id                    int //roomId
	LastRoomDownsyncFrame *pb.RoomDownsyncFrame
	BattleState           int
	c                     *websocket.Conn
	Player                *pb.Player
	CollidableWorld       *box2d.B2World
	Barrier               map[int32]*models.Barrier
	PlayerCollidableBody  *box2d.B2Body `json:"-"`

	Radian float64
	Dir    models.Direction

	TmxIns *models.TmxMap

	//上一帧时宝物的数量(因为现在每当一个宝物被吃掉时, 后端downFrame.Treasures会带上它的信息,保存该参数用于判断有没有宝物被吃掉)
	RemovedTreasuresNum         int
	LastFrameRemovedTreasureNum int

	//寻路抽象(Incomplete) --kobako
	pathFinding *models.PathFinding
	Started     bool

	BotSpeed    *int32
	StayedCount int
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
		log.Panic("dial:", err)
	}
	defer c.Close()

	client := &Client{
		LastRoomDownsyncFrame: nil,
		BattleState:           -1,
		c:                     c,
		Player:                &pb.Player{Id: int32(playerId)},
		Barrier:               make(map[int32]*models.Barrier),
		Radian:                math.Pi / 2,
		Dir:                   models.Direction{Dx: 0, Dy: 1},
		pathFinding:           new(models.PathFinding),
		StayedCount:           0,
	}

	client.Started = false
	killSignal := int32(0)
	client.BotSpeed = new(int32)

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
			time.Sleep(time.Second / 20)
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
				log.Println("websocket read json err", err)
			}

			switch resp.Act {
			case "RoomDownsyncFrame":
				var respPb *wsRespPb
				respPb = new(wsRespPb)
				err := c.ReadJSON(respPb)
				if err != nil {
					log.Println("Err unmarshalling downsync respPb:", err)
				}
				client.decodeProtoBuf(respPb.Data)
			case "HeartbeatRequirements":
				var respPb *HeartbeatRequirementsData
				respPb = new(HeartbeatRequirementsData)
				err := json.Unmarshal(resp.Data, respPb)
				if err != nil {
					log.Println("Err unmarshalling heartbeat respPb:", err)
				}
				var battleColliderInfo pb.BattleColliderInfo
				err = proto.Unmarshal(respPb.BattleColliderInfo, &battleColliderInfo)
				if err != nil {
					log.Println("Err unmarshalling data:", err)
				}
				//初始化地图资源
				tmx := models.TmxMap{
					Width:      int(battleColliderInfo.StageDiscreteW),
					Height:     int(battleColliderInfo.StageDiscreteH),
					TileWidth:  int(battleColliderInfo.StageTileW),
					TileHeight: int(battleColliderInfo.StageTileH),
				}
				//tmx, _ := models.InitMapStaticResource("./map/map/pacman/map.tmx")
				client.TmxIns = &tmx

				log.Println("collideMap init", tmx)
				collideMap := models.InitCollideMapNeo(&tmx, battleColliderInfo.StrToPolygon2DListMap)
				client.pathFinding.SetCollideMap(collideMap)
				client.playerBattleColliderAck()
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
func reFindPath(tmx *models.TmxMap, client *Client, excludeTreasureID map[int32]bool) {
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
			notInExcluedMap := true
			if excludeTreasureID != nil {
				if _, ok := excludeTreasureID[id]; ok {
					notInExcluedMap = false
				}
			}
			if dist < min && notInExcluedMap {
				min = dist
				endPoint = treasurePoint
				client.pathFinding.UpdateTargetTreasureId(id)
			}
		}
	}

	//fmt.Printf("NEW END POINT %v , NEW TID %d \n", endPoint, client.pathFinding.TargetTreasureId)

	//fmt.Printf("++++++ start point %v, end point %v\n", startPoint, endPoint)
	//fmt.Printf("++++++ current point %v\n", client.pathFinding.CurrentCoord)
	pointPath := client.pathFinding.FindPointPath(startPoint, endPoint)
	fmt.Printf("The point path: %v\n", pointPath)

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

	reFindPath(tmx, client, nil)
}

func (client *Client) checkReFindPath() {
	// 仅当 (当前帧的宝物数量比上一帧少 && 目标宝物id被吃掉)  的时候重新寻路
	if client.LastRoomDownsyncFrame == nil || client.LastRoomDownsyncFrame.RefFrameId == 0 || client.BattleState != IN_BATTLE {
		return
	}
	var needReFindPath = false
	if len(client.LastRoomDownsyncFrame.Treasures) != client.LastFrameRemovedTreasureNum {
		//fmt.Printf("last number %d, now number %d \n", client.LastFrameRemovedTreasureNum, len(client.LastRoomDownsyncFrame.Treasures))
		client.LastFrameRemovedTreasureNum = len(client.LastRoomDownsyncFrame.Treasures)

		for id, _ := range client.LastRoomDownsyncFrame.Treasures {
			//删除以减轻后续最短距离计算量
			delete(client.pathFinding.TreasureMap, id)
			if id == client.pathFinding.TargetTreasureId {
				needReFindPath = true
			}
		}
	}

	var excludeTreasureID map[int32]bool
	// 防止server漏判吃草导致挂机
	if !needReFindPath && atomic.LoadInt32(client.BotSpeed) > 0 &&
		(client.pathFinding.NextGoalIndex >= len(client.pathFinding.CoordPath) || client.StayedCount > 20) {
		//fmt.Println("prevent stop by not eat treasure")
		excludeTreasureID = make(map[int32]bool)
		needReFindPath = true
		excludeTreasureID[client.pathFinding.TargetTreasureId] = true
		if client.StayedCount > 20 {
			fmt.Println("prevent stop by StayedCount")
			client.StayedCount = 0
		}
	}
	//p.NextGoalIndex >= len(p.CoordPath)

	if needReFindPath {
		reFindPath(client.TmxIns, client, excludeTreasureID)
		retryCount := 0
		if excludeTreasureID == nil {
			excludeTreasureID = make(map[int32]bool)
		}
		for retryCount < 5 && client.pathFinding.NextGoalIndex == -1 {
			retryCount = retryCount + 1
			excludeTreasureID[client.pathFinding.TargetTreasureId] = true
			reFindPath(client.TmxIns, client, excludeTreasureID)
		}
	}
}

func (client *Client) controller() {
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
		client.initTreasureAndPlayers()
		fmt.Printf("Init coord: %.2f, %.2f\n", client.Player.X, client.Player.Y)
		client.pathFinding.SetCurrentCoord(client.Player.X, client.Player.Y)
		fmt.Printf("Receive id: %d, treasure length %d, refId: %d \n", client.LastRoomDownsyncFrame.Id, len(client.LastRoomDownsyncFrame.Treasures), client.LastRoomDownsyncFrame.RefFrameId)
	} else {
		step := float64(atomic.LoadInt32(client.BotSpeed)) / 20
		pathFindingMove(client, step)
	}

}

func pathFindingMove(client *Client, step float64) {
	client.pathFinding.Move(step)
	if client.BattleState == IN_BATTLE &&
		client.Player.X == client.pathFinding.CurrentCoord.X &&
		client.Player.Y == client.pathFinding.CurrentCoord.Y {
		client.StayedCount++
	}
	client.Player.X = client.pathFinding.CurrentCoord.X
	client.Player.Y = client.pathFinding.CurrentCoord.Y
	client.pathFinding.SetCurrentCoord(client.Player.X, client.Player.Y)
}

//lastPos := Position{};

func (client *Client) upsyncFrameData() {
	if client.BattleState == IN_BATTLE {
		newFrame := &struct {
			Id            int32            `json:"id"`
			X             float64          `json:"x"`
			Y             float64          `json:"y"`
			Dir           models.Direction `json:"dir"`
			AckingFrameId int32            `json:"AckingFrameId"`
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

func (client *Client) playerBattleColliderAck() {
	req := &wsReq{
		MsgId: 1,
		Act:   "PlayerBattleColliderAck",
	}
	reqByte, err := json.Marshal(req)
	err = client.c.WriteMessage(websocket.TextMessage, reqByte)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

//kobako: 从下行帧解析宝物信息是否减少
func (client *Client) decodeProtoBuf(message []byte) {
	roomDownSyncFrame := pb.RoomDownsyncFrame{}
	err := proto.Unmarshal(message, &roomDownSyncFrame)
	if err != nil {
		fmt.Println("解析room_downsync_frame出错了!")
		log.Panic(err)
	}
	client.LastRoomDownsyncFrame = &roomDownSyncFrame
	atomic.StoreInt32(client.BotSpeed, roomDownSyncFrame.Players[int32(client.Player.Id)].Speed)

}

func ErrFatal(err error) {
	if err != nil {
		log.Fatal("ErrFatal", zap.NamedError("err", err))
	}
}

func clean() {
	log.Println("About to clean up the resources occupied by this server-process.")
}
