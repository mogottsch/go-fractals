package main

import (
	"fmt"
	"fractals/mandelbrot/calc"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
)

var img *image.RGBA
var width int
var height int

func main() {
	initImg()
	save()
	calc.Diverges()
	fmt.Println(translate(50, 50))
}

func initImg() {
	args := os.Args[1:]
	fmt.Println(args)

	width, _ = strconv.Atoi(args[0])
	height, _ = strconv.Atoi(args[1])

	rect := image.Rect(0, 0, width, height)
	img = image.NewRGBA(rect)

	setPixels(func(x, y int) color.Color {
		return color.RGBA{0, 0, 0, 255}
	})
}

func setPixels(fn func(x, y int) color.Color) {
	for y := 0; y < height; y++ {
		for x := 0; x < height; x++ {
			img.Set(x, y, fn(x, y))
		}
	}
}

func translate(x, y int) complex128 {
	return complex(
		float64(x)/float64(width)*2-1,
		(float64(y)/float64(height)*2-1)*-1,
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
