package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"sync"
	"time"
)

var img *safeImage
var conf *config

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
	conf = createConfig()
	img = createImg()
	measureTime(drawPartially)
	save()
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
	// we split the coordinate system into nThreads areas of equal width
	n := int(math.Sqrt(float64(conf.nThreads)))
	c := 0
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			c++
			wg.Add(1)
			go func(i, j int) {
				// we have to check if the area that will be calculated
				// is at the right border or lower border of the image.
				// if so, we will use conf.height and conf.width respectively,
				// so that there are no empty borders.
				yH := conf.height / n * (i + 1)
				if i == n-1 {
					yH = conf.height
				}
				xH := conf.width / n * (j + 1)
				if j == n-1 {
					xH = conf.width
				}
				setPixelsPartially(
					conf.height/n*(i),
					yH,
					conf.width/n*(j),
					xH)
				wg.Done()
			}(i, j)
		}
	}
	wg.Wait()
	fmt.Printf("n threads: %v \n", c)
}

func translate(x, y int) complex128 {
	return complex(
		float64(x)/float64(conf.width)*(conf.xMax-conf.xMin)+conf.xMin,
		(float64(y)/float64(conf.height)*(conf.yMax-conf.yMin)+conf.yMin)*-1,
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
