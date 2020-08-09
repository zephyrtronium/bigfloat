package bigfloat

import "math/big"

// AGM sets o to the limit of the arithmetic-geometric mean progression of a
// and b, to o's precision, and returns o. If o's precision is zero, then it is
// given the larger of a's and b's precision.
func AGM(o, a, b *big.Float) *big.Float {
	prec := o.Prec()
	if prec == 0 {
		if a.Prec() >= b.Prec() {
			prec = a.Prec()
		} else {
			prec = b.Prec()
		}
	}
	o.SetPrec(prec + 64)

	// do not overwrite a and b
	a2 := new(big.Float).Copy(a).SetPrec(prec + 64)
	b2 := new(big.Float).Copy(b).SetPrec(prec + 64)

	if a2.Cmp(b2) == -1 {
		a2, b2 = b2, a2
	}
	// a2 >= b2

	// set lim to 2**(-prec)
	lim := new(big.Float)
	lim.SetMantExp(big.NewFloat(1).SetPrec(prec+64), -int(prec+1))

	half := big.NewFloat(0.5)

	for {
		o.Copy(a2)
		a2.Add(a2, b2).Mul(a2, half)
		b2.Sqrt(b2.Mul(b2, o))
		if o.Sub(a2, b2).Cmp(lim) == -1 {
			break
		}
	}

	return o.Copy(a2).SetPrec(prec)
}

var piCache *big.Float // TODO: synchronize
var enablePiCache bool = true

func init() {
	if !enablePiCache {
		return
	}
	piCache, _, _ = new(big.Float).SetPrec(1024).Parse("3."+
		"14159265358979323846264338327950288419716939937510"+
		"58209749445923078164062862089986280348253421170679"+
		"82148086513282306647093844609550582231725359408128"+
		"48111745028410270193852110555964462294895493038196"+
		"44288109756659334461284756482337867831652712019091"+
		"45648566923460348610454326648213393607260249141273"+
		"72458700660631558817488152092096282925409171536444", 10)
}

// Pi sets a to pi to a's precision (including if a's precision is zero) and
// returns a.
func Pi(a *big.Float) *big.Float {
	prec := a.Prec()
	if prec == 0 {
		// Zero-precision floats represent only ±0 or ±inf.
		return a.Set(&gzero)
	}
	if enablePiCache && prec <= piCache.Prec() {
		return a.Set(piCache)
	}

	// Following R. P. Brent, Multiple-precision zero-finding
	// methods and the complexity of elementary function evaluation,
	// in Analytic Computational Complexity, Academic Press,
	// New York, 1975, Section 8.

	half := big.NewFloat(0.5)
	two := big.NewFloat(2).SetPrec(prec + 64)
	sqrt2 := new(big.Float).Sqrt(two)

	// initialization
	a.SetFloat64(1).SetPrec(prec + 64)         // a = 1
	b := new(big.Float).Mul(sqrt2, half)       // b = 1/√2
	t := big.NewFloat(0.25).SetPrec(prec + 64) // t = 1/4
	x := big.NewFloat(1).SetPrec(prec + 64)    // x = 1

	// limit is 2**(-prec)
	lim := new(big.Float)
	lim.SetMantExp(big.NewFloat(1).SetPrec(prec+64), -int(prec+1))

	// temp variables
	y := new(big.Float)
	for y.Sub(a, b).Cmp(lim) != -1 { // assume a > b
		y.Copy(a)
		a.Add(a, b).Mul(a, half) // a = (a+b)/2
		b.Sqrt(b.Mul(b, y))      // b = √(ab)

		y.Sub(a, y)           // y = a - y
		y.Mul(y, y).Mul(y, x) // y = x(a-y)²
		t.Sub(t, y)           // t = t - x(a-y)²
		x.Mul(x, two)         // x = 2x
	}

	a.Mul(a, a).Quo(a, t) // π = a² / t
	a.SetPrec(prec)

	if enablePiCache {
		piCache.Copy(a)
	}
	return a
}

// cachedPi returns the cached pi value with at least prec precision. If the pi
// cache is enabled and has a precision of at least prec, then this does not
// allocate. The returned value must not be modified.
func cachedPi(prec uint) *big.Float {
	if !enablePiCache {
		return Pi(new(big.Float).SetPrec(prec))
	}
	if piCache.Prec() >= prec {
		return piCache
	}
	return Pi(piCache) // updates piCache
}

// returns an approximate (to precision dPrec) solution to
//    f(t) = 0
// using the Newton Method.
// fOverDf needs to be a fuction returning f(t)/f'(t).
// t must not be changed by fOverDf.
// guess is the initial guess (and it's not preserved).
func newton(fOverDf func(z *big.Float) *big.Float, guess *big.Float, dPrec uint) *big.Float {

	prec, guard := guess.Prec(), uint(64)
	guess.SetPrec(prec + guard)

	for prec < 2*dPrec {
		guess.Sub(guess, fOverDf(guess))
		prec *= 2
		guess.SetPrec(prec + guard)
	}

	return guess.SetPrec(dPrec)
}

// gzero is a global zero that is never modified.
var gzero big.Float

// An ErrNaN panic is raised by an operation that would lead to a NaN under
// IEEE-754 rules. ErrNaN implements the error interface, and it unwraps to a
// big.ErrNaN value with an empty message.
type ErrNaN struct {
	msg string
}

func (err ErrNaN) Error() string {
	return err.msg
}

func (err ErrNaN) Unwrap() error {
	return big.ErrNaN{}
}
