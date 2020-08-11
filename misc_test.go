package bigfloat

import (
	"fmt"
	"math/big"
	"sync"
	"testing"
)

const maxPrec uint = 1100

func TestAGM(t *testing.T) {
	for _, test := range []struct {
		a, b string
		want string
	}{
		// 350 decimal digits are enough to give us up to 1000 binary digits
		{"1", "2", "1.4567910310469068691864323832650819749738639432213055907941723832679264545802509002574737128184484443281894018160367999355762430743401245116912132499522793768970211976726893728266666782707432902072384564600963133367494416649516400826932239086263376738382410254887262645136590660408875885100466728130947439789355129117201754471869564160356411130706061"},
		{"1", "10", "4.2504070949322748617281643183731348667984678641901928596701476622237553127409037845252854607876171790458817135897668652366410690187825866854343005714304399718866701345600268795095037823053677248108795697049522041225723229732458947507697835936406527028150257238518982793084569470658500853106997941082919334694146843915361847332301248942222685517896377"},
		{"1", "0.125", "0.45196952219967034359164911331276507645541557018306954112635037493237190371123433961098897571407153216488726488616781446636283304514042965741376539315003644325377859387794608118242990700589889155408232061013871480906595147189700268152276449512798584772002737950386745259435790965051247641106770187776231088478906739003673011639874297764052324720923824"},
		{"1", "0.00390625", "0.2266172673264813935990249059047521131153183423554951008357647589399579243281007098800682366778894106068183449922373565084840603788091294841822891406755449218057751291845474188560350241555526734834267320629182988862200822134426714354129001630331838172767684623648755579758508073234772093745831056731263684472818466567279847347734121500617411676068370"},
		{"1", "0.0001220703125", "0.15107867088555894565277006051956059212554039802503247524478909254186086852737399490629222674071181480492157167137547694132610166031526264375084434300568336411139925857454913414480542768807718797335060713475211709310835676172131569048902323084439330888400622327072954342544508199547787750415198261456314278054748992781108231991187512975110547417178045"},
	} {
		for _, prec := range []uint{24, 53, 64, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000} {
			want := new(big.Float).SetPrec(prec)
			want.Parse(test.want, 10)

			a := new(big.Float).SetPrec(prec)
			a.Parse(test.a, 10)

			b := new(big.Float).SetPrec(prec)
			b.Parse(test.b, 10)

			z := AGM(new(big.Float), a, b)

			if z.Cmp(want) != 0 {
				t.Errorf("prec = %d, Agm(%v, %v) =\ngot  %g;\nwant %g", prec, test.a, test.b, z, want)
			}
		}
	}
}

func TestPi(t *testing.T) {
	enablePiCache = false
	piStr := "3.1415926535897932384626433832795028841971693993751058209749445923078164062862089986280348253421170679821480865132823066470938446095505822317253594081284811174502841027019385211055596446229489549303819644288109756659334461284756482337867831652712019091456485669234603486104543266482133936072602491412737245870066063155881748815209209628292540917153644"
	for _, prec := range []uint{24, 53, 64, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000} {

		want := new(big.Float).SetPrec(prec)
		want.Parse(piStr, 10)

		z := Pi(new(big.Float).SetPrec(prec))

		if z.Cmp(want) != 0 {
			t.Errorf("Pi(%d) =\ngot  %g;\nwant %g", prec, z, want)
		}
	}
	enablePiCache = true
}

func TestPiConcurrent(t *testing.T) {
	if !enablePiCache {
		t.SkipNow()
	}
	const piStr = "3.1415926535897932384626433832795028841971693993751058209749445923078164062862089986280348253421170679821480865132823066470938446095505822317253594081284811174502841027019385211055596446229489549303819644288109756659334461284756482337867831652712019091456485669234603486104543266482133936072602491412737245870066063155881748815209209628292540917153644"
	// The pi cache starts at a precision of 1024, so to make this test more
	// meaningful, we'll cheat and set it to a zero-precision value.
	cached := loadPi()
	piCache.Store(new(big.Float))
	defer piCache.Store(cached)
	cases := []uint{24, 53, 64, 100, 200, 300, 400, 500, 600, 700, 800, 900, 1000}
	const procs = 100
	var wg sync.WaitGroup
	wg.Add(procs)
	for i := 0; i < procs; i++ {
		go func() {
			for _, prec := range cases {
				want := new(big.Float).SetPrec(prec)
				want.Parse(piStr, 10)
				z := Pi(new(big.Float).SetPrec(prec))
				if z.Cmp(want) != 0 {
					t.Errorf("Pi(%d) = \ngot  %g;\nwant %g", prec, z, want)
				}
			}
			wg.Done()
		}()
	}
	wg.Wait()
}

