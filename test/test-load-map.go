package main

import(
  "fmt"
	"AI/models"
	"AI/astar"
	"path/filepath"
	"os"
	"io/ioutil"
)

func main(){
  //fmt.Println();
  tmxMapIns := initMapStaticResource()
  //tmxMapIns.PathFindingMap()

  //初始化奖励位置
  for _, hignTreasure := range tmxMapIns.HighTreasuresInfo{
    fmt.Println(hignTreasure.DiscretePos.Y, hignTreasure.DiscretePos.X);
    tmxMapIns.PathFindingMap[hignTreasure.DiscretePos.Y][hignTreasure.DiscretePos.X] = 3;
  }
  //初始化起点位置
  tmxMapIns.PathFindingMap[tmxMapIns.StartPoint.Y][tmxMapIns.StartPoint.X] = 2;


  fmt.Println("The Start Point: ");
  fmt.Println(tmxMapIns.StartPoint);


  path := astar.AstarByMap(tmxMapIns.PathFindingMap);
  fmt.Println(path);

  for _, pt := range path{
    tmxMapIns.PathFindingMap[pt.Y][pt.X] = 9;
  }

  astar.PrintMap(tmxMapIns.PathFindingMap);
}

func initMapStaticResource() models.TmxMap{

	//relativePath := "./map/map/kobako_test.tmx"
	relativePath := "./map/map/treasurehunter.tmx"
	execPath, err := os.Executable()
  if err != nil{
    panic(err);
  }

	pwd, err := os.Getwd()
  if err != nil{
    panic(err);
  }

	fmt.Printf("execPath = %v, pwd = %s, returning...\n", execPath, pwd)

	tmxMapIns := models.TmxMap{}
	pTmxMapIns := &tmxMapIns
	fp := filepath.Join(pwd, relativePath)
	fmt.Printf("fp == %v\n", fp)
	if !filepath.IsAbs(fp) {
		panic("Tmx filepath must be absolute!")
	}

	byteArr, err := ioutil.ReadFile(fp)
  if err != nil{
    panic(err);
  }
	models.DeserializeToTmxMapIns(byteArr, pTmxMapIns)

	tsxIns := models.Tsx{}
	pTsxIns := &tsxIns
	relativePath = "./map/map/tile_1.tsx"
	fp = filepath.Join(pwd, relativePath)
	fmt.Printf("fp == %v\n", fp)
	if !filepath.IsAbs(fp) {
		panic("Filepath must be absolute!")
	}

	byteArr, err = ioutil.ReadFile(fp)
  if err != nil{
    panic(err);
  }
	models.DeserializeToTsxIns(byteArr, pTsxIns);

	//client.InitBarrier(pTmxMapIns, pTsxIns)
  //fmt.Println("++++++++++++");
  //fmt.Println(tmxMapIns.HighTreasuresInfo);
  //return nil;
  return tmxMapIns;
}
