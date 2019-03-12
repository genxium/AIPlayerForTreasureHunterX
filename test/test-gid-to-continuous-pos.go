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
  //tmx, _ := models.InitMapStaticResource("./map/map/treasurehunter.tmx");
  //tmx, _ := models.InitMapStaticResource("./map/map/kobako_test.tmx");
  tmx, _ := models.InitMapStaticResource("./map/map/kobako_test2.tmx");

  /*
  barriers := models.InitBarriers2(&tmx, &tsx);
  fmt.Println("There are %d barriers", len(barriers))

  tmx.PathFindingMap = models.CollideMap(tmx.World, &tmx);
  models.SignItemPosOnMap(&tmx);

  tmx.Path = models.FindPath(&tmx);
  */

  /*
  fmt.Println(tmx.GetCoordByGid(5))
  fmt.Println(tmx.GetCoordByGid(6))
  fmt.Println(tmx.GetCoordByGid(9))
  fmt.Println(tmx.GetCoordByGid(10))
  fmt.Println(tmx.GetCoordByGid(11))
  */
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
