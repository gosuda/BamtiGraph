// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"reflect"
	"testing"
	"time"
	_ "time/tzdata"
)

func TestSIAndIEC(t *testing.T) {
	for _, c := range []struct {
		v      float64
		base   int
		factor float64
		suffix string
	}{{2e8, 1000, 1e6, "M"}, {.001, 1000, .001, "m"}, {0, 1000, 1, ""}, {1 << 30, 1024, 1 << 30, "Gi"}, {.1, 1024, 1, ""}} {
		f, s := unitFor(c.v, c.base)
		if f != c.factor || s != c.suffix {
			t.Fatal(c, f, s)
		}
	}
}
func TestYAutoscale(t *testing.T) {
	a := DefaultYAxis()
	s, e := resolveY(a, []float64{100e6, 232e6})
	if e != nil || s.factor != 1e6 || s.maximum < 232e6 || s.step != 50e6 {
		t.Fatal(s, e)
	}
	a.Minimum = nil
	s, e = resolveY(a, []float64{-2, 5})
	if e != nil || s.minimum >= -2 || s.maximum <= 5 {
		t.Fatal(s, e)
	}
}
func TestYNegativeFixed(t *testing.T) {
	a := DefaultYAxis()
	a.Minimum = Float(-10)
	a.Maximum = Float(10)
	a.MajorStep = Float(5)
	s, e := resolveY(a, nil)
	if e != nil || !reflect.DeepEqual(s.major, []float64{-10, -5, 0, 5, 10}) {
		t.Fatal(s, e)
	}
}
func TestYCustomScale(t *testing.T) {
	a := DefaultYAxis()
	a.ScaleFactor = Float(1)
	a.Suffix = String("%")
	a.Maximum = Float(100)
	s, e := resolveY(a, []float64{20})
	if e != nil || s.label(20, false) != "20 %" || s.label(0, false) != "0" || s.label(0, true) != "0 %" {
		t.Fatal(s, e)
	}
}
func TestTickCountGuard(t *testing.T) {
	if _, e := multiples(0, 100000, 1); e == nil {
		t.Fatal("no guard")
	}
	if _, e := wallTicks(0, 100000, 1, time.UTC); e == nil {
		t.Fatal("no time guard")
	}
	if _, e := niceStep(math.SmallestNonzeroFloat64); e == nil {
		t.Fatal("underflow not rejected")
	}
}
func TestLocalWallClockAlignment(t *testing.T) {
	loc, e := time.LoadLocation("Asia/Kathmandu")
	if e != nil {
		t.Fatal(e)
	}
	start := Epoch(time.Date(2026, 1, 1, 0, 0, 0, 0, loc))
	ticks, e := wallTicks(start, start+4*3600, 3600, loc)
	if e != nil || len(ticks) != 5 {
		t.Fatal(ticks, e)
	}
	for _, v := range ticks {
		tm := fromEpoch(v).In(loc)
		if tm.Minute() != 0 {
			t.Fatal(tm)
		}
	}
}
func TestDSTFoldAndGap(t *testing.T) {
	loc, e := time.LoadLocation("America/New_York")
	if e != nil {
		t.Fatal(e)
	}
	for _, c := range []struct {
		name       string
		month      time.Month
		day        int
		ones, twos int
	}{{"spring", 3, 8, 1, 0}, {"fall", 11, 1, 2, 1}} {
		t.Run(c.name, func(t *testing.T) {
			start := Epoch(time.Date(2026, c.month, c.day, 0, 0, 0, 0, loc))
			ticks, e := wallTicks(start, start+5*3600, 3600, loc)
			if e != nil {
				t.Fatal(e)
			}
			ones, twos := 0, 0
			for _, v := range ticks {
				h := fromEpoch(v).In(loc).Hour()
				if h == 1 {
					ones++
				}
				if h == 2 {
					twos++
				}
			}
			if ones != c.ones || twos != c.twos {
				t.Fatal(ones, twos, ticks)
			}
		})
	}
}
func TestFallLabelsDisambiguate(t *testing.T) {
	a := DailyAxis("America/New_York")
	loc, _ := time.LoadLocation(a.Timezone)
	start := Epoch(time.Date(2026, 11, 1, 0, 0, 0, 0, loc))
	a.MinorSeconds = Float(3600)
	a.MajorSeconds = Float(3600)
	a.LabelSeconds = Float(3600)
	s, e := resolveX(a, start, start+5*3600, 500)
	if e != nil {
		t.Fatal(e)
	}
	found := map[string]bool{}
	for _, l := range s.labels {
		found[l.Label] = true
	}
	if !found["01:00 -0400"] || !found["01:00 -0500"] {
		t.Fatal(s.labels)
	}
}
func TestWeeklyNoonDST(t *testing.T) {
	a := WeeklyAxis("America/New_York")
	loc, _ := time.LoadLocation(a.Timezone)
	start := Epoch(time.Date(2026, 3, 7, 0, 0, 0, 0, loc))
	end := Epoch(time.Date(2026, 3, 10, 0, 0, 0, 0, loc))
	s, e := resolveX(a, start, end, 500)
	if e != nil || len(s.labels) != 3 {
		t.Fatal(s.labels, e)
	}
	for _, l := range s.labels {
		if fromEpoch(l.Time).In(loc).Hour() != 12 {
			t.Fatal(l)
		}
	}
}
func TestCalendarMonthTicks(t *testing.T) {
	start := Epoch(time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC))
	end := Epoch(time.Date(2024, 4, 1, 0, 0, 0, 0, time.UTC))
	xs, e := monthTicks(start, end, time.UTC, 1)
	if e != nil || len(xs) != 4 {
		t.Fatal(xs, e)
	}
	if xs[2]-xs[1] != 29*86400 {
		t.Fatal("February ignored", xs)
	}
}
func TestEmptyExplicitTicks(t *testing.T) {
	a := DefaultTimeAxis()
	a.Mode = "yearly"
	a.Ticks = []Tick{}
	a.MajorTicks = []float64{}
	a.MinorTicks = []float64{}
	s, e := resolveX(a, 0, 365*86400, 500)
	if e != nil || len(s.labels) != 0 || len(s.major) != 0 || len(s.minor) != 0 {
		t.Fatal(s, e)
	}
}
func TestTimeFormatting(t *testing.T) {
	v := Epoch(time.Date(2026, 9, 8, 13, 4, 5, 0, time.UTC))
	s, e := formatTime(v, time.UTC, "%a %b %d %Y %H:%M:%S %z %%")
	if e != nil || s != "Tue Sep 08 2026 13:04:05 +0000 %" {
		t.Fatal(s, e)
	}
	if _, e = formatTime(v, time.UTC, "%Q"); e == nil {
		t.Fatal("unknown directive")
	}
	if _, e = formatTime(v, time.UTC, "abc%"); e == nil {
		t.Fatal("trailing %")
	}
}
