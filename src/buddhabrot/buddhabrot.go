package main

import (
	"bufio"
	"crypto/rand"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io"
	"math/big"
	"moritz/go-fractals/src/complexbig"
	"moritz/go-fractals/src/core"
	"moritz/go-fractals/src/optimizations"
	"moritz/go-fractals/src/utils"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uilive"
	"github.com/schollz/progressbar/v3"
)

const MAX_UINT16 = ^uint16(0)

var (
	prec       int
	maxIt      int
	cycleSize  int
	nCycles    int
	density    *SafeDensity
	xMax       float64 = 2
	xMin       float64 = -2
	yMax       float64 = 1
	yMin       float64 = -1
	maxThreads int
	endless    bool
	warmStart  bool
)

// const width int = 7205 * 2
const width int = 500 * 2
const height int = width / 2

var wg sync.WaitGroup
var mu sync.Mutex
var quitInitiated bool
var nFoundPoints *utils.SafeCounter = utils.MakeSafeCounter()
var nOldPoints int64 = 0
var nCyclesRun *utils.SafeCounter = utils.MakeSafeCounter()
var lastMax uint16 = 0
var start time.Time

var grid *optimizations.Grid

var (
	xDelta float64 = xMax - xMin
	yDelta float64 = yMax - yMin
)

type pixel struct {
	x uint16
	y uint16
}

type SafeDensity struct {
	sync.Mutex
	d *[width][width * 2]uint16
}

type writers struct {
	cyclesWriter   *uilive.Writer
	speedWriter    io.Writer
	totalWriter    io.Writer
	totalNewWriter io.Writer
	maxWriter      io.Writer
	timeWriter     io.Writer
}

func init() {
	flag.IntVar(&prec, "prec", 100, "precision of big.float numbers")
	flag.IntVar(&maxIt, "maxIt", 100, "maximum number of iteratations")
	flag.IntVar(&cycleSize, "cycleSize", 100, "number of points per cycle")
	flag.IntVar(&nCycles, "nCycles", 100, "number of cycles")
	flag.IntVar(&maxThreads, "maxThreads", 4, "maximum number of threads")
	flag.BoolVar(&endless, "endless", false, "endless mode, nCycles is ignored")
	flag.BoolVar(&warmStart, "warmStart", false, "warm start, load density and max from files")

	flag.Parse()
}

func main() {

	fmt.Println("Creating image with resolution", width, "x", height)
	initDensityArray()

	start = time.Now()
	grid = optimizations.NewGrid(500, maxIt, maxThreads)
	fmt.Printf("Grid created in %s\n", time.Since(start))

	go renderPeriodically(2)

	start = time.Now()

	if endless {
		runEndless()
	} else {
		runNCycles()
	}

}

func initDensityArray() {
	if !warmStart {
		density = &SafeDensity{d: &[width][width * 2]uint16{}}
		return
	}

	loadedDensity, err := loadDensity("buddhabrot.png", "max.txt")
	if err != nil {
		density = &SafeDensity{d: &[width][width * 2]uint16{}}
		return
	}

	density = &SafeDensity{d: loadedDensity}

	sumPoints := int64(0)
	max := uint16(0)
	for i := 0; i < width; i++ {
		for j := 0; j < width*2; j++ {
			sumPoints += int64(density.d[i][j])
			if density.d[i][j] > max {
				max = density.d[i][j]
			}
		}
	}
	nOldPoints = int64(sumPoints)
	lastMax = max
	fmt.Println("Loaded", humanize.Comma(int64(sumPoints)), "points")
	fmt.Println("Max:", max)

}

func runNCycles() {
	guard := make(chan struct{}, maxThreads)
	bar := progressbar.Default(int64(nCycles))
	for i := 0; i < nCycles; i++ {
		guard <- struct{}{}
		go func() {
			wg.Add(1)
			runCycle()
			bar.Add(1)
			<-guard
			wg.Done()
		}()
	}
	wg.Wait()
	os.Exit(0)
}

func runEndless() {
	guard := make(chan struct{}, maxThreads)

	go quitOnInput()

	cyclesWriter := uilive.New() // writer for the first line
	cyclesWriter.Start()

	totalNewWriter := cyclesWriter.Newline()
	totalWriter := cyclesWriter.Newline()
	speedWriter := cyclesWriter.Newline()
	maxWriter := cyclesWriter.Newline()
	timeWriter := cyclesWriter.Newline()

	writers := &writers{
		cyclesWriter:   cyclesWriter,
		speedWriter:    speedWriter,
		totalWriter:    totalWriter,
		timeWriter:     timeWriter,
		maxWriter:      maxWriter,
		totalNewWriter: totalNewWriter}

	go printStatsPeriodically(1, writers)

	for {
		guard <- struct{}{}
		go func() {
			runCycle()
			<-guard
		}()
	}
}

func printStatsPeriodically(every int, writers *writers) {
	for {
		time.Sleep(time.Duration(every) * time.Second)
		if quitInitiated {
			return
		}
		printStats(writers)
	}
}

func printStats(writers *writers) {
	secondsSinceStart := int64(time.Since(start).Seconds())
	if secondsSinceStart == 0 {
		secondsSinceStart = 1
	}
	pointsPerSecond := nFoundPoints.Value() / secondsSinceStart
	printStat(writers.cyclesWriter, "Cycles started", nCyclesRun.Value())
	printStat(writers.totalWriter, "Total points", (nFoundPoints.Value() + nOldPoints))
	printStat(writers.totalNewWriter, "New points", (nFoundPoints.Value()))
	printStat(writers.maxWriter, "Maximum number of trajectory hits", int64(lastMax))
	printStat(writers.speedWriter, "Avg. points / second", (pointsPerSecond))

	fmt.Fprintf(writers.timeWriter, "Time elapsed %s \n", time.Since(start).String())
}
func printStat(writer io.Writer, label string, value int64) {
	fmt.Fprintf(writer, "%s %s \n", label, humanize.Comma(value))
}

