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
	"moritz/go-fractals/src/utils"
	"os"
	"sync"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/gosuri/uilive"
	"github.com/schollz/progressbar/v3"
)

var zero *big.Float = big.NewFloat(0)
var one *big.Float = big.NewFloat(1)
var two *big.Float = big.NewFloat(2)

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
)

const width int = 7205 * 2
const height int = width / 2

var wg sync.WaitGroup
var mu sync.Mutex
var quitInitiated bool
var nFoundPoints *utils.SafeCounter = utils.MakeSafeCounter()
var nCyclesRun *utils.SafeCounter = utils.MakeSafeCounter()
var start time.Time

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
	cyclesWriter *uilive.Writer
	speedWriter  io.Writer
	totalWriter  io.Writer
	timeWriter   io.Writer
}

func init() {
	flag.IntVar(&prec, "prec", 100, "precision of big.float numbers")
	flag.IntVar(&maxIt, "maxIt", 100, "maximum number of iteratations")
	flag.IntVar(&cycleSize, "cycleSize", 100, "number of points per cycle")
	flag.IntVar(&nCycles, "nCycles", 100, "number of cycles")
	// flag.IntVar(&prec, "width", 500, "resolution width in pixels")
	flag.IntVar(&maxThreads, "maxThreads", 4, "maximum number of threads")
	flag.BoolVar(&endless, "endless", false, "endless mode, nCycles is ignored")

	flag.Parse()
}

func main() {
	fmt.Println("Creating image with resolution", width, "x", height)
	createArray()
	go renderPeriodically(10)

	start = time.Now()

	if endless {
		runEndless()
	} else {
		runNCycles()
	}

}

func createArray() {
	density = &SafeDensity{d: &[width][width * 2]uint16{}}
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

	speedWriter := cyclesWriter.Newline()
	totalWriter := cyclesWriter.Newline()
	timeWriter := cyclesWriter.Newline()

	writers := &writers{cyclesWriter: cyclesWriter, speedWriter: speedWriter, totalWriter: totalWriter, timeWriter: timeWriter}

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
	printStat(writers.speedWriter, "Avg. points / second", (pointsPerSecond))
	printStat(writers.totalWriter, "Total points", (nFoundPoints.Value()))

	fmt.Fprintf(writers.timeWriter, "Time elapsed %s \n", time.Since(start).String())
}
func printStat(writer io.Writer, label string, value int64) {
	fmt.Fprintf(writer, "%s %s \n", label, humanize.Comma(value))
}

func quitOnInput() {
	fmt.Println("Press enter to quit\n")
	// quit after user inputs any string
	reader := bufio.NewReader(os.Stdin)
	reader.ReadString('\n')
	fmt.Println("Quitting...")
	quitInitiated = true
}

func runCycle() {
	numbers := generateNumbers()
	trajectories := iteratePoints(numbers)
	pixels := translatePoints(trajectories)

	nFoundPoints.Add(int64(len(pixels)))
	nCyclesRun.Add(1)

	incrementDensity(pixels)
}

func generateNumbers() []*complexbig.ComplexBig {

	numbers := make([]*complexbig.ComplexBig, cycleSize)
	for j := 0; j < cycleSize; j++ {
		r := generateRandom(3, two)
		i := generateRandom(2, one)
		// throw away if |c| > 2
		numbers[j] = &complexbig.ComplexBig{R: r, I: i}
	}
	return numbers
}

func renderPeriodically(every int) {
	for {
		time.Sleep(time.Duration(every) * time.Second)
		render()
	}
}

func iteratePoints(numbers []*complexbig.ComplexBig) []*complexbig.ComplexBig {
	trajectories := make([]*complexbig.ComplexBig, 0, len(numbers))

	for j := 0; j < cycleSize; j++ {
		trajectoryPoints := iterate(numbers[j])

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

func iterate(c *complexbig.ComplexBig) []*complexbig.ComplexBig {
	z := &complexbig.ComplexBig{R: big.NewFloat(0), I: big.NewFloat(0)}
	previous := make([]*complexbig.ComplexBig, 0, maxIt)

	for i := 0; i < maxIt; i++ {
		// z = z*z + c
		z = complexbig.Mul(z, z)
		z.Add(c)

		// detect loop <=> series does not diverge
		for _, p := range previous {
			if p.Equals(z) {

				return nil
			}
		}

		// if |z| > 2 -> series diverges
		if z.Abs().Cmp(two) == 1 {
			return previous
		}
		previous = append(previous, z)
	}

	// series did not diverge after maxIt iterations
	return nil
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

func render() {

	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	max := uint16(0)
	localDensity := copyDensity()
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if localDensity[x][y] > max {
				max = localDensity[x][y]
			}
		}
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sat := uint8(float32(localDensity[x][y]) / float32(max) * float32(255))
			c := color.RGBA{sat, sat, sat, 255}
			img.Set(x, y, c)
		}
	}

	file, err := os.Create("buddhabrot.png")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	png.Encode(file, img)

	if quitInitiated {
		file.Close()
		os.Exit(0)
	}
}
