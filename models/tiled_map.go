package models

import (
	"AI/astar"
	pb "AI/pb_output"
	"github.com/Tarliton/collision2d"
	"log"
	"math"
)

// w map
type TmxMap struct {
	Width      int
	Height     int
	TileWidth  int
	TileHeight int
	//kobako
	ContinuousPosMap [][]Vec2D
}

type Point struct {
	X int
	Y int
}

type TreasuresInfo struct {
	InitPos Vec2D
	Type    int32
	Score   int32
	//DiscretePos Point
	DiscretePos Point
}

type SpeedShoesInfo struct {
	InitPos Vec2D
	Type    int32
}

type TmxData struct {
	Encoding    string `xml:"encoding,attr"`
	Compression string `xml:"compression,attr"`
	Value       string `xml:",chardata"`
}

func (m *TmxMap) GetCoordByGid(index int) (x float64, y float64) {
	h := index / m.Width
	w := index % m.Width
	x = float64(w*m.TileWidth) + 0.5*float64(m.TileWidth)
	y = float64(h*m.TileHeight) + 0.5*float64(m.TileHeight)
	tmp := &Vec2D{x, y}
	vec2 := m.continuousObjLayerVecToContinuousMapNodeVec(tmp)
	return vec2.X, vec2.Y
}

/**
 *  TODO: 现在的做法是遍历整个地图, 选取距离最近的一个离散点, 正确的做法需要
 *  从cocos解析器搬运过来
 */
func (tmx *TmxMap) CoordToPoint(coord Vec2D) Point {
	var minDistance float64 = 9999999

	var result Point = Point{
		X: -1,
		Y: -1,
	}

	/*
	  //fmt.Println(tmx.ContinuousPosMap)
	  fmt.Println(tmx.ContinuousPosMap)
	  fmt.Println(tmx.Width, tmx.Height)
	*/

	for i := 0; i < tmx.Height; i++ {
		for j := 0; j < tmx.Width; j++ {
			tilePos := tmx.ContinuousPosMap[i][j]
			distance := Distance(&coord, &tilePos)
			if distance < minDistance {
				minDistance = distance
				result.X = j
				result.Y = i
			}
		}
	}
	return result
}

type TileRectilinearSize struct {
	Width  float64
	Height float64
}

func (pTmxMapIns *TmxMap) continuousObjLayerVecToContinuousMapNodeVec(continuousObjLayerVec *Vec2D) Vec2D {
	var tileRectilinearSize TileRectilinearSize
	tileRectilinearSize.Width = float64(pTmxMapIns.TileWidth)
	tileRectilinearSize.Height = float64(pTmxMapIns.TileHeight)
	tileSizeUnifiedLength := math.Sqrt(tileRectilinearSize.Width*tileRectilinearSize.Width*0.25 + tileRectilinearSize.Height*tileRectilinearSize.Height*0.25)
	isometricObjectLayerPointOffsetScaleFactor := (tileSizeUnifiedLength / tileRectilinearSize.Height)
	// fmt.Printf("tileWidth = %d,tileHeight = %d\n", pTmxMapIns.TileWidth, pTmxMapIns.TileHeight)
	cosineThetaRadian := (tileRectilinearSize.Width * 0.5) / tileSizeUnifiedLength
	sineThetaRadian := (tileRectilinearSize.Height * 0.5) / tileSizeUnifiedLength

	transMat := [...][2]float64{
		{isometricObjectLayerPointOffsetScaleFactor * cosineThetaRadian, -isometricObjectLayerPointOffsetScaleFactor * cosineThetaRadian},
		{-isometricObjectLayerPointOffsetScaleFactor * sineThetaRadian, -isometricObjectLayerPointOffsetScaleFactor * sineThetaRadian},
	}
	convertedVecX := transMat[0][0]*continuousObjLayerVec.X + transMat[0][1]*continuousObjLayerVec.Y
	convertedVecY := transMat[1][0]*continuousObjLayerVec.X + transMat[1][1]*continuousObjLayerVec.Y
	var converted Vec2D
	converted.X = convertedVecX + 0
	converted.Y = convertedVecY + 0.5*float64(pTmxMapIns.Height*pTmxMapIns.TileHeight)
	return converted
}

//通过离散的二维数组进行寻路, 返回一个Point数组
func FindPathByStartAndGoal(collideMap astar.Map, start astar.Point, goal astar.Point) []astar.Point {
	path := astar.AstarByStartAndGoalPoint(collideMap, start, goal)

	/*
	    * 打印地图
	    *
	   fmt.Printf("Path: %v \n",path);

	   //tmxMapIns.Path = path;
	   for _, pt := range path{
	     if(collideMap[pt.Y][pt.X] != 1){
	       collideMap[pt.Y][pt.X] = 9;
	     }
	   }
	   astar.PrintMap(collideMap);

	   //清空绿色点, 方便下次打印
	   for i:=0; i< len(collideMap); i++ {
	     for j:=0; j< len(collideMap[i]); j++ {
	       if(collideMap[i][j] == 9){
	         collideMap[i][j] = 0
	       }
	     }
	   }
	*/

	return path
}

