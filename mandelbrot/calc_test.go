package main

import (
	"math/big"
	"testing"
)

func TestMul(t *testing.T) {
	a := &complexBig{big.NewFloat(2), big.NewFloat(4)}
	b := &complexBig{big.NewFloat(3), big.NewFloat(5)}

	z := mul(a, b)

	if z.r.Cmp(big.NewFloat(-14)) != 0 {
		t.Fatalf("expected -14, got %v", z.r)
	}
	if z.i.Cmp(big.NewFloat(22)) != 0 {
		t.Fatalf("expected 22, got %v", z.i)
	}

	a = &complexBig{big.NewFloat(6), big.NewFloat(3)}
	b = &complexBig{big.NewFloat(7), big.NewFloat(-1)}

	z = mul(a, b)

	if z.r.Cmp(big.NewFloat(45)) != 0 {
		t.Fatalf("expected 45, got %v", z.r)
	}
	if z.i.Cmp(big.NewFloat(15)) != 0 {
		t.Fatalf("expected 15, got %v", z.i)
	}
}

func TestAdd(t *testing.T) {
	a := &complexBig{big.NewFloat(2), big.NewFloat(52)}
	b := &complexBig{big.NewFloat(-5), big.NewFloat(-2)}

	a.add(b)

	if a.r.Cmp(big.NewFloat(-3)) != 0 {
		t.Fatalf("expected -3, got %v", a.r)
	}
	if a.i.Cmp(big.NewFloat(50)) != 0 {
		t.Fatalf("expected 50, got %v", a.i)
	}
}

func TestAbs(t *testing.T) {
	a := &complexBig{big.NewFloat(5), big.NewFloat(12)}

	if a.abs().Cmp(big.NewFloat(13)) != 0 {
		t.Fatalf("expected 13, got %v", a.i)
	}
	a = &complexBig{big.NewFloat(3), big.NewFloat(-2)}

	sqrt13 := new(big.Float).Sqrt(big.NewFloat(13))
	if a.abs().Cmp(sqrt13) != 0 {
		t.Fatalf("expected sqrt(13), got %v", a.i)
	}
}
