package bigfloat

import "math/big"

// Pow sets o to z**w to o's precision and returns o. Panics with ErrNaN when
// z is negative. If o's precision is zero, then it is given the larger
// of z's and w's precision.
func Pow(o, z, w *big.Float) *big.Float {
	if o.Prec() == 0 {
		if z.Prec() >= w.Prec() {
			o.SetPrec(z.Prec())
		} else {
			o.SetPrec(w.Prec())
		}
	}
	if z.Signbit() {
		panic(ErrNaN{msg: "Pow: negative base"})
	}

	// Pow(z, 0) = 1.0
	if w.Sign() == 0 {
		return big.NewFloat(1).SetPrec(z.Prec())
	}

	// Pow(z, 1) = z
	// Pow(+Inf, n) = +Inf
	if w.Cmp(big.NewFloat(1)) == 0 || z.IsInf() {
		return new(big.Float).Copy(z)
	}

	// Pow(z, -w) = 1 / Pow(z, w)
	// TODO: is this actually better? Lots of allocations...
	// if w.Sign() < 0 {
	// 	zExt := new(big.Float).Copy(z).SetPrec(z.Prec() + 64)
	// 	wNeg := new(big.Float).Neg(w)
	// 	return o.Quo(big.NewFloat(1), Pow(o, zExt, wNeg))
	// }

	// w integer fast path (disabled because introduces rounding
	// errors)
	if false && w.IsInt() {
		wi, _ := w.Int64()
		return powInt(z, int(wi))
	}

	// compute w**z as exp(z log(w))
	o.SetPrec(o.Prec() + 64) // guard digits
	logZ := Log(new(big.Float).SetPrec(z.Prec()+64), z)
	o.Mul(new(big.Float).Set(w).SetPrec(z.Prec()+64), logZ)
	o = Exp(o, o)
	return o.SetPrec(o.Prec() - 64)

}

// fast path for z**w when w is an integer
func powInt(z *big.Float, w int) *big.Float {

	// get mantissa and exponent of z
	mant := new(big.Float)
	exp := z.MantExp(mant)

	// result's exponent
	exp = exp * w

	// result's mantissa
	x := big.NewFloat(1).SetPrec(z.Prec())

	// Classic right-to-left binary exponentiation
	for w > 0 {
		if w%2 == 1 {
			x.Mul(x, mant)
		}
		w >>= 1
		mant.Mul(mant, mant)
	}

	return new(big.Float).SetMantExp(x, exp)
}
