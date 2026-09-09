// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph_test

import (
	rrd "github.com/gosuda/BamtiGraph"
	"github.com/gosuda/BamtiGraph/internal/fixture"
	"math"
	"testing"
	_ "time/tzdata"
)

func TestPythonBehaviorParity(t *testing.T) {
	cases, err := fixture.Load("testdata/python_cases.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, c := range cases {
		t.Run(c.Name, func(t *testing.T) {
			g, err := c.Graph()
			if err != nil {
				t.Fatal(err)
			}
			got, err := g.RenderResult()
			if err != nil {
				t.Fatal(err)
			}
			m, w := got.Metadata, c.Expected
			if m.ImageSize != w.ImageSize || m.LogicalSize != w.LogicalSize || m.PlotBox != w.PlotBox {
				t.Fatalf("geometry differs: %+v", m)
			}
			near := func(label string, a, b float64) {
				t.Helper()
				if math.Abs(a-b) > 1e-11*math.Max(1, math.Max(math.Abs(a), math.Abs(b))) {
					t.Errorf("%s: Go %.17g Python %.17g", label, a, b)
				}
			}
			for i := 0; i < 2; i++ {
				near("time_range", m.TimeRange[i], w.TimeRange[i])
				near("y_range", m.YRange[i], w.YRange[i])
			}
			near("y_step", m.YStep, w.YStep)
			near("unit factor", m.YUnit.Factor, w.YUnit.Factor)
			if m.YUnit.Suffix != w.YUnit.Suffix {
				t.Errorf("unit suffix %q != %q", m.YUnit.Suffix, w.YUnit.Suffix)
			}
			if len(m.XLabels) != len(w.XLabels) {
				t.Fatalf("x label counts: Go %d; Python %d", len(m.XLabels), len(w.XLabels))
			}
			for i, tick := range m.XLabels {
				want := w.XLabels[i]
				if tick.Label != want.Label || tick.Time != want.Time {
					t.Errorf("x label %d: Go %+v; Python %+v", i, tick, want)
				}
				near("x label position", tick.X, want.X)
			}
			if len(m.Statistics) != len(w.Statistics) {
				t.Fatal("statistic counts differ")
			}
			for i, a := range m.Statistics {
				b := w.Statistics[i]
				if a.Name != b.Name || a.Count != b.Count || a.Missing != b.Missing {
					t.Errorf("statistic sample policy: %+v != %+v", a, b)
				}
				for j, p := range []*float64{a.Current, a.Average, a.Maximum, a.Minimum} {
					q := []*float64{b.Current, b.Average, b.Maximum, b.Minimum}[j]
					if (p == nil) != (q == nil) {
						t.Errorf("missing statistic %d/%d", i, j)
					} else if p != nil {
						near("sample statistics", *p, *q)
					}
				}
			}
		})
	}
}

func BenchmarkRenderTraffic(b *testing.B) {
	g, e := rrd.DemoTraffic(false)
	if e != nil {
		b.Fatal(e)
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e = g.Render(); e != nil {
			b.Fatal(e)
		}
	}
}
func BenchmarkRender100kSamples(b *testing.B) {
	const n = 100000
	ts, vs := make([]float64, n), make([]float64, n)
	for i := range ts {
		ts[i] = float64(i)
		vs[i] = 100 + 50*math.Sin(float64(i)/100)
	}
	s, e := rrd.NewSeries("Load", ts, vs)
	if e != nil {
		b.Fatal(e)
	}
	g := rrd.NewGraph(s)
	g.VerticalLabel = "requests per second"
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, e = g.Render(); e != nil {
			b.Fatal(e)
		}
	}
}
