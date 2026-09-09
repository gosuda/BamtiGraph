// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"math/big"
	"sort"
)

type point struct{ x, y float64 }

// Statistics describe the original samples in the inclusive viewport. Nil
// numbers marshal as JSON null; Current is nil when the LAST sample is missing.
type Statistics struct {
	Current *float64 `json:"current"`
	Average *float64 `json:"average"`
	Maximum *float64 `json:"maximum"`
	Minimum *float64 `json:"minimum"`
	Count   int      `json:"count"`
	Missing int      `json:"missing"`
}

func nullable(v float64) *float64 {
	if math.IsNaN(v) {
		return nil
	}
	return Float(v)
}
func lowerBound(a []float64, v float64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] >= v })
}
func upperBound(a []float64, v float64) int {
	return sort.Search(len(a), func(i int) bool { return a[i] > v })
}

// Neumaier compensation after division handles ordinary means cheaply.
// Rare overflowing intermediates use wide exponent arithmetic; a finite mean
// of finite samples must stay inside their convex hull.
func compensatedSum(vs []float64, divisor float64) float64 {
	var sum, c float64
	for _, v := range vs {
		if math.IsNaN(v) {
			continue
		}
		v /= divisor
		t := sum + v
		if math.Abs(sum) >= math.Abs(v) {
			c += (sum - t) + v
		} else {
			c += (v - t) + sum
		}
		sum = t
	}
	result := sum + c
	if !finite(result) && divisor != 1 {
		acc := new(big.Float).SetPrec(256)
		term := new(big.Float).SetPrec(256)
		for _, v := range vs {
			if !math.IsNaN(v) {
				acc.Add(acc, term.SetFloat64(v))
			}
		}
		acc.Quo(acc, term.SetFloat64(divisor))
		result, _ = acc.Float64()
	}
	return result
}
func statsFor(s Series, start, end float64) Statistics {
	a, b := lowerBound(s.Timestamps, start), upperBound(s.Timestamps, end)
	v := s.Values[a:b]
	st := Statistics{}
	if len(v) > 0 {
		st.Current = nullable(v[len(v)-1])
	}
	lo, hi := math.Inf(1), math.Inf(-1)
	for _, x := range v {
		if math.IsNaN(x) {
			st.Missing++
			continue
		}
		st.Count++
		lo = math.Min(lo, x)
		hi = math.Max(hi, x)
	}
	if st.Count > 0 {
		st.Minimum = Float(lo)
		st.Maximum = Float(hi)
		st.Average = Float(compensatedSum(v, float64(st.Count)))
	}
	return st
}

// visibleRuns keeps a neighbor at each viewport boundary for interpolation, but
// it never bridges explicit NaN or GapAfter gaps and never extrapolates.
func visibleRuns(s Series, start, end float64) [][]point {
	a := max(0, lowerBound(s.Timestamps, start)-1)
	b := min(len(s.Timestamps), upperBound(s.Timestamps, end)+1)
	var raw, out [][]point
	var run []point
	for i := a; i < b; i++ {
		t, v := s.Timestamps[i], s.Values[i]
		gap := i > a && s.GapAfter > 0 && t-s.Timestamps[i-1] > s.GapAfter
		if math.IsNaN(v) || gap {
			if len(run) > 0 {
				raw = append(raw, run)
				run = nil
			}
		}
		if !math.IsNaN(v) {
			run = append(run, point{t, v})
		}
	}
	if len(run) > 0 {
		raw = append(raw, run)
	}
	for _, r := range raw {
		if r[len(r)-1].x < start || r[0].x > end {
			continue
		}
		if len(r) == 1 {
			if r[0].x >= start && r[0].x <= end {
				out = append(out, r)
			}
			continue
		}
		var clipped []point
		push := func(p point) {
			if len(clipped) == 0 || clipped[len(clipped)-1] != p {
				clipped = append(clipped, p)
			}
		}
		for i := 1; i < len(r); i++ {
			p, q := r[i-1], r[i]
			if q.x < start || p.x > end {
				continue
			}
			l, h := math.Max(p.x, start), math.Min(q.x, end)
			if l > h {
				continue
			}
			if s.Interpolation == Linear {
				al, ar := (l-p.x)/(q.x-p.x), (h-p.x)/(q.x-p.x)
				push(point{l, p.y*(1-al) + q.y*al})
				push(point{h, p.y*(1-ar) + q.y*ar})
			} else {
				push(point{l, p.y})
				push(point{h, p.y})
				if h == q.x {
					push(point{h, q.y})
				}
			}
		}
		if len(clipped) > 0 {
			out = append(out, clipped)
		}
	}
	return out
}
func decimate(ps []point, start, end float64, pixels int) []point {
	if len(ps) <= pixels*4 {
		return ps
	}
	out := make([]point, 0, pixels*4)
	for a := 0; a < len(ps); {
		key := min(pixels-1, max(0, int((ps[a].x-start)/(end-start)*float64(pixels))))
		b := a + 1
		mi, ma := a, a
		for b < len(ps) {
			k := min(pixels-1, max(0, int((ps[b].x-start)/(end-start)*float64(pixels))))
			if k != key {
				break
			}
			if ps[b].y < ps[mi].y {
				mi = b
			}
			if ps[b].y > ps[ma].y {
				ma = b
			}
			b++
		}
		idx := []int{a, mi, ma, b - 1}
		sort.Ints(idx)
		prev := -1
		for _, i := range idx {
			if i != prev {
				out = append(out, ps[i])
				prev = i
			}
		}
		a = b
	}
	return out
}
func clipPolygon(ps []point, w, h float64) []point {
	for _, edge := range []struct {
		dim     int
		b, sign float64
	}{{0, 0, 1}, {0, w, -1}, {1, 0, 1}, {1, h, -1}} {
		if len(ps) == 0 {
			break
		}
		var out []point
		coord := func(p point) float64 {
			if edge.dim == 0 {
				return p.x
			}
			return p.y
		}
		prev := ps[len(ps)-1]
		pin := (coord(prev)-edge.b)*edge.sign >= 0
		for _, cur := range ps {
			cin := (coord(cur)-edge.b)*edge.sign >= 0
			if pin != cin {
				f := (edge.b - coord(prev)) / (coord(cur) - coord(prev))
				cross := point{prev.x*(1-f) + cur.x*f, prev.y*(1-f) + cur.y*f}
				if edge.dim == 0 {
					cross.x = edge.b
				} else {
					cross.y = edge.b
				}
				out = append(out, cross)
			}
			if cin {
				out = append(out, cur)
			}
			prev, pin = cur, cin
		}
		ps = out
	}
	return ps
}
func clipLine(a, b point, w, h float64) (point, point, bool) {
	dx, dy := b.x-a.x, b.y-a.y
	u, v := 0., 1.
	for _, pq := range [][2]float64{{-dx, a.x}, {dx, w - a.x}, {-dy, a.y}, {dy, h - a.y}} {
		p, q := pq[0], pq[1]
		if p == 0 {
			if q < 0 {
				return point{}, point{}, false
			}
		} else {
			r := q / p
			if p < 0 {
				u = math.Max(u, r)
			} else {
				v = math.Min(v, r)
			}
			if u > v {
				return point{}, point{}, false
			}
		}
	}
	return point{a.x + u*dx, a.y + u*dy}, point{a.x + v*dx, a.y + v*dy}, true
}
