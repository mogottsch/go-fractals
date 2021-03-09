package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"strconv"
	"sync"
	"time"
)

var img *safeImage
var width int
var height int

var xMax float64 = 1
var xMin float64 = -2
var yMax float64 = 1
var yMin float64 = -1

var wg sync.WaitGroup

type safeImage struct {
	img *image.RGBA
	mu  sync.Mutex
}

func (img *safeImage) setPixel(x, y int, c color.Color) {
	// img.mu.Lock()
	// defer img.mu.Unlock()
	img.img.Set(x, y, c)
}

func main() {
	initImg()
	measureTime(drawPartially)
	save()
}

func initImg() {
	args := os.Args[1:]
	width, _ = strconv.Atoi(args[0])
	height = width * 2 / 3

	rect := image.Rect(0, 0, width, height)
	nImg := image.NewRGBA(rect)

	img = &safeImage{img: nImg}
}

func setPixels(fn func(x, y int) color.Color) {
	for y := 0; y < height; y++ {
		wg.Add(1)
		go func(y int) {
			for x := 0; x < width; x++ {
				img.setPixel(x, y, fn(x, y))
			}
			wg.Done()
		}(y)
	}
}

func setPixelsPartially(yL, yH, xL, xH int) {
	for y := yL; y < yH; y++ {
		for x := xL; x < xH; x++ {
			img.setPixel(x, y, getPixelColor(x, y))
		}
	}
}

func getPixelColor(x, y int) color.Color {
	if diverges(translate(x, y)) {
		return color.RGBA{255, 255, 255, 255}
	}
	return color.RGBA{0, 0, 0, 255}
}

func drawPartially() {
	n := 8
	c := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			c++
			wg.Add(1)
			go func(i, j int) {
				setPixelsPartially(height/n*(i), height/n*(i+1), width/n*(j), width/n*(j+1))
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()
	fmt.Printf("n threads: %v \n", c)
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
		float64(x)/float64(width)*(xMax-xMin)+xMin,
		(float64(y)/float64(height)*(yMax-yMin)+yMin)*-1,
	)
}

func measureTime(fn func()) {
	start := time.Now()
	fn()
	elapsed := time.Since(start)
	fmt.Printf("%s\n", elapsed)
}

func save() {
	file, err := os.Create("mandelbrot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img.img)
}
