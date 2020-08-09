### Floats 

Package bigfloat provides the implementation of a few additional operations (exponentiation, basic elementary functions) for the standard library `big.Float` type.

[![GoDoc](https://godoc.org/github.com/zephyrtronium/bigfloat?status.png)](https://godoc.org/github.com/zephyrtronium/bigfloat)

This package is a fork of [https://github.com/ALTree/bigfloat]. The general API was changed to use the output argument convention of `math/big`, and there are a number of other changes besides.

#### Install

With Go version 1.10 or higher installed:

```
go get github.com/zephyrtronium/bigfloat
```

#### Example

```go
package main

import (
	"fmt"
	"math/big"

	"github.com/zephyrtronium/bigfloat"
)

// In this example, we'll compute the value of the
// trascendental number 2 ** √2, also known as
// the Gelfond–Schneider constant, to 1000 bits.
func main() {
	// Work with 1000 binary digits of precision.
	const prec = 1000

	two := big.NewFloat(2).SetPrec(prec)
	sqrtTwo := new(big.Float).Sqrt(two)

	// Compute 2 ** √2
	// Pow uses the first argument's precision, or the greater of the others'
	// if the first's is zero like here.
	gsc := bigfloat.Pow(new(big.Float), two, sqrtTwo)

	// Print gsc, truncated to 60 decimal digits.
	fmt.Printf("gsc = %.60f...\n", gsc)
}
```

outputs
```
gsc = 2.665144142690225188650297249873139848274211313714659492835980...
```

#### Documentation

See https://godoc.org/github.com/zephyrtronium/bigfloat
