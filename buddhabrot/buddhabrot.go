package main

import (
	"crypto/rand"
	"flag"
	"fractals/complexbig"
	"image"
	"image/color"
	"image/png"
	"math/big"
	"os"
)

var zero *big.Float = big.NewFloat(0)
var one *big.Float = big.NewFloat(1)
var two *big.Float = big.NewFloat(2)

var (
	prec      int
	maxIt     int
	cycleSize int
	density   [width][width * 2]int
	xMax      float64 = 2
	xMin      float64 = -2
	yMax      float64 = 1
	yMin      float64 = -1
)

const width int = 500
const height int = 250

var (
	xDelta float64 = xMax - xMin
	yDelta float64 = yMax - yMin
)

type pixel struct {
	x int
	y int
}

func main() {
	createArray()

	numbers := make([]*complexbig.ComplexBig, cycleSize)
	for j := 0; j < cycleSize; j++ {
		r := generateRandom(3, two)
		i := generateRandom(2, one)
		// throw away if |c| > 2
		numbers[j] = &complexbig.ComplexBig{R: r, I: i}
	}
	for j := 0; j < cycleSize; j++ {
		points := iterate(numbers[j])

		if points == nil {
			continue
		}
		pixels := translatePoints(points)
		incrementDensity(pixels)
	}

	render()
}

func createArray() {
	density = [width][width * 2]int{}
}

func init() {
	flag.IntVar(&prec, "prec", 100, "precision of big.float numbers")
	flag.IntVar(&maxIt, "maxIt", 100, "maximum number of iteratations")
	flag.IntVar(&cycleSize, "cycleSize", 100, "number of points per cycle")
	// flag.IntVar(&prec, "width", 500, "resolution width in pixels")

	flag.Parse()
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
		// panic(fmt.Sprintf("unexpected value for r: %v", r))
	}
	i, _ := point.I.Float64()
	if i > yMax || i < yMin {
		return nil
		// panic(fmt.Sprintf("unexpected value for i: %v", i))
	}

	return &pixel{
		x: int(((r - xMin) / xDelta) * float64(width)),
		y: int(((i - yMin) / yDelta) * float64(height))}
}

func incrementDensity(pixels []*pixel) {
	for _, c := range pixels {
		density[c.x][c.y]++
	}
}

func render() {
	rect := image.Rect(0, 0, width, height)
	img := image.NewRGBA(rect)
	max := 0
	// fmt.Println(density)
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			if density[x][y] > max {
				max = density[x][y]
			}
		}
	}
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			sat := uint8(float32(density[x][y]) / float32(max) * float32(255))
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
}
