package main

import (
	"AI/models"
	"fmt"
	//"AI/astar"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main() {
	tmx, _ := models.InitMapStaticResource("./map/map/kobako_test2.tmx")

	fmt.Println(tmx.GetCoordByGid(12))
	fmt.Println(tmx.GetCoordByGid(13))
	fmt.Println(tmx.GetCoordByGid(18))

	//walkInfo := models.AstarPathToWalkInfo(tmxMapIns.Path);
	//step := 300.0;

	/*
	  for {
	    end := models.GotToGoal(step, &walkInfo);
	    if end{
	      break;
	    }else{
	     time.Sleep(1 * time.Second);
	    }
	  }
	*/

}
