package complexbig

import "math/big"

// ComplexBig are complex numbers represented by big float
type ComplexBig struct {
	R *big.Float
	I *big.Float
}

// Mul a*b=z
func Mul(a, b *ComplexBig) (z *ComplexBig) {
	// newR = a.R * b.R - a.I * b.I
	newR := new(big.Float).Sub(
		new(big.Float).Mul(a.R, b.R),
		new(big.Float).Mul(a.I, b.I))
	newI := new(big.Float).Add(
		new(big.Float).Mul(a.I, b.R),
		new(big.Float).Mul(a.R, b.I))
	return &ComplexBig{R: newR, I: newI}
}

// Add a+b=z
func (a *ComplexBig) Add(b *ComplexBig) (z *ComplexBig) {
	a.R.Add(a.R, b.R)
	a.I.Add(a.I, b.I)
	return a
}

// Equals checks two complex bigs for equality
func (a *ComplexBig) Equals(b *ComplexBig) (isEqual bool) {
	return a.R.Cmp(b.R) == 0 && a.I.Cmp(b.I) == 0
}

// Abs gets the absolute value of a
func (a *ComplexBig) Abs() (z *big.Float) {
	r := new(big.Float).Mul(a.R, a.R)
	i := new(big.Float).Mul(a.I, a.I)

	r.Add(r, i)

	return r.Sqrt(r)
}
