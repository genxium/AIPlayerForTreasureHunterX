package main

import (
	"AI/models"
	"fmt"
	//"AI/astar"
	"github.com/ByteArena/box2d"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main() {
	tmx, tsx := models.InitMapStaticResource("./map/map/pacman/map.tmx")
	gravity := box2d.MakeB2Vec2(0.0, 0.0)
	world := box2d.MakeB2World(gravity)

	models.CreateBarrierBodysInWorld(&tmx, &tsx, &world)

	theMap := models.CollideMap(tmx.World, &tmx)
	fmt.Println(theMap)
}
