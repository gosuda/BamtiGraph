// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"math"
	"math/big"
)

type Samples struct{ Timestamps, Values []float64 }
type AggregateOptions struct {
	Interval    float64
	Method      string
	Origin      float64
	MinCoverage float64
	// Zero uses observed sample count. Positive values define expected cadence.
	ExpectedStep float64
	MaxBuckets   int
}

func DefaultAggregateOptions() AggregateOptions {
	return AggregateOptions{Interval: 300, Method: "mean", MaxBuckets: 1_000_000}
}
func Aggregate(timestamps, values []float64, o AggregateOptions) (Samples, error) {
	if e := validateSamples(timestamps, values); e != nil {
		return Samples{}, e
	}
	if !finite(o.Interval) || o.Interval <= 0 || !validEpoch(o.Origin) || !finite(o.MinCoverage) || o.MinCoverage < 0 || o.MinCoverage > 1 || !finite(o.ExpectedStep) || o.ExpectedStep < 0 || o.MaxBuckets < 1 || o.MaxBuckets > 10_000_000 {
		return Samples{}, fmt.Errorf("invalid aggregate options")
	}
	switch o.Method {
	case "mean", "min", "max", "last", "sum":
	default:
		return Samples{}, fmt.Errorf("unsupported aggregation method")
	}
	if len(timestamps) == 0 {
		return Samples{[]float64{}, []float64{}}, nil
	}
	first, last := math.Floor((timestamps[0]-o.Origin)/o.Interval), math.Floor((timestamps[len(timestamps)-1]-o.Origin)/o.Interval)
	n := last - first + 1
	if !finite(n) || n < 1 || n > float64(o.MaxBuckets) || math.Abs(first) > 9e15 || math.Abs(last) > 9e15 {
		return Samples{}, fmt.Errorf("aggregate bucket count/precision guard exceeded")
	}
	out := Samples{make([]float64, int(n)), make([]float64, int(n))}
	cursor := 0
	for k := range out.Values {
		idx := first + float64(k)
		start := cursor
		for cursor < len(timestamps) && math.Floor((timestamps[cursor]-o.Origin)/o.Interval) == idx {
			cursor++
		}
		all := values[start:cursor]
		count := 0
		lo, hi := math.Inf(1), math.Inf(-1)
		for _, v := range all {
			if !math.IsNaN(v) {
				count++
				lo = math.Min(lo, v)
				hi = math.Max(hi, v)
			}
		}
		denom := float64(len(all))
		if o.ExpectedStep > 0 {
			denom = math.Max(denom, o.Interval/o.ExpectedStep)
		}
		v := missing()
		if count > 0 && denom > 0 && float64(count)/denom >= o.MinCoverage {
			switch o.Method {
			case "mean":
				v = compensatedSum(all, float64(count))
			case "sum":
				v = compensatedSum(all, 1)
			case "min":
				v = lo
			case "max":
				v = hi
			case "last":
				v = all[len(all)-1]
			}
			if o.Method != "last" && !finite(v) {
				return Samples{}, fmt.Errorf("aggregation overflow in bucket %d", k)
			}
		}
		out.Timestamps[k] = o.Origin + idx*o.Interval
		if !validEpoch(out.Timestamps[k]) {
			return Samples{}, fmt.Errorf("bucket time outside supported range")
		}
		if k > 0 && out.Timestamps[k] <= out.Timestamps[k-1] {
			return Samples{}, fmt.Errorf("bucket timestamps lost floating-point precision")
		}
		out.Values[k] = v
	}
	return out, nil
}

type CounterOptions struct {
	Factor     float64
	OnDecrease string
	// 0 disables width validation (allowed only when not wrapping).
	CounterBits int
	MaxRate     *float64
	// nil treats every counter as present; false invalidates both adjacent rates.
	Valid []bool
}

func DefaultCounterOptions() CounterOptions { return CounterOptions{Factor: 1, OnDecrease: "gap"} }
func (o CounterOptions) validate(n, maxBits int) error {
	if !finite(o.Factor) || o.Factor <= 0 || o.CounterBits < 0 || o.CounterBits > maxBits {
		return fmt.Errorf("invalid counter factor or bit width")
	}
	if o.OnDecrease != "gap" && o.OnDecrease != "wrap" {
		return fmt.Errorf("OnDecrease must be gap or wrap")
	}
	if o.OnDecrease == "wrap" && o.CounterBits == 0 {
		return fmt.Errorf("wrap requires CounterBits")
	}
	if o.MaxRate != nil && (!finite(*o.MaxRate) || *o.MaxRate <= 0) {
		return fmt.Errorf("MaxRate must be positive")
	}
	if o.Valid != nil && len(o.Valid) != n {
		return fmt.Errorf("counter validity mask length mismatch")
	}
	return nil
}
func checkCounterTimes(ts []float64, n int) error {
	if len(ts) != n {
		return fmt.Errorf("counter length mismatch")
	}
	for i, t := range ts {
		if !validEpoch(t) || i > 0 && t <= ts[i-1] {
			return fmt.Errorf("counter timestamps must be finite and strictly increasing")
		}
	}
	return nil
}

