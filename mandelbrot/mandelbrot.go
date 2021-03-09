package main

import (
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"

	"fractals/progressbar"
)

var img *image.RGBA
var width int
var height int

var xMax int = 1
var xMin int = -2
var yMax int = 1
var yMin int = -1

func main() {
	initImg()
	draw()
	save()
}

func initImg() {
	args := os.Args[1:]
	width, _ = strconv.Atoi(args[0])
	height = width * 2 / 3

	rect := image.Rect(0, 0, width, height)
	img = image.NewRGBA(rect)
}

func setPixels(fn func(x, y int) color.Color) {
	bar := progressbar.New(height)
	for y := 0; y < height; y++ {
		// fmt.Printf("%.f/%v\r", float64(y)/float64(height)*100, 100)
		bar.Update(y)
		for x := 0; x < width; x++ {
			img.Set(x, y, fn(x, y))
		}
	}
}

func draw() {
	setPixels(func(x, y int) color.Color {
		if diverges(translate(x, y)) {
			return color.RGBA{255, 255, 255, 255}
		}
		return color.RGBA{0, 0, 0, 255}
	})
}

func translate(x, y int) complex128 {
	return complex(
		float64(x)/float64(width)*float64((xMax-xMin))+float64(xMin),
		(float64(y)/float64(height)*float64(yMax-yMin)+float64(yMin))*-1,
	)
}

func save() {
	file, err := os.Create("mandelbrot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}
