package main

import(
  "fmt"
	"AI/models"
	//"AI/astar"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main(){
  tmx, tsx := models.InitMapStaticResource("./map/map/treasurehunter.tmx");

  barriers := models.InitBarriers2(&tmx, &tsx);
  fmt.Println("There are %d barriers", len(barriers))

  tmx.PathFindingMap = models.CollideMap(tmx.World, &tmx);
  models.SignItemPosOnMap(&tmx);

  tmx.Path = models.FindPath(&tmx);



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