func TestRound(t *testing.T) {
	cases := []struct {
		o, z *big.Float
		r    [6]*big.Float
		op   uint
	}{
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(0),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(0),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(0),
				big.ToNegativeInf: big.NewFloat(0),
				big.ToPositiveInf: big.NewFloat(0),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(1),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(1),
				big.ToNearestAway: big.NewFloat(1),
				big.ToZero:        big.NewFloat(1),
				big.AwayFromZero:  big.NewFloat(1),
				big.ToNegativeInf: big.NewFloat(1),
				big.ToPositiveInf: big.NewFloat(1),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(0.25),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(0),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(1),
				big.ToNegativeInf: big.NewFloat(0),
				big.ToPositiveInf: big.NewFloat(1),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(-0.25),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(0),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(-1),
				big.ToNegativeInf: big.NewFloat(-1),
				big.ToPositiveInf: big.NewFloat(0),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(0.5),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(1),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(1),
				big.ToNegativeInf: big.NewFloat(0),
				big.ToPositiveInf: big.NewFloat(1),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(-0.5),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(-1),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(-1),
				big.ToNegativeInf: big.NewFloat(-1),
				big.ToPositiveInf: big.NewFloat(0),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(0.75),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(1),
				big.ToNearestAway: big.NewFloat(1),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(1),
				big.ToNegativeInf: big.NewFloat(0),
				big.ToPositiveInf: big.NewFloat(1),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(-0.75),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(-1),
				big.ToNearestAway: big.NewFloat(-1),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(-1),
				big.ToNegativeInf: big.NewFloat(-1),
				big.ToPositiveInf: big.NewFloat(0),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(1.25),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(1),
				big.ToNearestAway: big.NewFloat(1),
				big.ToZero:        big.NewFloat(1),
				big.AwayFromZero:  big.NewFloat(2),
				big.ToNegativeInf: big.NewFloat(1),
				big.ToPositiveInf: big.NewFloat(2),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(-1.25),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(-1),
				big.ToNearestAway: big.NewFloat(-1),
				big.ToZero:        big.NewFloat(-1),
				big.AwayFromZero:  big.NewFloat(-2),
				big.ToNegativeInf: big.NewFloat(-2),
				big.ToPositiveInf: big.NewFloat(-1),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(1.5),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(2),
				big.ToNearestAway: big.NewFloat(2),
				big.ToZero:        big.NewFloat(1),
				big.AwayFromZero:  big.NewFloat(2),
				big.ToNegativeInf: big.NewFloat(1),
				big.ToPositiveInf: big.NewFloat(2),
			},
		},
		{
			o: new(big.Float), op: 53,
			z: big.NewFloat(-1.5),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(-2),
				big.ToNearestAway: big.NewFloat(-2),
				big.ToZero:        big.NewFloat(-1),
				big.AwayFromZero:  big.NewFloat(-2),
				big.ToNegativeInf: big.NewFloat(-2),
				big.ToPositiveInf: big.NewFloat(-1),
			},
		},
		{
			o: new(big.Float).SetPrec(1), op: 1,
			z: big.NewFloat(0),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(0),
				big.ToNearestAway: big.NewFloat(0),
				big.ToZero:        big.NewFloat(0),
				big.AwayFromZero:  big.NewFloat(0),
				big.ToNegativeInf: big.NewFloat(0),
				big.ToPositiveInf: big.NewFloat(0),
			},
		},
		{
			o: new(big.Float).SetPrec(1).SetMode(big.ToPositiveInf), op: 1,
			z: big.NewFloat(12),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(16),
				big.ToNearestAway: big.NewFloat(16),
				big.ToZero:        big.NewFloat(16),
				big.AwayFromZero:  big.NewFloat(16),
				big.ToNegativeInf: big.NewFloat(16),
				big.ToPositiveInf: big.NewFloat(16),
			},
		},
		{
			o: new(big.Float).SetPrec(1).SetMode(big.ToNegativeInf), op: 1,
			z: big.NewFloat(12),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(8),
				big.ToNearestAway: big.NewFloat(8),
				big.ToZero:        big.NewFloat(8),
				big.AwayFromZero:  big.NewFloat(8),
				big.ToNegativeInf: big.NewFloat(8),
				big.ToPositiveInf: big.NewFloat(8),
			},
		},
		{
			o: new(big.Float).SetPrec(128), op: 128,
			z: big.NewFloat(3.5),
			r: [...]*big.Float{
				big.ToNearestEven: big.NewFloat(4),
				big.ToNearestAway: big.NewFloat(4),
				big.ToZero:        big.NewFloat(3),
				big.AwayFromZero:  big.NewFloat(4),
				big.ToNegativeInf: big.NewFloat(3),
				big.ToPositiveInf: big.NewFloat(4),
			},
		},
	}
	modes := []big.RoundingMode{big.ToNearestEven, big.ToNearestAway, big.ToZero, big.AwayFromZero, big.ToNegativeInf, big.ToPositiveInf}
	for _, c := range cases {
		t.Run(fmt.Sprintf("%gto%d", c.z, c.o.Prec()), func(t *testing.T) {
			for _, mode := range modes {
				Round(c.o, c.z, mode)
				if c.o.Cmp(c.r[mode]) != 0 {
					t.Errorf("wrong result: want %g, got %g", c.r[mode], c.o)
				}
				if c.o.Prec() != c.op {
					t.Errorf("result has wrong precision: want %d, got %d", c.o.Prec(), c.op)
				}
			}
		})
	}
}

// ---------- Benchmarks ----------

func BenchmarkAGM(b *testing.B) {
	for _, prec := range []uint{1e2, 1e3, 1e4, 1e5} {
		x := new(big.Float).SetPrec(prec).SetFloat64(1)
		y := new(big.Float).SetPrec(prec).SetFloat64(0.125)
		o := new(big.Float).SetPrec(prec + 64).SetPrec(0)
		b.Run(fmt.Sprintf("%v", prec), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				AGM(o, x, y)
			}
		})
	}
}

func BenchmarkPi(b *testing.B) {
	enablePiCache = false
	p := new(big.Float)
	for _, prec := range []uint{1e2, 1e3, 1e4, 1e5} {
		p.SetPrec(prec)
		b.Run(fmt.Sprintf("%v", prec), func(b *testing.B) {
			b.ReportAllocs()
			for n := 0; n < b.N; n++ {
				Pi(p)
			}
		})
	}
}
