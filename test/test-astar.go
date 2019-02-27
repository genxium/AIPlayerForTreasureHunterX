package main

import (
  "AI/astar"
)

func main(){

  theMap := astar.Map{
    {2, 0, 0, 0, 0},
    {0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0},
    {0, 0, 0, 0, 0},
    {0, 0, 0, 3, 0},
  }

  astar.AstarByMap(theMap);

}
