package main

import (
	"AI/models"
	"fmt"
	"github.com/ByteArena/box2d"
	//"AI/astar"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main() {
	//tmx, tsx := models.InitMapStaticResource("./map/map/treasurehunter.tmx");
	tmx, tsx := models.InitMapStaticResource("./map/map/pacman/map.tmx")

	gravity := box2d.MakeB2Vec2(0.0, 0.0)
	world := box2d.MakeB2World(gravity)

	models.CreateBarrierBodysInWorld(&tmx, &tsx, &world)
	//fmt.Println("There are %d barriers", len(barriers))

	tmx.CollideMap = models.CollideMap(&world, &tmx)
	models.SignItemPosOnMap(&tmx)

	tmx.Path = models.FindPath(&tmx)

	fmt.Println("Complete")
}
