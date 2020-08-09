package bigfloat

import (
	"math"
	"math/big"
)

// Exp sets o to exp(z) to o's precision and returns o. The result is zero if z
// is -inf and +inf if z is +inf. If o's precision is zero, then it is given
// the precision of z.
func Exp(o, z *big.Float) *big.Float {
	if o.Prec() == 0 {
		o.SetPrec(z.Prec())
	}
	if z.Sign() == 0 {
		return o.SetFloat64(1)
	}
	if z.IsInf() {
		if z.Sign() < 0 {
			return o.Set(&gzero)
		}
		return o.Set(z)
	}

	p := o
	if p == z {
		// We need z for Newton's algorithm, so ensure we don't overwrite it.
		p = new(big.Float).SetPrec(z.Prec())
	}
	// try to get initial estimate using IEEE-754 math
	// TODO: save work (and an import of math) by checking the exponent of z
	zf, _ := z.Float64()
	zf = math.Exp(zf)
	if math.IsInf(zf, 1) || zf == 0 {
		// too big or too small for IEEE-754 math,
		// perform argument reduction using
		//     e^{2z} = (e^z)²
		// TODO: use MantExp instead of Mul
		halfZ := new(big.Float).SetPrec(p.Prec()+64).Mul(z, big.NewFloat(0.5))
		// TODO: avoid recursion
		halfExp := Exp(halfZ, halfZ)
		return p.Mul(halfExp, halfExp)
	}
	// we got a nice IEEE-754 estimate
	guess := big.NewFloat(zf)

	// f(t)/f'(t) = t*(log(t) - z)
	f := func(t *big.Float) *big.Float {
		p.Sub(Log(new(big.Float).Copy(t)), z)
		return p.Mul(p, t)
	}

	x := newton(f, guess, z.Prec()) // TODO: make newton operate in place

	return o.Set(x)
}
