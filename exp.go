package bigfloat

import (
	"math"
	"math/big"
)

// Exp returns a big.Float representation of exp(z). Precision is
// the same as that of z. The function returns +Inf
// when z = +Inf, and 0 when z = -Inf.
func Exp(z *big.Float) *big.Float {
	// exp(0) == 1
	if z.Sign() == 0 {
		return big.NewFloat(1).SetPrec(z.Prec())
	}
	if z.IsInf() {
		r := new(big.Float).SetPrec(z.Prec())
		if z.Sign() > 0 {
			r.SetInf(false)
		}
		return r
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
		return new(big.Float).Mul(halfExp, halfExp).SetPrec(z.Prec())
	}
	// we got a nice IEEE-754 estimate
	guess.SetFloat64(zf)

	// f(t)/f'(t) = t*(log(t) - z)
	f := func(t *big.Float) *big.Float {
		x := new(big.Float)
		x.Sub(Log(t), z)
		return x.Mul(x, t)
	}

	x := newton(f, guess, z.Prec())

	return x
}
