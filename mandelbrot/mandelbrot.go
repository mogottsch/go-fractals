package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"strconv"
)

var img image.Image

func main() {
	initImage()
	saveImage()
}

func initImage() {
	args := os.Args[1:]
	fmt.Println(args)

	width, _ := strconv.Atoi(args[0])
	height, _ := strconv.Atoi(args[1])

	rect := image.Rect(0, 0, width, height)
	img = image.NewRGBA(rect)
}

func saveImage() {
	file, err := os.Create("mandelbrot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}