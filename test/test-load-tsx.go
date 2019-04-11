package main

import (
	"AI/models"
	"fmt"
	//"AI/astar"
	"os"
	"path/filepath"
	//"time"
	"io/ioutil"
)

func main() {
	pwd, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	tsxIns := models.Tsx{}
	pTsxIns := &tsxIns
	relativePath := "./map/map/tile_1.tsx"
	fp := filepath.Join(pwd, relativePath)
	fmt.Printf("fp == %v\n", fp)
	if !filepath.IsAbs(fp) {
		panic("Filepath must be absolute!")
	}

	byteArr, err := ioutil.ReadFile(fp)

	if err != nil {
		panic(err)
	}

	models.DeserializeToTsxIns(byteArr, pTsxIns)

	fmt.Println(pTsxIns.BarrierPolyLineList)
}
