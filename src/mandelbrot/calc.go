package main

import (
	"math/big"
)

var two *big.Float = big.NewFloat(2)
var skipped int

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

func (a *complexBig) equals(b *complexBig) (isEqual bool) {
	return a.r.Cmp(b.r) == 0 && a.i.Cmp(b.i) == 0
}

func (a *complexBig) abs() (z *big.Float) {
	r := new(big.Float).Mul(a.r, a.r)
	i := new(big.Float).Mul(a.i, a.i)

	r.Add(r, i)

	return r.Sqrt(r)
}

func diverges(c *complexBig) (bool, int) {
	z := &complexBig{big.NewFloat(0), big.NewFloat(0)}
	previous := make([]*complexBig, 0, conf.maxIt)

	for i := 0; i < conf.maxIt; i++ {
		// detect loop <=> does not diverge
		if conf.skip {
			for _, p := range previous {
				if p.equals(z) {
					skipped++
					return false, i
				}
			}
		}
		if conf.skip {
			previous = append(previous, z)
		}
		// z = z*z + c
		z = mul(z, z)
		z.add(c)

		// if |z| > 2
		if z.abs().Cmp(two) == 1 {
			return true, i
		}
	}
	return false, conf.maxIt
}
