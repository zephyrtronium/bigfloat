package bigfloat

import (
	"math/big"
	"sync"
	"sync/atomic"
)

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

	for {
		o.Copy(a2)
		quicksh(a2, a2.Add(a2, b2), -1)
		b2.Sqrt(b2.Mul(b2, o))
		if o.Sub(a2, b2).Cmp(lim) == -1 {
			break
		}
	}

	return o.Copy(a2).SetPrec(prec)
}

// Round sets o to z rounded to the nearest integer as constrained by mode and
// returns o. If o's precision is zero, then it is given z's precision.
// Otherwise, if o has insufficient precision to represent the result of
// rounding z, the result will be rounded again according to o's rounding mode.
// As a special case, if z is infinite, o is set to the same infinity. If o and
// z are the same, Round does not allocate.
func Round(o, z *big.Float, mode big.RoundingMode) *big.Float {
	if z.IsInt() || z.IsInf() {
		// This branch notably includes z == 0 and z.Prec() == 0.
		return o.Set(z)
	}
	if o.Prec() == 0 {
		o.SetPrec(z.Prec())
	}
	exp := z.MantExp(nil)
	if exp <= 0 {
		// z ∈ (-1, 1) \ {0}. There is no trick to pull off this rounding.
		switch mode {
		case big.ToNearestEven:
			return round0even(o, z)
		case big.ToNearestAway:
			return round0away(o, z)
		case big.ToZero:
			return o.Set(&gzero)
		case big.AwayFromZero:
			if z.Signbit() {
				return o.Set(&gonem)
			}
			return o.Set(&gonep)
		case big.ToNegativeInf:
			if z.Signbit() {
				return o.Set(&gonem)
			}
			return o.Set(&gzero)
		case big.ToPositiveInf:
			if z.Signbit() {
				return o.Set(&gzero)
			}
			return o.Set(&gonep)
		default:
			panic("bigfloat: unknown rounding mode " + mode.String())
		}
	}
	// z has a nonzero integer part. Give o exactly enough precision to
	// represent that integer part, then set it to z and restore its precision.
	// But first, check that o actually needs to shrink to do this.
	if o.Prec() <= uint(exp) {
		return o.Set(z)
	}
	defer o.SetMode(o.Mode())
	o.SetMode(mode)
	p := o.Prec()
	return o.SetPrec(uint(exp)).Set(z).SetPrec(p)
}

// round0even sets o to -1 if z < -0.5, 0 if -0.5 <= z <= 0.5, or 1 if 0.5 < z.
func round0even(o, z *big.Float) *big.Float {
	if z.Signbit() {
		if z.Cmp(&ghalfm) < 0 {
			return o.Set(&gonem)
		}
		return o.Set(&gzero)
	}
	if z.Cmp(&ghalfp) > 0 {
		return o.Set(&gonep)
	}
	return o.Set(&gzero)
}

// round0away sets o to -1 if z <= -0.5, 0 if -0.5 < z < 0.5, or 1 if 0.5 <= z.
func round0away(o, z *big.Float) *big.Float {
	if z.Signbit() {
		if z.Cmp(&ghalfm) <= 0 {
			return o.Set(&gonem)
		}
		return o.Set(&gzero)
	}
	if z.Cmp(&ghalfp) >= 0 {
		return o.Set(&gonep)
	}
	return o.Set(&gzero)
}

var piCache atomic.Value
var enablePiCache bool = true
var piMu sync.Mutex // writers only

func init() {
	if !enablePiCache {
		return
	}
	pi, _, err := new(big.Float).SetPrec(1024).Parse("3."+
		"14159265358979323846264338327950288419716939937510"+
		"58209749445923078164062862089986280348253421170679"+
		"82148086513282306647093844609550582231725359408128"+
		"48111745028410270193852110555964462294895493038196"+
		"44288109756659334461284756482337867831652712019091"+
		"45648566923460348610454326648213393607260249141273"+
		"72458700660631558817488152092096282925409171536444", 10)
	if err != nil {
		panic(err)
	}
	piCache.Store(pi)
}

