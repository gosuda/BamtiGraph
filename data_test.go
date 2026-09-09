// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"math/big"
	"testing"
)

func TestCounterHighPrecision(t *testing.T) {
	o := DefaultCounterOptions()
	o.Factor = 8
	o.CounterBits = 64
	b := uint64(1) << 60
	r, e := CounterRate([]float64{0, 1, 2}, []uint64{b, b + 1, b + 3}, o)
	if e != nil || !math.IsNaN(r.Values[0]) || r.Values[1] != 8 || r.Values[2] != 16 {
		t.Fatal(r, e)
	}
}
func TestCounterWraps(t *testing.T) {
	for _, bits := range []int{8, 32, 64} {
		t.Run(string(rune('A'+bits)), func(t *testing.T) {
			o := DefaultCounterOptions()
			o.CounterBits = bits
			o.OnDecrease = "wrap"
			maxVal := ^uint64(0)
			if bits < 64 {
				maxVal = (uint64(1) << bits) - 1
			}
			r, e := CounterRate([]float64{0, 1}, []uint64{maxVal - 1, 3}, o)
			if e != nil || r.Values[1] != 5 {
				t.Fatal(r, e)
			}
		})
	}
}
func TestCounterMissingResetMaxRate(t *testing.T) {
	o := DefaultCounterOptions()
	o.Valid = []bool{true, false, true, true, true}
	o.MaxRate = Float(10)
	r, e := CounterRate([]float64{0, 1, 2, 3, 4}, []uint64{100, 200, 300, 1, 50}, o)
	if e != nil {
		t.Fatal(e)
	}
	for _, v := range r.Values {
		if !math.IsNaN(v) {
			t.Fatal(r)
		}
	}
}
func TestCounter128Bit(t *testing.T) {
	o := DefaultCounterOptions()
	o.CounterBits = 128
	b := new(big.Int).Lsh(big.NewInt(1), 100)
	c := new(big.Int).Add(b, big.NewInt(3))
	before := b.String()
	r, e := CounterRateBig([]float64{0, 1}, []*big.Int{b, c}, o)
	if e != nil || r.Values[1] != 3 || b.String() != before {
		t.Fatal(r, e)
	}
}
func TestCounterBigWrap(t *testing.T) {
	o := DefaultCounterOptions()
	o.CounterBits = 128
	o.OnDecrease = "wrap"
	a := new(big.Int).Sub(new(big.Int).Lsh(big.NewInt(1), 128), big.NewInt(3))
	r, e := CounterRateBig([]float64{0, 1}, []*big.Int{a, big.NewInt(2)}, o)
	if e != nil || r.Values[1] != 5 {
		t.Fatal(r, e)
	}
}
func TestCounterInvalidOptions(t *testing.T) {
	for _, c := range []struct {
		name   string
		change func(*CounterOptions)
	}{{"factor", func(o *CounterOptions) { o.Factor = 0 }}, {"mask", func(o *CounterOptions) { o.Valid = []bool{true} }}, {"wrap", func(o *CounterOptions) { o.OnDecrease = "wrap" }}, {"bits", func(o *CounterOptions) { o.CounterBits = 128 }}, {"max", func(o *CounterOptions) { o.MaxRate = Float(-1) }}} {
		t.Run(c.name, func(t *testing.T) {
			o := DefaultCounterOptions()
			c.change(&o)
			if _, e := CounterRate([]float64{0, 1}, []uint64{1, 2}, o); e == nil {
				t.Fatal("accepted")
			}
		})
	}
	o := DefaultCounterOptions()
	o.CounterBits = 8
	if _, e := CounterRate([]float64{0, 1}, []uint64{256, 257}, o); e == nil {
		t.Fatal("range")
	}
}
func TestCounterFloat(t *testing.T) {
	o := DefaultCounterOptions()
	r, e := CounterRateFloat([]float64{0, 2, 4}, []float64{1.5, 4, 3}, o)
	if e != nil || r.Values[1] != 1.25 || !math.IsNaN(r.Values[2]) {
		t.Fatal(r, e)
	}
}
func TestAggregateMethods(t *testing.T) {
	for _, c := range []struct {
		method string
		want   float64
	}{{"mean", 2}, {"sum", 6}, {"min", 1}, {"max", 3}, {"last", 3}} {
		t.Run(c.method, func(t *testing.T) {
			o := DefaultAggregateOptions()
			o.Interval = 3
			o.Method = c.method
			r, e := Aggregate([]float64{0, 1, 2}, []float64{1, 2, 3}, o)
			if e != nil || len(r.Values) != 1 || r.Values[0] != c.want {
				t.Fatal(r, e)
			}
		})
	}
}
func TestAggregateEmptyBuckets(t *testing.T) {
	o := DefaultAggregateOptions()
	o.Interval = 10
	r, e := Aggregate([]float64{0, 30}, []float64{1, 2}, o)
	if e != nil || len(r.Values) != 4 || !math.IsNaN(r.Values[1]) || !math.IsNaN(r.Values[2]) || r.Timestamps[3] != 30 {
		t.Fatal(r, e)
	}
}
func TestAggregateCoverageAndLastGap(t *testing.T) {
	o := DefaultAggregateOptions()
	o.Interval = 10
	o.ExpectedStep = 1
	o.MinCoverage = .8
	r, e := Aggregate([]float64{0, 1}, []float64{1, 2}, o)
	if e != nil || !math.IsNaN(r.Values[0]) {
		t.Fatal(r, e)
	}
	o.MinCoverage = 0
	o.Method = "last"
	r, e = Aggregate([]float64{0, 1}, []float64{1, math.NaN()}, o)
	if e != nil || !math.IsNaN(r.Values[0]) {
		t.Fatal(r, e)
	}
}
func TestAggregateBucketBoundary(t *testing.T) {
	o := DefaultAggregateOptions()
	o.Interval = 10
	r, e := Aggregate([]float64{-1, 0, 9.5, 10}, []float64{1, 2, 4, 8}, o)
	if e != nil || len(r.Values) != 3 || r.Timestamps[0] != -10 || r.Values[1] != 3 || r.Values[2] != 8 {
		t.Fatal(r, e)
	}
}
func TestAggregateGuards(t *testing.T) {
	o := DefaultAggregateOptions()
	o.MaxBuckets = 1
	if _, e := Aggregate([]float64{0, 1000}, []float64{1, 2}, o); e == nil {
		t.Fatal("count guard")
	}
	o = DefaultAggregateOptions()
	o.Method = "sum"
	if _, e := Aggregate([]float64{0, 1}, []float64{math.MaxFloat64, math.MaxFloat64}, o); e == nil {
		t.Fatal("overflow")
	}
}
