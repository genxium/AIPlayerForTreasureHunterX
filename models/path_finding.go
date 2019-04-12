package models

import (
	"AI/astar"
	"math"
	"fmt"
)

//寻路的抽象

//States
const (
  UnReady = 0
  CollideMapPrepared = 1
  TreasureMapPrepared = 2
)

type PathFinding struct {
	CollideMap astar.Map
	CurrentCoord Vec2D //当前玩家坐标
	CoordPath []Vec2D //离散的路径转换成连续路径
	PointPath []astar.Point //寻路得到的离散路径
	NextGoalIndex int //-1表示没有下一个可走的点
	TreasureMap map[int32]Point //id -> point of position
  TargetTreasureId int32 //用于判断这个宝物是否已经被吃掉

  State int
}

func (p *PathFinding) transitState(state int){
  fmt.Printf("PathFinding transit, from state%d to state%d \n", p.State, state)
  p.State = state
}

func (p *PathFinding) SetCollideMap(collideMap astar.Map){
  p.CollideMap = collideMap
  p.transitState(CollideMapPrepared)
}

func (p *PathFinding) SetTreasureMap(treasureDiscreteMap map[int32]Point){
  p.TreasureMap = treasureDiscreteMap
  p.transitState(TreasureMapPrepared)
}

func (p *PathFinding) Move(step float64){
	if p.NextGoalIndex >= len(p.CoordPath) {
    //已经移动到最后一个点
	} else {
		eps := step / 2

		//tarPos := walkInfo.Path[walkInfo.CurrentTarIndex]
		tarPos := p.CoordPath[p.NextGoalIndex]
		//curPos := walkInfo.CurrentPos
		curPos := p.CurrentCoord

		dy := tarPos.Y - curPos.Y
		dx := tarPos.X - curPos.X

		var stepX float64
		var stepY float64
		if dx == 0 {
			if dy < 0 {
				stepY = -step
			} else {
				stepY = step
			}
		} else {
			radian := math.Abs(math.Atan(dy / dx))
			stepX = step * math.Cos(radian)
			stepY = step * math.Sin(radian)
			if dx < 0 {
				stepX = -stepX
			}
			if dy < 0 {
				stepY = -stepY
			}
		}

		//fmt.Println(stepX, stepY);

		nextPos := Vec2D{
			X: curPos.X + stepX,
			Y: curPos.Y + stepY,
		}

		//fmt.Println(nextPos);

		d := Distance(&nextPos, &tarPos)
		//fmt.Println(d);

		if d < eps {
			p.CurrentCoord = tarPos
			p.NextGoalIndex = p.NextGoalIndex + 1
		} else {
      p.CurrentCoord = nextPos
		}
	}
}

func (p *PathFinding) SetCurrentCoord(x float64, y float64){
  p.CurrentCoord.X = x
  p.CurrentCoord.Y = y
}


func (p *PathFinding) FindPointPath(startPoint astar.Point, endPoint astar.Point) []astar.Point{
  p.PointPath = FindPathByStartAndGoal(p.CollideMap, startPoint, endPoint)
	//fmt.Printf("The point path: %v \n", p.PointPath)
  return p.PointPath
}

func (p *PathFinding) SetNewCoordPath(coordPath []Vec2D){
  p.CoordPath = coordPath
	if len(coordPath) < 1 {
		fmt.Println("There is no path to the goal")
    p.NextGoalIndex = -1
	} else {
    p.NextGoalIndex = 1 //不为0的原因是0为当前坐标
	}
}

func (p *PathFinding) UpdateTargetTreasureId(id int32){
  //fmt.Printf("目标宝物%d被吃掉了, 换个新的目标%d \n", p.TargetTreasureId, id)
  p.TargetTreasureId = id
}