/*
func ComputeColliderMapByCollision2d(pTmxMapIns *TmxMap) []int {
	var barrierLayerName string = "barrier"

	for _, objGroup := range pTmxMapIns.ObjectGroups {

		//fmt.Printf("objGroupName: %v \n", objGroup.Name)

		if barrierLayerName == objGroup.Name {
			//fmt.Printf("objGroup: %v \n", objGroup)
			barrierList := make([]collision2d.Polygon, len(objGroup.Objects))
			barrierCounter := 0
			for _, obj := range objGroup.Objects {

				initPos := Vec2D{
					X: obj.X,
					Y: obj.Y,
				}

				//Init Polygon body
				singleValueArray := strings.Split(obj.Polyline.Points, " ")
				pointsArrayWrtInit := make([]Vec2D, len(singleValueArray))


				for key, value := range singleValueArray {
					for k, v := range strings.Split(value, ",") {
						n, err := strconv.ParseFloat(v, 64)
						if err != nil {
							fmt.Printf("ERRRRRRRRRR!!!!!!!! parse float %f \n" + value)
							panic(err)
							//return err
						}
						if k%2 == 0 {
							pointsArrayWrtInit[key].X = n + initPos.X
						} else {
							pointsArrayWrtInit[key].Y = n + initPos.Y
						}
					}
				}

				//fmt.Printf("PointsArrayWrtInit: %v \n", pointsArrayWrtInit)

				pointsArrayTransted := make([]*Vec2D, len(pointsArrayWrtInit))
				pointList := make([]float64, len(singleValueArray) * 2)
				{
					var scale float64 = 1.0
					for key, value := range pointsArrayWrtInit {

						vec := pTmxMapIns.continuousObjLayerVecToContinuousMapNodeVec(&Vec2D{
							X: value.X * scale,
							Y: value.Y * scale,
						})
						pointsArrayTransted[key] = &vec
						pointList[2 * key] = vec.X
						pointList[2 * key + 1] = vec.Y

						//fmt.Printf("PointsArrayTransted: %v \n", vec)
					}
				}

				{
					//CreateBody
					pos := collision2d.NewVector(0.0, 0.0)
    				offset := collision2d.NewVector(0.0, 0.0)
    				angle := 0.0
					polygon := collision2d.NewPolygon(pos, offset, angle, pointList[:])

					barrierList[barrierCounter] = polygon
					barrierCounter++
					//log.Printf("barrier %v: %v \n", barrierCounter, polygon.Points)
				}
			}

			width := pTmxMapIns.Width
			height := pTmxMapIns.Height

			collideMap := make([]int, width*height)

			playerCircle := collision2d.Circle{collision2d.Vector{0, 0}, 12}

			for k, _ := range collideMap {
				x, y := pTmxMapIns.GetCoordByGid(k)

				playerCircle.Pos = collision2d.NewVector(x, y)

				for _, barrier := range barrierList {
					result, _ := collision2d.TestPolygonCircle(barrier, playerCircle)
					if result {
						collideMap[k] = 1
						break
					}
				}
			}

			log.Printf("collideMap %v ", collideMap)
			return collideMap
		}
	}

	return nil
}

*/

func ComputeColliderMapByCollision2dNeo(strToPolygon2DListMap map[string]*pb.Polygon2DList, pTmxMapIns *TmxMap) []int {
	barrierGroup := strToPolygon2DListMap["Barrier"]
	barrierList := make([]collision2d.Polygon, len(barrierGroup.Polygon2DList))
	barrierCounter := 0
	for _, polygon := range barrierGroup.Polygon2DList {
		pointList := make([]float64, len(polygon.Points)*2)
		for index, val := range polygon.Points {
			pointList[2*index] = val.X
			pointList[2*index+1] = val.Y
		}

		//CreateBody
		pos := collision2d.NewVector(0.0, 0.0)
		offset := collision2d.NewVector(0.0, 0.0)
		angle := 0.0
		polygon := collision2d.NewPolygon(pos, offset, angle, pointList[:])

		barrierList[barrierCounter] = polygon
		barrierCounter++
	}

	width := pTmxMapIns.Width
	height := pTmxMapIns.Height

	collideMap := make([]int, width*height)

	playerCircle := collision2d.Circle{collision2d.Vector{0, 0}, 12}

	for k, _ := range collideMap {
		x, y := pTmxMapIns.GetCoordByGid(k)

		playerCircle.Pos = collision2d.NewVector(x, y)

		for _, barrier := range barrierList {
			result, _ := collision2d.TestPolygonCircle(barrier, playerCircle)
			if result {
				collideMap[k] = 1
				break
			}
		}
	}

	log.Printf("collideMap %v ", collideMap)
	return collideMap
}

/*根据world里的collidableBody信息初始化一个离散的二维数组, 0为可通行区域, 1为障碍物
func InitCollideMap(world *box2d.B2World, pTmx *TmxMap) astar.Map {
	return astar.AstarArrayToMap(ComputeColliderMapByCollision2d(pTmx), pTmx.Width, pTmx.Height)
}

*/

func InitCollideMapNeo(pTmx *TmxMap, strToPolygon2DListMap map[string]*pb.Polygon2DList) astar.Map {
	return astar.AstarArrayToMap(ComputeColliderMapByCollision2dNeo(strToPolygon2DListMap, pTmx), pTmx.Width, pTmx.Height)
}