func quitOnInput() {
	fmt.Println("Press enter to quit")
	fmt.Println("")
	// quit after user inputs any string
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	fmt.Println("Quitting...")
	quitInitiated = true
}

func renderPeriodically(every int) {
	for {
		time.Sleep(time.Duration(every) * time.Second)
		render(copyDensity())

		if quitInitiated {
			os.Exit(0)
		}
	}
}

func runCycle() {
	numbers := generateNumbers()
	numbers = filterNumbers(numbers)
	trajectories := iteratePoints(numbers)
	pixels := translatePoints(trajectories)

	nFoundPoints.Add(int64(len(pixels)))
	nCyclesRun.Add(1)

	incrementDensity(pixels)
}

func generateNumbers() []*complexbig.ComplexBig {

	numbers := make([]*complexbig.ComplexBig, cycleSize)
	for j := 0; j < cycleSize; j++ {
		r := generateRandom(3, complexbig.Two)
		i := generateRandom(2, complexbig.One)
		numbers[j] = &complexbig.ComplexBig{R: r, I: i}
	}
	return numbers
}

func filterNumbers(numbers []*complexbig.ComplexBig) []*complexbig.ComplexBig {
	filtered := make([]*complexbig.ComplexBig, 0, cycleSize)
	for _, z := range numbers {
		if !optimizations.IsAtBorder(z, grid) {
			continue
		}

		if optimizations.IsInMainCardiod(z) {
			continue
		}

		filtered = append(filtered, z)
	}
	return filtered
}

func iteratePoints(numbers []*complexbig.ComplexBig) []*complexbig.ComplexBig {
	trajectories := make([]*complexbig.ComplexBig, 0, len(numbers))

	for j := 0; j < len(numbers); j++ {
		trajectoryPoints, _ := core.Iterate(numbers[j], maxIt)

		if trajectoryPoints == nil {
			continue
		}
		trajectories = append(trajectories, trajectoryPoints...)
	}
	return trajectories
}

func generateRandom(size int, offset *big.Float) *big.Float {

	max := new(big.Int)
	max.Exp(big.NewInt(2), big.NewInt(int64(prec)), nil).Sub(max, big.NewInt(1))

	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		//error handling
	}

	r := new(big.Float).SetInt(n)

	r.SetMantExp(r, r.MantExp(r)-(prec-size+1))
	r.Sub(r, offset)
	return r
}

func translatePoints(points []*complexbig.ComplexBig) []*pixel {
	pixels := make([]*pixel, 0, len(points))
	for _, c := range points {
		pixel := translatePoint(c)
		if pixel == nil {
			continue
		}
		pixels = append(pixels, pixel)
	}
	return pixels
}

func translatePoint(point *complexbig.ComplexBig) *pixel {
	r, _ := point.R.Float64()
	if r > xMax || r < xMin {
		return nil
	}
	i, _ := point.I.Float64()
	if i > yMax || i < yMin {
		return nil
	}

	return &pixel{
		x: uint16(((r - xMin) / xDelta) * float64(width)),
		y: uint16(((i - yMin) / yDelta) * float64(height))}
}

func incrementDensity(pixels []*pixel) {
	mu.Lock()
	for _, pixel := range pixels {
		// if density.d[pixel.x][pixel.y] == MAX_UINT16 {
		// 	panic("overflow")
		// }
		density.d[pixel.x][pixel.y]++
	}
	mu.Unlock()
}

func copyDensity() *[width][width * 2]uint16 {
	mu.Lock()
	defer mu.Unlock()
	d := &[width][width * 2]uint16{}
	for i := 0; i < width; i++ {
		for j := 0; j < width*2; j++ {
			d[i][j] = density.d[i][j]
		}
	}
	return d
}

func render(density *[width][width * 2]uint16) {

	max := findMax(density)
	lastMax = max

	saveMax(max)

	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)

	drawImage(img, density, max)

	saveImage(img)
}

func drawImage(img *image.RGBA, density *[width][width * 2]uint16, max uint16) {
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sat := uint8(float32(density[x][y]) / float32(max) * float32(255))
			c := color.RGBA{sat, sat, sat, 255}
			img.Set(x, y, c)
		}
	}
}

func findMax(density *[width][width * 2]uint16) uint16 {
	max := uint16(0)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if density[x][y] > max {
				max = density[x][y]
			}
		}
	}
	return max
}

func saveMax(max uint16) {
	f, err := os.Create("max.txt")
	if err != nil {
		panic(err)
	}
	defer f.Close()
	f.WriteString(strconv.Itoa(int(max)))
}

func saveImage(img *image.RGBA) {
	file, err := os.Create("buddhabrot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)
}

func loadDensity(imagePath string, maxPath string) (*[width][width * 2]uint16, error) {
	imgFile, err := os.Open(imagePath)
	if err != nil {
		return nil, err
	}
	defer imgFile.Close()

	img, err := png.Decode(imgFile)

	if err != nil {
		return nil, err
	}

	maxFile, err := os.Open(maxPath)
	if err != nil {
		return nil, err
	}
	defer maxFile.Close()

	// read single number from file
	scanner := bufio.NewScanner(maxFile)
	scanner.Scan()
	max, err := strconv.Atoi(scanner.Text())
	if err != nil {
		return nil, err
	}

	density := &[width][width * 2]uint16{}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			c := img.At(x, y)
			r, _, _, _ := c.RGBA()
			r = r / 256
			density[x][y] = uint16(int(r) * max / 255)
		}
	}
	return density, nil
}
