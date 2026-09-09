// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"reflect"
	"testing"
	"time"
)

func mustSeries(t *testing.T, ts, vs []float64) Series {
	t.Helper()
	s, e := NewSeries("sample", ts, vs)
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestSeriesValidation(t *testing.T) {
	for _, c := range []struct {
		name   string
		ts, vs []float64
		bad    bool
	}{
		{"valid", []float64{0, 1}, []float64{3, 4}, false}, {"empty", nil, nil, false}, {"missing", []float64{0, 1}, []float64{math.NaN(), 3}, false},
		{"length", []float64{0}, nil, true}, {"duplicate", []float64{0, 0}, []float64{1, 2}, true}, {"reverse", []float64{1, 0}, []float64{1, 2}, true},
		{"timestamp_inf", []float64{math.Inf(1)}, []float64{1}, true}, {"timestamp_nan", []float64{math.NaN()}, []float64{1}, true}, {"value_inf", []float64{0}, []float64{math.Inf(-1)}, true}, {"extreme_date", []float64{1e16}, []float64{1}, true},
	} {
		t.Run(c.name, func(t *testing.T) {
			_, e := NewSeries("x", c.ts, c.vs)
			if (e != nil) != c.bad {
				t.Fatalf("error=%v want bad=%v", e, c.bad)
			}
		})
	}
}
func TestConstructorsCopyData(t *testing.T) {
	ts, vs := []float64{0, 1}, []float64{2, 3}
	s, e := NewSeries("x", ts, vs)
	if e != nil {
		t.Fatal(e)
	}
	ts[0] = 99
	vs[0] = 99
	if s.Timestamps[0] != 0 || s.Values[0] != 2 {
		t.Fatal("NewSeries retained input slices")
	}
	g := NewGraph(s)
	s.Values[0] = 99
	if g.Series[0].Values[0] != 2 {
		t.Fatal("NewGraph retained slice")
	}
}
func TestRegularSeries(t *testing.T) {
	s, e := RegularSeries("x", []float64{1, 2, 3}, 100, .5)
	if e != nil || !reflect.DeepEqual(s.Timestamps, []float64{100, 100.5, 101}) {
		t.Fatal(s, e)
	}
	if _, e = RegularSeries("x", nil, 0, 0); e == nil {
		t.Fatal("accepted zero cadence")
	}
}
func TestEpochRoundtrip(t *testing.T) {
	for _, x := range []float64{0, -.5, 1.25, 1770000000.5} {
		if got := Epoch(fromEpoch(x)); math.Abs(got-x) > 1e-6 {
			t.Fatalf("%v -> %v", x, got)
		}
	}
	t0 := time.Date(2026, 1, 1, 0, 0, 0, 0, time.FixedZone("K", 32400))
	if Epochs([]time.Time{t0})[0] != Epoch(t0) {
		t.Fatal("Epochs")
	}
}
func TestConfigurationValidation(t *testing.T) {
	checks := []struct {
		name   string
		change func(*Graph)
	}{
		{"kind", func(g *Graph) { g.Series[0].Kind = "bars" }}, {"interp", func(g *Graph) { g.Series[0].Interpolation = "spline" }}, {"line_width", func(g *Graph) { g.Series[0].LineWidth = 0 }}, {"gap", func(g *Graph) { g.Series[0].GapAfter = -1 }},
		{"mode", func(g *Graph) { g.TimeAxis.Mode = "bad" }}, {"zone", func(g *Graph) { g.TimeAxis.Timezone = "Unknown/Zone" }}, {"time_range", func(g *Graph) { g.TimeAxis.Start = Float(3); g.TimeAxis.End = Float(2) }}, {"zero_tick", func(g *Graph) { g.TimeAxis.MajorSeconds = Float(0) }},
		{"y_range", func(g *Graph) { g.YAxis.Maximum = Float(0) }}, {"base", func(g *Graph) { g.YAxis.Base = 2 }}, {"zero_step", func(g *Graph) { g.YAxis.MajorStep = Float(0) }}, {"minor_divisions", func(g *Graph) { g.YAxis.MinorDivisions = 0 }}, {"decimals", func(g *Graph) { g.YAxis.LegendDecimals = 13 }},
		{"width", func(g *Graph) { g.Layout.Width = 200 }}, {"aa", func(g *Graph) { g.Layout.Antialias = 0 }}, {"scale", func(g *Graph) { g.Layout.PixelScale = 9 }}, {"margins", func(g *Graph) { g.Layout.Left = 500 }}, {"legend", func(g *Graph) { g.Layout.Legend = "wat" }}, {"columns", func(g *Graph) { g.Layout.LegendLayout.Compact[0][0] = 0 }},
		{"theme_font", func(g *Graph) { g.Theme.TitleSize = math.NaN() }}, {"dash", func(g *Graph) { g.Theme.GridDash = [2]int{0, 1} }}, {"strict_font", func(g *Graph) { g.Fonts.Strict = true }}, {"font_index", func(g *Graph) { g.Fonts.FaceIndex = -1 }},
		{"text", func(g *Graph) { g.Title = "two\nlines" }}, {"override", func(g *Graph) { g.Series[0].LegendValues = &LegendValues{Current: math.Inf(1)} }}, {"hrule", func(g *Graph) { g.HRules = []HRule{{Value: 1, Width: 0}} }}, {"vrule", func(g *Graph) { r := VerticalRule(2); r.Dash = &[2]int{1, 0}; g.VRules = []VRule{r} }},
		{"allocation", func(g *Graph) { g.Layout.Width = 8192; g.Layout.PlotHeight = 4096; g.Layout.Antialias = 8 }},
	}
	for _, c := range checks {
		t.Run(c.name, func(t *testing.T) {
			g := NewGraph(mustSeries(t, []float64{0, 1}, []float64{1, 2}))
			c.change(g)
			if e := g.Validate(); e == nil {
				t.Fatal("invalid configuration accepted")
			}
		})
	}
}
func TestColorParsing(t *testing.T) {
	for _, s := range []string{"#123", "#010203", "#01020304", "black", "transparent"} {
		if _, e := ParseColor(s); e != nil {
			t.Fatal(s, e)
		}
	}
	for _, s := range []string{"#xx0000", "#12345", "#", "unknown", "fff"} {
		if _, e := ParseColor(s); e == nil {
			t.Fatal("accepted", s)
		}
	}
	c, _ := ParseColor("#123")
	if c != RGB(17, 34, 51) {
		t.Fatal(c)
	}
}
func TestReferenceDimensions(t *testing.T) {
	l := ReferenceLayout()
	if l.PlotBox() != [4]int{64, 34, 564, 156} || l.Height(2) != 211 {
		t.Fatal(l)
	}
	l.Legend = "none"
	if l.Height(128) != 174 {
		t.Fatal(l.Height(128))
	}
}
