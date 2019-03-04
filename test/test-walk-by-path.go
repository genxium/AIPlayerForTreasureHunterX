package main

import(
  "AI/models"
  "time"
  "fmt"
)

func main(){

  path := []models.AccuratePosition{
    models.AccuratePosition{X: 0, Y: 0, },
    models.AccuratePosition{X: 100, Y: 100, },
    models.AccuratePosition{X: 0, Y: 200, },
    models.AccuratePosition{X: -100, Y: 100, },
  }

  fmt.Println(path);

  step := 20.0;

  walkInfo := models.WalkInfo{
    Path: path,
    CurrentPos: path[0],
    CurrentTarIndex: 1,
  }

  for {
    end := models.GotToGoal(step, &walkInfo)
    //fmt.Println(walkInfo);
    //fmt.Println(end);
    if end{
      break;
    }else{
     time.Sleep(1 * time.Second);
    }
  }


}
