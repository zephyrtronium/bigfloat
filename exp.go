package bigfloat

import (
	"math"
	"math/big"
)

// Exp sets z to exp(z) to its precision and returns z. The result is zero if z
// is -inf and +inf if z is +inf. If the precision of z is zero, then the
// result will be 1 with precision 53.
func Exp(z *big.Float) *big.Float {
	if z.Sign() == 0 {
		return z.SetFloat64(1)
	}
	if z.IsInf() {
		if z.Sign() < 0 {
			z.Set(&gzero)
		}
		return z
	}

	guess := new(big.Float)
	// try to get initial estimate using IEEE-754 math
	// TODO: save work (and an import of math) by checking the exponent of z
	zf, _ := z.Float64()
	zf = math.Exp(zf)
	if math.IsInf(zf, 1) || zf == 0 {
		// too big or too small for IEEE-754 math,
		// perform argument reduction using
		//     e^{2z} = (e^z)²
		halfZ := new(big.Float).Mul(z, big.NewFloat(0.5))
		halfExp := Exp(halfZ.SetPrec(z.Prec() + 64))
		// TODO: avoid recursion
		return z.Mul(halfExp, halfExp)
	}
	// we got a nice IEEE-754 estimate
	guess.SetFloat64(zf)

	// f(t)/f'(t) = t*(log(t) - z)
	f := func(t *big.Float) *big.Float {
		x := new(big.Float)
		x.Sub(Log(new(big.Float).Copy(t)), z)
		return x.Mul(x, t)
	}

	x := newton(f, guess, z.Prec()) // TODO: make newton operate in place

	return z.Set(x)
}