// loadPi returns the current cached pi value. It may panic if enablePiCache is
// false. Use cachedPi or Pi instead; this is just a convenience function for
// those safe wrappers.
func loadPi() *big.Float {
	return piCache.Load().(*big.Float)
}

// cachedPi returns the cached pi value with at least prec precision. If the pi
// cache is enabled and has a precision of at least prec, then this does not
// allocate. The returned value must not be modified. It is safe to call this
// concurrently.
func cachedPi(prec uint) *big.Float {
	if !enablePiCache {
		return Pi(new(big.Float).SetPrec(prec))
	}
	pi := piCache.Load().(*big.Float)
	if pi.Prec() >= prec {
		return pi
	}

	// The current cached value doesn't have enough precision. Calculate a new
	// pi value.
	piMu.Lock()
	defer piMu.Unlock()
	// It's possible another goroutine obtained a more precise pi value while
	// we were locking piMu. Re-check the cached value.
	pi = piCache.Load().(*big.Float)
	if pi.Prec() >= prec {
		return pi
	}
	pi = piCalc(new(big.Float).SetPrec(prec))
	piCache.Store(pi)
	return pi
}

// Pi sets a to π to a's precision (even if a's precision is zero) and
// returns a.
func Pi(a *big.Float) *big.Float {
	prec := a.Prec()
	if prec == 0 {
		// Zero-precision floats represent only ±0 or ±inf.
		return a.Set(&gzero)
	}
	if enablePiCache {
		pi := loadPi()
		if prec <= pi.Prec() {
			return a.Set(pi)
		}
	}
	piCalc(a)
	if enablePiCache {
		piMu.Lock()
		defer piMu.Unlock()
		if loadPi().Prec() < prec {
			piCache.Store(new(big.Float).Copy(a))
		}
	}
	return a
}

// piCalc performs the actual computation to obtain a value for π.
func piCalc(a *big.Float) *big.Float {
	prec := a.Prec()

	// Following R. P. Brent, Multiple-precision zero-finding
	// methods and the complexity of elementary function evaluation,
	// in Analytic Computational Complexity, Academic Press,
	// New York, 1975, Section 8.

	sqrt2 := new(big.Float).SetPrec(prec + 64).Set(&gtwop)
	sqrt2.Sqrt(sqrt2)
	// initialization
	a.SetFloat64(1).SetPrec(prec + 64)         // a = 1
	b := quicksh(new(big.Float), sqrt2, -1)    // b = 1/√2
	t := big.NewFloat(0.25).SetPrec(prec + 64) // t = 1/4
	x := big.NewFloat(1).SetPrec(prec + 64)    // x = 1
	// limit is 2**(-prec)
	lim := new(big.Float)
	lim.SetMantExp(big.NewFloat(1).SetPrec(prec+64), -int(prec+1))
	y := new(big.Float)
	for y.Sub(a, b).Cmp(lim) != -1 { // assume a > b
		y.Copy(a)
		quicksh(a, a.Add(a, b), -1) // a = (a+b)/2
		b.Sqrt(b.Mul(b, y))         // b = √(ab)

		y.Sub(a, y)           // y = a - y
		y.Mul(y, y).Mul(y, x) // y = x(a-y)²
		t.Sub(t, y)           // t = t - x(a-y)²
		quicksh(x, x, 1)      // x = 2x
	}
	a.Mul(a, a).Quo(a, t) // π = a² / t
	return a.SetPrec(prec)
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

// quicksh efficiently multiplies z by 2**n and sets o to the result. o's
// precision and rounding mode are overwritten.
func quicksh(o, z *big.Float, n int) *big.Float {
	exp := z.MantExp(o)
	return o.SetMantExp(o, exp+n)
}

// Global variables that are never modified.
var (
	gzero  big.Float // +0
	ghalfp = *big.NewFloat(0.5)
	ghalfm = *big.NewFloat(-0.5)
	gonep  = *big.NewFloat(1)
	gonem  = *big.NewFloat(-1)
	gtwop  = *big.NewFloat(2)
)

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
