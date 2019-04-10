package main

import (
	"AI/models"
	"AI/astar"
	"AI/constants"
	"AI/login"
	"encoding/json"
	//"flag"
	"fmt"
	"github.com/ByteArena/box2d"
	"github.com/golang/protobuf/proto"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
	"io/ioutil"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"time"
	"math"
  "strconv"
  //"sync"
  "github.com/gin-gonic/gin"
)

const (
	WAITING       = 0
	IN_BATTLE     = 1
	IN_SETTLEMENT = 2
	IN_DISMISSAL  = 3
	uniformTimeStepSeconds = 1.0 / 60.0
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
	COLLISION_MASK_FOR_BARRIER           = (COLLISION_CATEGORY_BARRIER)
  COLLISION_MASK_FOR_PUMPKIN           = (COLLISION_CATEGORY_PUMPKIN)
	COLLISION_MASK_FOR_SPEED_SHOES       = (COLLISION_CATEGORY_CONTROLLED_PLAYER)
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

type Direction struct{
  Dx          float64
  Dy          float64
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
  AstarMap              astar.Map

  Radian                float64
  Dir                   Direction

  TmxIns                *models.TmxMap
  WalkInfo              models.WalkInfo
  RemovedTreasuresNum   int
	Treasures             map[int32]*models.Treasure //接收到第一帧的时候初始化全部宝物
  LastFrameTreasureNum  int //上一帧时宝物的数量(因为现在每当一个宝物被吃掉时, 后端downFrame.Treasures会带上它的信息,保存该参数用于判断有没有宝物被吃掉)
  TargetTreasureId      int32 //仅当目标宝物被吃掉时重新寻路

  //寻路抽象(Incomplete) --kobako
  pathFinding           *models.PathFinding
}

func spawnBot(botName string, expectedRoomId int, botManager *models.BotManager){
  defer botManager.ReleaseBot(botName)

	log.SetFlags(0)

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: constants.HOST + ":" + constants.PORT, Path: "/tsrht"}
	q := u.Query()

  //TODO: Error handle
  intAuthToken, playerId := login.GetIntAuthTokenByBotName(botName)

  //local
	q.Set("intAuthToken", intAuthToken)
  if(expectedRoomId > 0){
	  q.Set("expectedRoomId", strconv.Itoa(expectedRoomId))
  }
	u.RawQuery = q.Encode()
	//ref to the NewClient and DefaultDialer.Dial https://github.com/gorilla/websocket/issues/54
	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {

    defer func(){
      if r := recover(); r != nil{
        fmt.Println("kobako: Panic hadled!!!!!!, 但是正常退出")
      }else{
        fmt.Println("kobako: 正常退出goroutine")
      }
      close(done)
    }()

		client := &Client{
			LastRoomDownsyncFrame: nil,
			BattleState:           -1,
			c:                     c,
			Player:                &models.Player{Id: int64(playerId)},
			Barrier:               make(map[int32]*models.Barrier),
      Radian:                math.Pi / 2,
      Dir:                   Direction{Dx: 0, Dy: 1},
      pathFinding:           new(models.PathFinding),
		}

    //初始化地图资源
    tmx, _ := models.InitMapStaticResource("./map/map/pacman/map.tmx");
    client.TmxIns = &tmx
  
  	gravity := box2d.MakeB2Vec2(0.0, 0.0);
    world := box2d.MakeB2World(gravity);

    client.CollidableWorld = &world;
  
    models.CreateBarrierBodysInWorld(&tmx, &world);

    tmx.CollideMap = models.InitCollideMap(tmx.World, &tmx);
    client.pathFinding.CollideMap = tmx.CollideMap;



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
					log.Println("marshal respPb:", err)
				}
				client.decodeProtoBuf(respPb.Data)
				client.controller()
        client.checkReFindPath()//kobako
				client.upsyncFrameData()
			} else {
				//handleHbRequirements(resp)
			}
			//time.Sleep(time.Duration(int64(20)))
		}
	}()

	for {
		select {
		case <-done:
      fmt.Println("All done. ")
			return
		case <-interrupt:
			log.Println("!!!!!!!!!!!!!!!!!!!!!!!!!!!!!interrupt")
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


func main() {

  startServer(15351)

}

func startServer(port int){
  var botManager *models.BotManager
  {
    botManager = new(models.BotManager)
    botManager.SetBots([]string{"bot1","bot2","bot3","bot4"})
  }


  r := gin.Default()
  { //启动服务器
  	r.GET("/spawnBot", func(c *gin.Context) {
      expectedRoomId, err := strconv.Atoi(c.Query("expectedRoomId"))
      if err != nil {
        fmt.Println("请求中没有或者转换expectedRoomId出错")
        c.JSON(200, gin.H{
  			  "ret": 1001,
          "err": "请求中没有或者转换expectedRoomId出错",
  		  })
      }else{
        botName,err := botManager.GetLeisureBot()
        if err != nil {
          fmt.Println("获取空闲bot出错: " + err.Error())
      		c.JSON(200, gin.H{
            "ret": 1001,
     			  "botName": "获取空闲bot出错: " + err.Error(),
     		  })
        }else{
          go spawnBot(botName, expectedRoomId, botManager)
          fmt.Printf("Get bot: %s, expectedRoomId: %d \n", botName, expectedRoomId)
      		c.JSON(200, gin.H{
            "ret": 1000,
     			  "botName": botName,
     		  })
        }
      }
  	})
  }
  r.Run(":" + strconv.Itoa(port))
}

//通过当前玩家的坐标, 和treasureMap来计算start, end point, 用于寻路, 重新初始化walkInfo
func reFindPath(tmx *models.TmxMap, client *Client){
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
    for id,v := range client.Treasures{
      treasureVec := models.Vec2D{
        X: v.X,
        Y: v.Y,
      }
      dist := models.Distance(treasureVec, playerVec)
      if dist < min{
        min = dist
        temp := tmx.CoordToPoint(treasureVec)
        endPoint = astar.Point{
          X: temp.X,
          Y: temp.Y,
        }
        client.TargetTreasureId = id
      }
    }
  }

  fmt.Printf("NEW END POINT %v , NEW TID %d \n", endPoint, client.TargetTreasureId)

  tmx.Path = models.FindPathByStartAndGoal(tmx.CollideMap, startPoint, endPoint);

  fmt.Printf("TMX path: %v", tmx.Path)

  //将离散的路径转为连续坐标, 初始化walkInfo, 每次controller的时候调用
  var path []models.Vec2D;
  for _, pt := range tmx.Path{
    gid := pt.Y * tmx.Width + pt.X;
    x, y := tmx.GetCoordByGid(gid);
    path = append(path, models.Vec2D{
      X: x,
      Y: y,
    })
  }
  fmt.Println("The coord path: ", path);

  if(len(path) < 1){
    fmt.Println("ERRRRRRRRRRRRROR, Find path failed")
    client.WalkInfo = models.WalkInfo{
      Path: path,
      CurrentTarIndex: 0,
    }
  }else{
    client.WalkInfo = models.WalkInfo{
      Path: path,
      CurrentPos: path[0],
      CurrentTarIndex: 1,
    }
  }
}


func (client *Client) initItemAndPlayers(){

  //根据第一帧的数据来设置好玩家的位置, 以及宝物的位置,以服务器为准
  initFullFrame := client.LastRoomDownsyncFrame
  //fmt.Printf("InitFullFrame: Id: %d, RefFrameId: %d, Treasures: %v", initFullFrame.Id, initFullFrame.RefFrameId, initFullFrame.Treasures)

  //Init treasures
  client.Treasures =  initFullFrame.Treasures
  client.LastFrameTreasureNum = 0


  //Sign on map
  tmx := client.TmxIns

  {//Init ContinuousPosMap
    //将离散的点转换成连续的点, 用于确认道具的位置(遍历每个点判断相对距离最短)
    var continuousPosMap [][]models.Vec2D
    continuousPosMap = make([][]models.Vec2D, tmx.Height)
    for i:=0; i<tmx.Height; i++{
      continuousPosMap[i] = make([]models.Vec2D, tmx.Width)
    }
    for i:=0; i<tmx.Height; i++ {
      for j:=0; j<tmx.Width; j++ {
        gid := i * tmx.Width + j
        x,y := tmx.GetCoordByGid(gid)
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
    for _, treasure := range client.Treasures{
      coord := models.Vec2D{
        X: treasure.X,
        Y: treasure.Y,
      }
      discretePoint := tmx.CoordToPoint(coord)
      treasureDiscreteMap[treasure.Id] = discretePoint

      //fmt.Printf("Treasure: %v \n", treasureDiscreteMap[treasure.Id])
    }
  }
  client.pathFinding.TreasureMap = treasureDiscreteMap

  fmt.Printf("INIT Treasure: %v \n", client.Treasures)

  //mark
  var playerPoint astar.Point
  {
    temp := tmx.CoordToPoint(models.Vec2D{
      X: client.Player.X,
      Y: client.Player.Y,
    })
    client.pathFinding.CurrentPoint = temp
    //类型转换
    playerPoint = astar.Point{
      X: temp.X,
      Y: temp.Y,
    }
  }



  fmt.Printf("++++++ player point: %v \n", playerPoint)

  //找出最近的一个宝物, 标记为client.TargetTreasureId
  minDistance := 99999.0
  for k, v := range treasureDiscreteMap{
    vPoint := astar.Point{
      X: v.X,
      Y: v.Y,
    }
    dist := astar.DistBetween(vPoint, playerPoint)
    if(dist < minDistance){
      minDistance = dist
      client.TargetTreasureId = k
    }
  }

  fmt.Printf("++++++ minDistance %f, %v \n",minDistance  ,playerPoint)



  reFindPath(tmx, client);

  //初始化client.WalkInfo

}

func (client *Client) checkReFindPath(){
  if client.LastRoomDownsyncFrame.RefFrameId!=0 && len(client.LastRoomDownsyncFrame.Treasures) != client.LastFrameTreasureNum{
    fmt.Printf("last number %d, now number %d \n",client.LastFrameTreasureNum,  len(client.LastRoomDownsyncFrame.Treasures))
    client.LastFrameTreasureNum = len(client.LastRoomDownsyncFrame.Treasures)


    var needReFindPath bool
    for id, _ := range client.LastRoomDownsyncFrame.Treasures{
      delete(client.Treasures, id)
      //fmt.Printf("DELETE Treasure: %d \n ", id)
      if id == client.TargetTreasureId{
        fmt.Printf("!!!!!!!!!!!!!!!! Refind path, id: %d, treasure length: %d, client.TargetTreasureId: %d \n", id, len(client.LastRoomDownsyncFrame.Treasures), client.TargetTreasureId)
        needReFindPath = true
      }
    }
    if needReFindPath {
      reFindPath(client.TmxIns, client)
    }
    //fmt.Println("++++++++++ pick up treasure")
  }else{
    //Do nothing
  }
}

func (client *Client) controller() {
	if client.Player.Speed == 0 {
		return
	}
	if client.LastRoomDownsyncFrame.Id == 1 || client.LastRoomDownsyncFrame.Id == 2 { // 初始帧
		log.Println("Game Start")
		client.BattleState = IN_BATTLE
    //初始化需要寻找的宝物和玩家位置
		client.InitPlayerCollider()
    client.initItemAndPlayers()
    fmt.Printf("Receive id: %d, treasure length %d, refId: %d \n", client.LastRoomDownsyncFrame.Id, len(client.LastRoomDownsyncFrame.Treasures), client.LastRoomDownsyncFrame.RefFrameId)
	} else {
    step := 16.0;

    pathFindingMove(client, step);

    //client :q

    //foolMove(client, step);

		//time.Sleep(time.Duration(int64(40)))
	}

}

//撞墙转向
func foolMove(client *Client, step float64){
  nowRadian := client.Radian;

  for nowRadian - client.Radian < math.Pi * 2 {
    xStep := step * math.Cos(nowRadian);
    yStep := step * math.Sin(nowRadian);
    //fmt.Println(xStep, yStep);

    //移动collideBody
		newB2Vec2Pos := box2d.MakeB2Vec2(client.Player.X + xStep, client.Player.Y - yStep);
		//newB2Vec2Pos := box2d.MakeB2Vec2(client.Player.X, client.Player.Y - yStep);
		models.MoveDynamicBody(client.PlayerCollidableBody, &newB2Vec2Pos, 0);

    //world.Step
    client.CollidableWorld.Step(uniformTimeStepSeconds, uniformVelocityIterations,uniformPositionIterations)

    //碰撞检测
    collided := false;
		for edge := client.PlayerCollidableBody.GetContactList(); edge != nil; edge = edge.Next {
			if edge.Contact.IsTouching() {
        collided = true;
        break;
				//log.Println("player conteact")
				if _, ok := edge.Other.GetUserData().(*models.Barrier); ok {
					//log.Println("player conteact to the barrier")
				}
			}
		}

    if(!collided){ //一直走
      client.Player.X = client.Player.X + xStep;
      client.Player.Y = client.Player.Y - yStep;

      //kobako
      //TODO: set correct direction
      dx, dy := func() (dx float64, dy float64){
        floorRadian := nowRadian - math.Pi * 2 * math.Floor(nowRadian / (2 * math.Pi)); 
        //fmt.Println(floorRadian);
        if floorRadian < math.Pi / 2 { return 2, -1; }else if floorRadian < math.Pi{
          return -2, -1;
        }else if floorRadian < math.Pi * 3 / 2{
          return -2, 1;
        }else {
          return 2, 1;
        }
      }()
      client.Dir = Direction{
        Dx: dx,
        Dy: dy,
      }
      //fmt.Println(dx, dy)
      //kobako
      break;
    }else{//转向
      log.Println("player collided with barriers & change direction: ", nowRadian);
      nowRadian = nowRadian + math.Pi / 16;

    }
  }

  client.Radian = nowRadian;
}

func pathFindingMove(client *Client, step float64){
  //通过服务器位置进行修正
  client.WalkInfo.CurrentPos.X = client.Player.X;
  client.WalkInfo.CurrentPos.Y = client.Player.Y;

  end := models.GotToGoal(step, &client.WalkInfo);
  if end {
    return
  }

  client.Player.X = client.WalkInfo.CurrentPos.X;
  client.Player.Y = client.WalkInfo.CurrentPos.Y;

  //log.Println(client.Player.X, client.Player.Y);
}

//lastPos := Position{};

func (client *Client) upsyncFrameData() {
  //if(lastPos)
  //fmt.Println(client.Player.X, client.Player.Y);
	if client.BattleState == IN_BATTLE {
		newFrame := &struct {
			Id            int64             `json:"id"`
			X             float64           `json:"x"`
			Y             float64           `json:"y"`
			//Dir           *models.Direction `json:"dir"`
			Dir           Direction         `json:"dir"`
			AckingFrameId int32             `json:"AckingFrameId"`
		}{client.Player.Id, client.Player.X, client.Player.Y, Direction{}, client.LastRoomDownsyncFrame.Id}

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
	room_downsync_frame := models.RoomDownsyncFrame{}
	err := proto.Unmarshal(message, &room_downsync_frame)
	if err != nil {
		log.Fatal(err)
	}

  /*
  fmt.Println("decodeProtoBuf(): ");
  fmt.Println(room_downsync_frame.Players);
  fmt.Println(client.Player.Id);
  */

  //fmt.Printf("Receive id: %d, treasure length %d, refId: %d \n", room_downsync_frame.Id, len(room_downsync_frame.Treasures), room_downsync_frame.RefFrameId)

  //根据最新一帧的信息设置bot玩家的新位置及方向等
	client.LastRoomDownsyncFrame = &room_downsync_frame
	client.Player.Speed = room_downsync_frame.Players[int32(client.Player.Id)].Speed
	client.Player.Dir = room_downsync_frame.Players[int32(client.Player.Id)].Dir
	client.Player.X = room_downsync_frame.Players[int32(client.Player.Id)].X
	client.Player.Y = room_downsync_frame.Players[int32(client.Player.Id)].Y


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
func (client *Client) initMapStaticResource() models.TmxMap{

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

  client.AstarMap = pTmxMapIns.CollideMap;

	client.InitBarrier(pTmxMapIns, pTsxIns)

  //kobako

  fmt.Println("Barrier");
  fmt.Println(client.Barrier);

  //kobako

  return tmxMapIns;
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
	fd.Filter.CategoryBits = 1;
	fd.Filter.MaskBits = 2;

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
    fmt.Println("InitBarrier:");
    fmt.Println(lay.Name, len(lay.Tile));
    counter := 0;
		for index, tile := range lay.Tile {
      counter = counter + 1;
      if counter > 20{
        break;
      }

      fmt.Printf("tile: %v \n", tile);
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
	    fd.Filter.CategoryBits = 2;
	    fd.Filter.MaskBits = 1;

			b2EmelementBody.CreateFixtureFromDef(&fd)

			barrier.CollidableBody = b2EmelementBody
			b2EmelementBody.SetUserData(barrier)
			client.Barrier[int32(index)] = barrier
		}
    fmt.Println(client.Barrier);
	}
}

func ErrFatal(err error) {
	if err != nil {
		log.Fatal("ErrFatal", zap.NamedError("err", err))
	}
}
