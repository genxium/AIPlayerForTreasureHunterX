package main

import(
  "fmt"
	"AI/models"
	"github.com/ByteArena/box2d"
	"AI/astar"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main(){
  tmx, tsx := models.InitMapStaticResource("./map/map/pacman/map.tmx");
	gravity := box2d.MakeB2Vec2(0.0, 0.0);
  world := box2d.MakeB2World(gravity);

  models.CreateBarrierBodysInWorld(&tmx, &tsx, &world);


  tmx.CollideMap = models.CollideMap(tmx.World, &tmx);
  //models.SignItemPosOnMap(&tmx);

  //Test
  start := astar.Point{
    X: 44,
    Y: 5,
  }

  goal := astar.Point{
    X: 18,
    Y: 41,
  }
  //Test

  path := models.FindPathByStartAndGoal(tmx.CollideMap, start, goal);
  fmt.Println(path)
}
