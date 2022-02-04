package optimizations

import (
	"fmt"
	"math/big"
	"moritz/go-fractals/src/complexbig"
	"moritz/go-fractals/src/core"

	"github.com/schollz/progressbar/v3"
)

type ComplexInSet struct {
	z     *complexbig.ComplexBig
	inSet bool
}
type Grid struct {
	values                 [][]ComplexInSet
	xMin, xMax, yMin, yMax float64
	nLanes                 int
}

func NewGrid(nLanes, maxIt, maxThreads int) *Grid {

	values := make([][]ComplexInSet, nLanes)
	for i := range values {
		values[i] = make([]ComplexInSet, nLanes)
	}

	grid := &Grid{values: values,
		xMin: -2, xMax: 2,
		yMin: -1, yMax: 1,
		nLanes: nLanes}

	fillGrid(grid, maxIt, maxThreads)

	return grid
}

func PrintGridValue(i, j int, grid *Grid) {
	z := grid.values[i][j].z
	if grid.values[i][j].inSet {
		fmt.Printf(z.String() + " is in the set\n")
	} else {
		fmt.Printf(z.String() + " is not in the set\n")
	}
}

func fillGrid(grid *Grid, maxIt, maxThreads int) {
	bar := progressbar.Default(int64(grid.nLanes * grid.nLanes))
	guard := make(chan bool, maxThreads)
	for i := 0; i < grid.nLanes; i++ {
		for j := 0; j < grid.nLanes; j++ {
			guard <- true
			go func(i, j int) {
				z := getZ(i, j, grid)
				_, inSet := core.Iterate(z, maxIt)
				grid.values[i][j] = ComplexInSet{
					z: z, inSet: inSet,
				}
				bar.Add(1)
				<-guard
			}(i, j)

		}
	}
}

func IsAtBorder(z *complexbig.ComplexBig, grid *Grid) bool {
	minI := 0 // min i for which grid.values[i].R is larget than z.R
	minJ := 0 // min j for which grid.values[i].I is larget than z.I

	for i := 0; i < grid.nLanes; i++ {
		if grid.values[i][0].z.R.Cmp(z.R) == 1 {
			minI = i
			break
		}
	}

	for j := 0; j < grid.nLanes; j++ {
		if grid.values[0][j].z.I.Cmp(z.I) == 1 {
			minJ = j
			break
		}
	}
	a := grid.values[minI][minJ].inSet
	b := grid.values[minI][minJ-1].inSet
	c := grid.values[minI-1][minJ-1].inSet
	d := grid.values[minI-1][minJ].inSet

	// the point is at the border if some of the 4 points around it are in
	// the set and some are not
	return a != b || a != c || a != d

}

func getZ(i, j int, grid *Grid) *complexbig.ComplexBig {
	xDelta := (grid.xMax - grid.xMin)
	yDelta := (grid.yMax - grid.yMin)

	r := ((float64(i) / float64(grid.nLanes-1)) * xDelta) + grid.xMin

	im := ((float64(j) / float64(grid.nLanes-1)) * yDelta) + grid.yMin

	imBig := big.NewFloat(im)
	rBig := big.NewFloat(r)

	return &complexbig.ComplexBig{R: rBig, I: imBig}
}

func IsInMainCardiod(z *complexbig.ComplexBig) bool {
	zAbs := z.Abs()
	zAbsSquared := new(big.Float).Mul(zAbs, zAbs)

	// a = 8 * |c|^2
	a := new(big.Float).Mul(big.NewFloat(8), zAbsSquared)

	// b = a - 4
	// b = 8 * |c|^2 - 3
	b := new(big.Float).Sub(a, big.NewFloat(3))

	// leftSide = |c|^2 * b
	// leftSide = |c|^2 * (8 * |c|^2 - 3)
	leftSide := new(big.Float).Mul(zAbsSquared, b)

	// rightSide = 3/32 * Re(z)
	rightSide := new(big.Float).Sub(big.NewFloat(float64(3)/float64(32)), z.R)

	// leftSide =< rightSide
	return leftSide.Cmp(rightSide) <= 0

}
