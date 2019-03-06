package main

import(
  "fmt"
	"AI/models"
	"AI/astar"
	//"path/filepath"
	//"os"
	//"time"
	//"io/ioutil"
)

func main(){
  //fmt.Println();
  //tmxMapIns, tsx := models.InitMapStaticResource();
  tmx, tsx := models.InitMapStaticResource();
  //tmxMapIns.PathFindingMap()

  //fmt.Println("1111111111111111111");
  //fmt.Println(tsx);

  barriers := models.InitBarriers2(&tmx, &tsx);

  fmt.Println("222222222222222222");
  fmt.Println(barriers);
  //fmt.Println(barriers);

  //fmt.Println(tsx);

  //models.InitItemsForPathFinding(&tmxMapIns);
  //models.FindPath(&tmxMapIns);


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


  theMap := models.CollideMap(tmx.World, &tmx)
  //fmt.Println(theMap)


  astar.PrintArray(theMap, tmx.Width, tmx.Height)
}


