package main

import (
	"math/big"
)

var two *big.Float = big.NewFloat(2)

type complexBig struct {
	r *big.Float
	i *big.Float
}

func mul(a, b *complexBig) (z *complexBig) {
	// newR = a.r * b.r - a.i * b.i
	newR := new(big.Float).Sub(
		new(big.Float).Mul(a.r, b.r),
		new(big.Float).Mul(a.i, b.i))
	newI := new(big.Float).Add(
		new(big.Float).Mul(a.i, b.r),
		new(big.Float).Mul(a.r, b.i))
	return &complexBig{r: newR, i: newI}
}

func (a *complexBig) add(b *complexBig) (z *complexBig) {
	a.r.Add(a.r, b.r)
	a.i.Add(a.i, b.i)
	return a
}

func (a *complexBig) abs() (z *big.Float) {
	r := new(big.Float).Mul(a.r, a.r)
	i := new(big.Float).Mul(a.i, a.i)

	r.Add(r, i)

	return r.Sqrt(r)
}

func diverges(c *complexBig) bool {
	z := &complexBig{big.NewFloat(0), big.NewFloat(0)}

	for i := 0; i < conf.maxIt; i++ {
		// z = z*z + c
		z = mul(z, z)
		z.add(c)

		// if |z| > 2
		if z.abs().Cmp(two) == 1 {
			return true
		}
	}
	return false
}
