package bigfloat

import (
	"math/big"
)

// Log sets o to z's natural logarithm to o's precision and returns o. Panics
// with ErrNaN if z is negative, including -0; returns -Inf when z = +0, and
// +Inf when z = +Inf. If o's precision is zero, then it is given the
// precision of z.
func Log(o, z *big.Float) *big.Float {
	if o.Prec() == 0 {
		o.SetPrec(z.Prec())
	}
	// panic on negative z
	if z.Signbit() {
		panic(ErrNaN{msg: "Log: argument is negative"})
	}
	// Log(0) = -Inf
	if z.Sign() == 0 {
		return o.SetInf(true)
	}
	// Log(+Inf) = +Inf
	if z.IsInf() {
		return o.Set(z)
	}

	prec := o.Prec() + 64 // guard digits

	one := big.NewFloat(1).SetPrec(prec)
	two := big.NewFloat(2).SetPrec(prec)
	four := big.NewFloat(4).SetPrec(prec)

	var neg bool
	switch z.Cmp(one) {
	case 1:
		o.SetPrec(prec).Set(z)
	case -1:
		// if 0 < z < 1 we compute log(z) as -log(1/z)
		o.SetPrec(prec).Quo(one, z)
		neg = true
	case 0:
		// Log(1) = 0
		return o.Set(&gzero)
	default:
		panic("bigfloat: unexpected comparison result, not 0, 1, or -1")
	}

	// We scale up x until x >= 2**(prec/2), and then we'll be allowed
	// to use the AGM formula for Log(x).
	//
	// Double x until the condition is met, and keep track of the
	// number of doubling we did (needed to scale back later).

	lim := new(big.Float)
	lim.SetMantExp(two, int(prec/2))

	k := 0
	for o.Cmp(lim) < 0 {
		o.Mul(o, o)
		k++
	}

	// Compute the natural log of z using the fact that
	//     log(z) = π / (2 * AGM(1, 4/z))
	// if
	//     z >= 2**(prec/2),
	// where prec is the desired precision (in bits)
	pi := pi(prec)
	agm := AGM(new(big.Float), one, o.Quo(four, o)) // agm = AGM(1, 4/z)
	o.Quo(pi, o.Mul(two, agm))

	if neg {
		o.Neg(o)
	}
	// scale the result back multiplying by 2**-k
	// reuse lim to reduce allocations.
	o.Mul(o, lim.SetMantExp(one, -k))

	return o.SetPrec(prec - 64)
}