// CounterRate subtracts uint64 counters BEFORE conversion to floating point.
// This preserves tiny deltas in counters greater than 2^53.
func CounterRate(ts []float64, counters []uint64, o CounterOptions) (Samples, error) {
	n := len(counters)
	if e := checkCounterTimes(ts, n); e != nil {
		return Samples{}, e
	}
	if e := o.validate(n, 64); e != nil {
		return Samples{}, e
	}
	out := Samples{clone(ts), make([]float64, n)}
	for i, c := range counters {
		out.Values[i] = missing()
		if o.Valid != nil && !o.Valid[i] {
			continue
		}
		if o.CounterBits > 0 && o.CounterBits < 64 && c >= uint64(1)<<o.CounterBits {
			return Samples{}, fmt.Errorf("counter %d outside unsigned bit range", i)
		}
		if i == 0 || o.Valid != nil && !o.Valid[i-1] {
			continue
		}
		prev := counters[i-1]
		var delta uint64
		if c >= prev {
			delta = c - prev
		} else {
			if o.OnDecrease == "gap" {
				continue
			}
			if o.CounterBits == 64 {
				delta = (^uint64(0) - prev) + 1 + c
			} else {
				delta = (uint64(1) << o.CounterBits) - prev + c
			}
		}
		rate := float64(delta) / (ts[i] - ts[i-1]) * o.Factor
		if finite(rate) && (o.MaxRate == nil || rate <= *o.MaxRate) {
			out.Values[i] = rate
		}
	}
	return out, nil
}

// CounterRateBig supports exact unsigned counters up to 128 bits. nil is missing.
// Arguments are read-only; no caller-owned big.Int is mutated.
func CounterRateBig(ts []float64, counters []*big.Int, o CounterOptions) (Samples, error) {
	n := len(counters)
	if e := checkCounterTimes(ts, n); e != nil {
		return Samples{}, e
	}
	if e := o.validate(n, 128); e != nil {
		return Samples{}, e
	}
	out := Samples{clone(ts), make([]float64, n)}
	present := func(i int) bool { return counters[i] != nil && (o.Valid == nil || o.Valid[i]) }
	for i, c := range counters {
		out.Values[i] = missing()
		if !present(i) {
			continue
		}
		if c.Sign() < 0 || c.BitLen() > 128 || o.CounterBits > 0 && c.BitLen() > o.CounterBits {
			return Samples{}, fmt.Errorf("counter %d outside unsigned bit range", i)
		}
		if i == 0 || !present(i-1) {
			continue
		}
		delta := new(big.Int).Sub(c, counters[i-1])
		if delta.Sign() < 0 {
			if o.OnDecrease == "gap" {
				continue
			}
			delta.Add(delta, new(big.Int).Lsh(big.NewInt(1), uint(o.CounterBits)))
		}
		df, _ := new(big.Float).SetInt(delta).Float64()
		rate := df / (ts[i] - ts[i-1]) * o.Factor
		if finite(rate) && (o.MaxRate == nil || rate <= *o.MaxRate) {
			out.Values[i] = rate
		}
	}
	return out, nil
}

// CounterRateFloat is for nonintegral counters only. Large integral counters
// should use CounterRate/CounterRateBig to avoid pre-subtraction precision loss.
func CounterRateFloat(ts, counters []float64, o CounterOptions) (Samples, error) {
	if e := validateSamples(ts, counters); e != nil {
		return Samples{}, e
	}
	if e := o.validate(len(counters), 0); e != nil {
		return Samples{}, e
	}
	if o.OnDecrease != "gap" {
		return Samples{}, fmt.Errorf("float counter rollover is unsupported")
	}
	out := Samples{clone(ts), make([]float64, len(counters))}
	for i, c := range counters {
		out.Values[i] = missing()
		if c < 0 {
			return Samples{}, fmt.Errorf("negative counter")
		}
		if i == 0 || math.IsNaN(c) || math.IsNaN(counters[i-1]) || o.Valid != nil && (!o.Valid[i] || !o.Valid[i-1]) {
			continue
		}
		delta := c - counters[i-1]
		if delta < 0 {
			continue
		}
		rate := delta / (ts[i] - ts[i-1]) * o.Factor
		if finite(rate) && (o.MaxRate == nil || rate <= *o.MaxRate) {
			out.Values[i] = rate
		}
	}
	return out, nil
}
