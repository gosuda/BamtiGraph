// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"reflect"
	"testing"
)

func TestStatisticsMissingLast(t *testing.T) {
	s := mustSeries(t, []float64{0, 1, 2, 3}, []float64{10, 20, 30, math.NaN()})
	v := statsFor(s, 1, 3)
	if v.Current != nil || v.Count != 2 || v.Missing != 1 || *v.Average != 25 || *v.Maximum != 30 || *v.Minimum != 20 {
		t.Fatal(v)
	}
}
func TestStatisticsViewport(t *testing.T) {
	s := mustSeries(t, []float64{0, 1, 2}, []float64{1, 100, 3})
	v := statsFor(s, .25, 1.75)
	if v.Count != 1 || *v.Current != 100 || *v.Average != 100 {
		t.Fatal(v)
	}
	v = statsFor(s, 4, 5)
	if v.Count != 0 || v.Average != nil {
		t.Fatal(v)
	}
}
func TestStableMean(t *testing.T) {
	s := mustSeries(t, []float64{0, 1, 2}, []float64{math.MaxFloat64, math.MaxFloat64, math.MaxFloat64})
	v := statsFor(s, 0, 2)
	if v.Average == nil || !finite(*v.Average) || math.Abs(*v.Average/math.MaxFloat64-1) > 1e-15 {
		t.Fatal(v.Average)
	}
}
func TestVisibleLinearBoundaries(t *testing.T) {
	s := mustSeries(t, []float64{0, 10}, []float64{0, 100})
	r := visibleRuns(s, 2, 8)
	want := [][]point{{{2, 20}, {8, 80}}}
	if !reflect.DeepEqual(r, want) {
		t.Fatal(r)
	}
}
func TestVisibleStepPost(t *testing.T) {
	s := mustSeries(t, []float64{0, 10, 20}, []float64{1, 2, 3})
	s.Interpolation = StepPost
	r := visibleRuns(s, 5, 15)
	want := [][]point{{{5, 1}, {10, 1}, {10, 2}, {15, 2}}}
	if !reflect.DeepEqual(r, want) {
		t.Fatal(r)
	}
}
func TestGapSeparation(t *testing.T) {
	s := mustSeries(t, []float64{0, 1, 2, 20, 21, 22}, []float64{1, math.NaN(), 2, 3, 4, 5})
	s.GapAfter = 5
	r := visibleRuns(s, 0, 22)
	if len(r) != 3 || len(r[0]) != 1 || len(r[1]) != 1 || len(r[2]) != 3 {
		t.Fatal(r)
	}
}
func TestNoExtrapolation(t *testing.T) {
	s := mustSeries(t, []float64{10, 20}, []float64{1, 2})
	if r := visibleRuns(s, 0, 5); len(r) != 0 {
		t.Fatal(r)
	}
	r := visibleRuns(s, 0, 30)
	if r[0][0].x != 10 || r[0][1].x != 20 {
		t.Fatal(r)
	}
}
func TestDecimationPreservesExtrema(t *testing.T) {
	ps := make([]point, 10000)
	for i := range ps {
		ps[i] = point{float64(i), 1}
	}
	ps[532].y = 999
	ps[538].y = -999
	out := decimate(ps, 0, 9999, 100)
	if len(out) > 400 {
		t.Fatal(len(out))
	}
	mi, ma := math.Inf(1), math.Inf(-1)
	for i, p := range out {
		mi = math.Min(mi, p.y)
		ma = math.Max(ma, p.y)
		if i > 0 && p.x < out[i-1].x {
			t.Fatal("reordered")
		}
	}
	if mi != -999 || ma != 999 || out[0] != ps[0] || out[len(out)-1] != ps[len(ps)-1] {
		t.Fatal(mi, ma)
	}
}
func TestLineClip(t *testing.T) {
	for _, c := range []struct {
		name string
		a, b point
		ok   bool
	}{{"cross", point{-5, 5}, point{15, 5}, true}, {"above", point{-5, -1}, point{15, -1}, false}, {"corner", point{0, 0}, point{0, 0}, true}, {"diag", point{-5, -5}, point{15, 15}, true}} {
		t.Run(c.name, func(t *testing.T) {
			a, b, ok := clipLine(c.a, c.b, 10, 10)
			if ok != c.ok {
				t.Fatal(ok)
			}
			if ok && (a.x < 0 || a.y < 0 || b.x > 10 || b.y > 10) {
				t.Fatal(a, b)
			}
		})
	}
}
func TestPolygonClip(t *testing.T) {
	ps := clipPolygon([]point{{-100, -100}, {100, -100}, {100, 100}, {-100, 100}}, 10, 20)
	if len(ps) != 4 {
		t.Fatal(ps)
	}
	for _, p := range ps {
		if p.x < 0 || p.y < 0 || p.x > 10 || p.y > 20 {
			t.Fatal(ps)
		}
	}
}
