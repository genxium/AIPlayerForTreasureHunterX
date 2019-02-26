package astar

import (
  "fmt"
  "math"
  "errors"
  mapset "github.com/deckarep/golang-set"
)

type Map = [][]int; //Itself is the pointer
type Point struct {
  X int
  Y int
};


const (
  ROAD = 0
  BARRIER = 1
  START = 2
  GOAL = 3
)

func findPoint(m Map, value int) (Point, error){
  for row := range m{
    for col := range m[row]{
      if m[row][col] == value{
        return Point{X: col, Y: row}, nil;
      }
    }
  }
  return Point{X: -1, Y: -1}, errors.New("Can't find start point");
}

func hash(pt Point) string{
  return fmt.Sprintf("%d, %d", pt.X, pt.Y);
}

func minimum(openSet mapset.Set, fScore map[string]float64) (string, Point){
  min := 0.0;
  key := "";
  point := Point{};
  for i := range openSet.Iterator().C{
    if pt, ok := i.(Point); ok {
      score := fScore[hash(pt)];
      if score < min{
        min = score;
        key = hash(pt);
        point = pt;
      }
    }
  }
  return key, point;
}

func heuristicCostEstimate(pt1 Point, pt2 Point) float64{
  return math.Sqrt(math.Pow(float64(pt1.X - pt2.X), 2.0) + math.Pow(float64(pt1.Y - pt2.Y), 2.0));
}

func AstarByMap(m Map){
  start,_ := findPoint(m, START);
  goal, _ := findPoint(m, GOAL);

  openSet := mapset.NewSet(start);
  //closeSet := mapset.NewSet();
  //gScore := map[string]float64{hash(start): 0};
  fScore := map[string]float64{hash(start): 0 + heuristicCostEstimate(start, goal)};

  //fmt.Println(start, goal, openSet, closeSet, gScore, fScore);

  pt2 := Point{X: 2, Y: 2};
  openSet.Add(pt2);
  fScore[hash(pt2)] = -1;
  //fmt.Println(minimum(openSet, fScore));

  count := 0;
  for openSet.Cardinality() > 0{
    count = count + 1;
    if count > 200{
      break;
    }
    //minKey, minPoint := minimum(openSet, fScore);
  }
}
