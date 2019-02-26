package main

import(
  "fmt"
	"AI/models"
	"path/filepath"
	"os"
	"io/ioutil"
)

func main(){
  //fmt.Println();
  initMapStaticResource()
}

func initMapStaticResource() error{

	relativePath := "./map/map/kobako_test.tmx"
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
	models.DeserializeToTsxIns(byteArr, pTsxIns)

	//client.InitBarrier(pTmxMapIns, pTsxIns)
  return nil;
  //return tmxMapIns;
}
