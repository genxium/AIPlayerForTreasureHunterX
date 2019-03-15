package models

import(
  "math"
  "fmt"
  "AI/astar"
)

/*
type Vec2D struct {
  X float64
  Y float64
}
*/

type WalkInfo struct {
  Path []Vec2D
  CurrentPos Vec2D
  CurrentTarIndex int
  TargetTreasureId int
}

func Distance(pt1 Vec2D, pt2 Vec2D) float64{
  dx := pt1.X - pt2.X;
  dy := pt1.Y - pt2.Y;
  return math.Sqrt(dx * dx + dy * dy);
}


func GotToGoal(step float64, walkInfo *WalkInfo) bool{
  //fmt.Println("1111111");
  //fmt.Println(walkInfo);
  if(walkInfo.CurrentTarIndex >= len(walkInfo.Path)){
    return true;
  }else{
    eps := step / 2;

    tarPos := walkInfo.Path[walkInfo.CurrentTarIndex];
    curPos := walkInfo.CurrentPos;
    dy := tarPos.Y - curPos.Y;
    dx := tarPos.X - curPos.X;

    var stepX float64;
    var stepY float64;
    if(dx == 0){
      if dy < 0{
        stepY = -step
      }else{
        stepY = step
      }
    }else{
      radian := math.Abs(math.Atan(dy / dx));
      stepX = step * math.Cos(radian);
      stepY = step * math.Sin(radian);
      if(dx < 0){
        stepX = -stepX;
      }
      if(dy < 0){
        stepY = -stepY;
      }
    }

    //fmt.Println(stepX, stepY);

    nextPos := Vec2D{
      X: curPos.X + stepX,
      Y: curPos.Y + stepY,
    }

    //fmt.Println(nextPos);

    d := Distance(nextPos, tarPos);
    //fmt.Println(d);

    if( d < eps ){
      walkInfo.CurrentPos = tarPos;
      walkInfo.CurrentTarIndex = walkInfo.CurrentTarIndex + 1;
      fmt.Println("Got to next point");
    }else{
      walkInfo.CurrentPos = nextPos;
    }

    return false;
  }
}

func AstarPathToWalkInfo(originPath []astar.Point) WalkInfo{
  var path []Vec2D;
  for _, pt := range originPath{
    //pt.X = pt.X * 64;
    //pt.Y = pt.Y * 64;
    path = append(path, Vec2D{
      X: float64(pt.X),
      Y: float64(pt.Y),
    });
  }

  walkInfo := WalkInfo{
    Path: path,
    CurrentPos: path[0],
    CurrentTarIndex: 1,
  }

  return walkInfo;
}
