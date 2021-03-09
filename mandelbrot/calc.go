package main

import "math/cmplx"

var maxIt int = 500

func diverges(c complex128) bool {
	z := complex128(0)

	for i := 0; i < maxIt; i++ {
		z = z*z + c
		if cmplx.Abs(z) > 2 {
			return true
		}
	}
	return false
}
