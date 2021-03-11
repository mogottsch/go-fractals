package main

import (
	"image"
	"math/big"
	"os"
	"strconv"
	"strings"
)

type config struct {
	width    int
	height   int
	xMax     *big.Float
	xMin     *big.Float
	xDelta   *big.Float
	yMax     *big.Float
	yMin     *big.Float
	yDelta   *big.Float
	maxIt    int
	skip     bool
	prec     int
	nThreads int
}

func createConfig() *config {
	newConf := &config{
		width:    1500,
		height:   1000,
		xMax:     big.NewFloat(1.0),
		xMin:     big.NewFloat(-2.0),
		yMax:     big.NewFloat(1.0),
		yMin:     big.NewFloat(-1.0),
		maxIt:    100,
		skip:     false,
		prec:     53,
		nThreads: 1024,
	}

	args := os.Args[1:]
	for _, arg := range args {
		argArr := strings.Split(strings.Replace(arg, "--", "", 1), "=")
		switch argArr[0] {
		case "width":
			newConf.width, _ = strconv.Atoi(argArr[1])
		case "height":
			newConf.height, _ = strconv.Atoi(argArr[1])
		case "xMax":
			float, _ := strconv.ParseFloat(argArr[1], 64)
			newConf.xMax = big.NewFloat(float)
		case "xMin":
			float, _ := strconv.ParseFloat(argArr[1], 64)
			newConf.xMin = big.NewFloat(float)
		case "yMax":
			float, _ := strconv.ParseFloat(argArr[1], 64)
			newConf.yMax = big.NewFloat(float)
		case "yMin":
			float, _ := strconv.ParseFloat(argArr[1], 64)
			newConf.yMin = big.NewFloat(float)
		case "nThreads":
			newConf.nThreads, _ = strconv.Atoi(argArr[1])
		case "maxIt":
			newConf.maxIt, _ = strconv.Atoi(argArr[1])
		case "skip":
			newConf.skip = true
		default:
			panic("Unknown arguement " + arg)
		}
	}

	newConf.xDelta = new(big.Float).Sub(newConf.xMax, newConf.xMin)
	newConf.yDelta = new(big.Float).Sub(newConf.yMax, newConf.yMin)
	return newConf
}

func createImg() *safeImage {
	rect := image.Rect(0, 0, conf.width, conf.height)
	nImg := image.NewRGBA(rect)

	return &safeImage{img: nImg}
}
