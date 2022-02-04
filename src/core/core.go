package core

import (
	"math/big"
	"moritz/go-fractals/src/complexbig"
)

func Iterate(c *complexbig.ComplexBig, maxIt int) ([]*complexbig.ComplexBig, bool) {
	z := &complexbig.ComplexBig{R: big.NewFloat(0), I: big.NewFloat(0)}
	oldZ := &complexbig.ComplexBig{R: big.NewFloat(0), I: big.NewFloat(0)}

	previous := make([]*complexbig.ComplexBig, 0, maxIt)

	stepsTaken := 0
	stepLimit := 2

	for i := 0; i < maxIt; i++ {
		// z = z*z + c
		z = complexbig.Mul(z, z)
		z.Add(c)

		// brents cycle detection
		if z.Equals(oldZ) {
			return nil, true
		}

		if stepsTaken == stepLimit {
			oldZ = z
			stepsTaken = 0
			stepLimit *= 2
		}

		stepsTaken++

		// if |z| > 2 -> series diverges
		if z.Abs().Cmp(complexbig.Two) == 1 {
			return previous, false
		}
		previous = append(previous, z)
	}

	// series did not diverge after maxIt iterations
	return nil, true
}
