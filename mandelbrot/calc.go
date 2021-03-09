package calc

import "math/cmplx"

var maxIt int = 100

func Diverges(c complex128) bool {
	z := complex128(0)

	for i := 0; i < maxIt; i++ {
		z = cmplx.Sqrt(z) + c
		if cmplx.Abs(z) > 2 {
			return true
		}
	}
	return false
}
