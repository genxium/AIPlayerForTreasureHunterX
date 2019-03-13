package main

import(
  //"fmt"
	"AI/models"
	"github.com/ByteArena/box2d"
	//"AI/astar"
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
  models.SignItemPosOnMap(&tmx);

  tmx.Path = models.FindPath(&tmx);
}
