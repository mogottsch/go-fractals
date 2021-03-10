package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strconv"
	"strings"
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

type config struct {
	width    int
	height   int
	xMax     float64
	xMin     float64
	yMax     float64
	yMin     float64
	nThreads int
}

func (img *safeImage) setPixel(x, y int, c color.Color) {
	// img.mu.Lock()
	// defer img.mu.Unlock()
	img.img.Set(x, y, c)
}

func main() {
	configure()
	initImg()
	measureTime(drawPartially)
	save()
}

func configure() {
	conf = &config{
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
}

func initImg() {
	rect := image.Rect(0, 0, conf.width, conf.height)
	nImg := image.NewRGBA(rect)

	img = &safeImage{img: nImg}
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
