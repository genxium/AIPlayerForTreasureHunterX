package models

import(
	"AI/astar"
)

//寻路的抽象

type PathFinding struct {
  CollideMap astar.Map
  //CurrentCoord Vec2D
  CurrentPoint Point
  //GoalPoint Point
  //GoalCoord Vec2D
  //CoordPath []Vec2D
  //PointPath []Point
  //StepDistance float64
  //CurrentPathIndex int
  TreasureMap map[int32] Point //id -> point of position
}

/*
func (p *PathFinding) Step(){

}

func (p *PathFinding) SetCollideMap(collideMap astar.Map){
  p.CollideMap = collideMap
}

func (p *PathFinding) GetCurrentCoord() Vec2D{
  return p.CurrentCoord
}

func (p *PathFinding) SetCurrentCoord(currentCoord Vec2D){
  p.CurrentCoord = currentCoord
  //TODO: Analyze currentPoint
}

func (p *PathFinding) NewGoal(goal Point){
  p.GoalPoint = goal
  //Refind path
  var path []Point
  {
    tempPath := FindPathByStartAndGoal(p.CollideMap, astar.Point(p.CurrentPoint), astar.Point(p.GoalPoint))
    for _, point := range tempPath{
      path = append(path, Point(point))
    }
  }
  p.PointPath = path
  //TODO: Analyse CoordPath
}
*/
