package bigfloat

import (
	"math/big"
)

// Log sets z to its natural logarithm. Panics if z is negative, including -0;
// returns -Inf when z = +0, and +Inf when z = +Inf
func Log(z *big.Float) *big.Float {
	// panic on negative z
	if z.Signbit() {
		panic(ErrNaN{msg: "Log: argument is negative"})
	}
	// Log(0) = -Inf
	if z.Sign() == 0 {
		return z.SetInf(true)
	}
	// Log(+Inf) = +Inf
	if z.IsInf() {
		return z
	}

	prec := z.Prec() + 64 // guard digits

	one := big.NewFloat(1).SetPrec(prec)
	two := big.NewFloat(2).SetPrec(prec)
	four := big.NewFloat(4).SetPrec(prec)

	// Log(1) = 0
	if z.Cmp(one) == 0 {
		return z.Set(&gzero)
	}

	z.SetPrec(prec)

	// if 0 < z < 1 we compute log(z) as -log(1/z)
	var neg bool
	if z.Cmp(one) < 0 {
		z.Quo(one, z)
		neg = true
	}

	// We scale up x until x >= 2**(prec/2), and then we'll be allowed
	// to use the AGM formula for Log(x).
	//
	// Double x until the condition is met, and keep track of the
	// number of doubling we did (needed to scale back later).

	lim := new(big.Float)
	lim.SetMantExp(two, int(prec/2))

	k := 0
	for z.Cmp(lim) < 0 {
		z.Mul(z, z)
		k++
	}

	// Compute the natural log of z using the fact that
	//     log(z) = π / (2 * AGM(1, 4/z))
	// if
	//     z >= 2**(prec/2),
	// where prec is the desired precision (in bits)
	pi := pi(prec)
	agm := agm(one, z.Quo(four, z)) // agm = AGM(1, 4/z)
	z.Quo(pi, z.Mul(two, agm))

	if neg {
		z.Neg(z)
	}
	// scale the result back multiplying by 2**-k
	// reuse lim to reduce allocations.
	z.Mul(z, lim.SetMantExp(one, -k))

	return z.SetPrec(prec - 64)
}
