package astar

import (
  "fmt"
  "math"
  "errors"
  mapset "github.com/deckarep/golang-set"
  . "github.com/logrusorgru/aurora"
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

func AstarArrayToMap(array []int, width int, height int) Map{
  theMap := make([][]int, height)
  for i := 0;i<height; i++ {
    theMap[i] = make([]int, width)
    for j:=0;j<width;j++{
      theMap[i][j] = array[i*width + j]
    }
  }
  return theMap
}

func minimum(openSet mapset.Set, fScore map[string]float64) (string, Point){
  min := math.MaxFloat64;
  key := "";
  point := Point{};
  for i := range openSet.Iterator().C{
    if pt, ok := i.(Point); ok {
      score := fScore[hash(pt)];
      if score <= min{
        min = score;
        key = hash(pt);
        point = pt;
      }
    }
  }
  //fmt.Println("minimun",key, point);
  return key, point;
}

func heuristicCostEstimate(pt1 Point, pt2 Point) float64{
  return distBetween(pt1, pt2);
}

func (pt Point) nabors() []Point{
  result := []Point{
    Point{pt.X - 1, pt.Y},
    Point{pt.X + 1, pt.Y},
    Point{pt.X, pt.Y - 1},
    Point{pt.X, pt.Y + 1},
    Point{pt.X - 1, pt.Y - 1},
    Point{pt.X - 1, pt.Y + 1},
    Point{pt.X + 1, pt.Y - 1},
    Point{pt.X + 1, pt.Y + 1},
  };
  return result;
}

func (pt1 Point) equal(pt2 Point) bool{
  return pt1.X == pt2.X && pt1.Y == pt2.Y;
}

func distBetween(pt1 Point, pt2 Point) float64{
  return math.Sqrt(math.Pow(float64(pt1.X - pt2.X), 2.0) + math.Pow(float64(pt1.Y - pt2.Y), 2.0));
}

func reconstructPath(cameFrom map[string]Point, current Point, path []Point) []Point{
  path = append(path, current);
  for true {
    parent, ok := cameFrom[hash(current)];
    if ok {
      current = parent;
      path = append(path, current);
    }else{
      break;
    }
  }
  //fmt.Println(path);
  return path;
}

func isBarrier(m Map, pt Point) bool{
  return pt.X < 0 || pt.Y < 0  || len(m) <= pt.Y || len(m[pt.Y]) <= pt.X || m[pt.Y][pt.X] == BARRIER;
}

func AstarByMap(m Map) []Point{
  start,_ := findPoint(m, START);
  goal, _ := findPoint(m, GOAL);

  fmt.Printf("Astar start: start at: %v, goal at: %v", start, goal);

  openSet := mapset.NewSet(start);
  closeSet := mapset.NewSet();
  gScore := map[string]float64{hash(start): 0};
  fScore := map[string]float64{hash(start): 0 + heuristicCostEstimate(start, goal)};
  cameFrom := map[string]Point{};

  count := 0;
  path := []Point{};
  var err error;
  for openSet.Cardinality() > 0{
    count = count + 1;
    if count > 3000{
      err = errors.New("Had tried too many times, there may be some logic error in your code!");
      break;
    }

    currentKey, current := minimum(openSet, fScore);
    if current.equal(goal){
      fmt.Println("Reach Goal");
      path = reconstructPath(cameFrom, current, path);
      //fmt.Println(cameFrom);
      break;
    }

    openSet.Remove(current);
    closeSet.Add(current);

    //Get nabors of minPoint then add them to the openset and update their fScore and gScore
    for _, nabor := range current.nabors(){
      if isBarrier(m, nabor){
        //Continue
      }else{
        naborKey := hash(nabor);
        if !closeSet.Contains(nabor) {
          tentativeGScore := gScore[currentKey] + distBetween(nabor, current);
          if !openSet.Contains(nabor){
            openSet.Add(nabor);
            gScore[naborKey] = tentativeGScore;
            fScore[naborKey] = tentativeGScore + heuristicCostEstimate(nabor, goal);
            cameFrom[naborKey] = current;
          }else{
            if(tentativeGScore < gScore[naborKey]){
              gScore[naborKey] = tentativeGScore;
              fScore[naborKey] = tentativeGScore + heuristicCostEstimate(nabor, goal);
              cameFrom[naborKey] = current;
            }
          }
        }
      }
    }//end for
  }

  if err != nil{
    fmt.Println(err);
  }else{
    //fmt.Println("path: ", path);
  }

  //reverse
  for i, j := 0, len(path)-1; i < j; i, j = i+1, j-1 {
    path[i], path[j] = path[j], path[i]
  }
  return path;
}

func PrintMap(m Map){
  for row := range m{
    for col := range m[row]{
      switch num := m[row][col]; num {
      case 1:
        fmt.Printf( "%d ", Red(num));
      case 9:
        fmt.Printf( "%d ", Green(num));
      default:
        fmt.Printf( "%d ", num);
      }
    }
    fmt.Println();
  }
  m[0][0] = 99;
}

func PrintArray(array []int,width int, height int){
  for row := 0; row < height; row++ {
    for col := 0; col < width; col++ {
      switch num := array[row * width + col]; num {
      case 1:
        fmt.Printf( "%d ", Red(num));
      case 9:
        fmt.Printf( "%d ", Green(num));
      default:
        fmt.Printf( "%d ", num);
      }
    }
    fmt.Println();
  }
}
