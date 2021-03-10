package main

import (
	"image"
	"os"
	"strconv"
	"strings"
)

type config struct {
	width    int
	height   int
	xMax     float64
	xMin     float64
	yMax     float64
	yMin     float64
	nThreads int
}

func createConfig() *config {
	newConf := &config{
		width:    1500,
		height:   1000,
		xMax:     1,
		xMin:     -2,
		yMax:     1,
		yMin:     -1,
		nThreads: 1024,
	}

	args := os.Args[1:]
	for _, arg := range args {
		argArr := strings.Split(strings.Replace(arg, "--", "", 1), "=")
		switch argArr[0] {
		case "width":
			conf.width, _ = strconv.Atoi(argArr[1])
		case "height":
			conf.height, _ = strconv.Atoi(argArr[1])
		case "xMax":
			conf.xMax, _ = strconv.ParseFloat(argArr[1], 64)
		case "xMin":
			conf.xMin, _ = strconv.ParseFloat(argArr[1], 64)
		case "yMax":
			conf.yMax, _ = strconv.ParseFloat(argArr[1], 64)
		case "yMin":
			conf.yMin, _ = strconv.ParseFloat(argArr[1], 64)
		case "nThreads":
			conf.nThreads, _ = strconv.Atoi(argArr[1])
		default:
			panic("Unknown arguement " + arg)
		}
	}
	return newConf
}

func createImg() *safeImage {
	rect := image.Rect(0, 0, conf.width, conf.height)
	nImg := image.NewRGBA(rect)

	return &safeImage{img: nImg}
}
