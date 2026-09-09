// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"
)

type yState struct {
	minimum, maximum, step, factor float64
	suffix                         string
	decimals                       int
	major, minor                   []float64
}

func niceStep(v float64) (float64, error) {
	if !finite(v) || v <= 0 {
		return 0, fmt.Errorf("axis span must be positive and finite")
	}
	if v < 1e-300 {
		return 0, fmt.Errorf("axis magnitude too small")
	}
	p := math.Pow(10, math.Floor(math.Log10(v)))
	if p == 0 {
		return 0, fmt.Errorf("axis magnitude too small")
	}
	for _, m := range []float64{1, 2, 5, 10} {
		if v <= m*p*(1+1e-12) {
			return m * p, nil
		}
	}
	return 10 * p, nil
}
func multiples(lo, hi, step float64) ([]float64, error) {
	if !finite(step) || step <= 0 || !finite((hi-lo)/step) || (hi-lo)/step > MaxTicks {
		return nil, fmt.Errorf("too many axis ticks or invalid step")
	}
	a, b := math.Ceil(lo/step-1e-10), math.Floor(hi/step+1e-10)
	if !finite(a) || !finite(b) || b-a > MaxTicks || math.Abs(a) > 9e15 || math.Abs(b) > 9e15 {
		return nil, fmt.Errorf("axis tick indices exceed precision guard")
	}
	out := make([]float64, 0, max(0, int(b-a+1)))
	for i := 0; i <= int(b-a); i++ {
		v := (a + float64(i)) * step
		if v == 0 {
			v = 0
		}
		out = append(out, v)
	}
	return out, nil
}
func unitFor(v float64, base int) (float64, string) {
	if v == 0 {
		return 1, ""
	}
	i := int(math.Floor(math.Log(math.Abs(v))/math.Log(float64(base)) + 1e-12))
	if base == 1024 {
		i = max(0, min(8, i))
		return math.Pow(1024, float64(i)), []string{"", "Ki", "Mi", "Gi", "Ti", "Pi", "Ei", "Zi", "Yi"}[i]
	}
	i = max(-8, min(8, i))
	return math.Pow(1000, float64(i)), []string{"y", "z", "a", "f", "p", "n", "u", "m", "", "k", "M", "G", "T", "P", "E", "Z", "Y"}[i+8]
}
func resolveY(a YAxis, values []float64) (yState, error) {
	var st yState
	ld, ud := 0., 0.
	if len(values) > 0 {
		ld, ud = values[0], values[0]
	}
	for _, v := range values {
		if !finite(v) {
			return st, fmt.Errorf("nonfinite projected data")
		}
		ld = math.Min(ld, v)
		ud = math.Max(ud, v)
	}
	lo, hi := math.Min(0, ld), math.Max(0, ud)
	if a.Minimum != nil {
		lo = *a.Minimum
	}
	if a.Maximum != nil {
		hi = *a.Maximum
	}
	if a.Maximum == nil && hi <= lo {
		hi = lo + math.Max(math.Abs(lo)*.05, 1)
	}
	if a.Minimum == nil && lo >= hi {
		lo = hi - math.Max(math.Abs(hi)*.05, 1)
	}
	span := hi - lo
	if !finite(span) || span <= 0 {
		return st, fmt.Errorf("invalid y span")
	}
	af, _ := unitFor(math.Max(math.Abs(lo), math.Abs(hi)), a.Base)
	if a.ScaleFactor != nil {
		af = *a.ScaleFactor
	}
	step := 0.
	if a.MajorStep != nil {
		step = *a.MajorStep
	} else {
		n, e := niceStep(span / af / 5)
		if e != nil {
			return st, e
		}
		step = n * af
	}
	q := step / float64(a.MinorDivisions)
	if q <= 0 || !finite(q) {
		return st, fmt.Errorf("invalid y-axis quantum")
	}
	if a.Minimum == nil && lo < 0 {
		lo = math.Floor((lo-span*.02)/q) * q
	}
	if a.Maximum == nil {
		hi = math.Ceil((hi+span*.02)/q) * q
	}
	if !finite(hi-lo) || hi <= lo {
		return st, fmt.Errorf("degenerate resolved y-axis")
	}
	maj, e := multiples(lo, hi, step)
	if e != nil {
		return st, e
	}
	all, e := multiples(lo, hi, q)
	if e != nil {
		return st, e
	}
	var minor []float64
	for _, v := range all {
		if math.Abs(v/step-math.RoundToEven(v/step)) > 1e-8 {
			minor = append(minor, v)
		}
	}
	factor, suffix := unitFor(math.Max(math.Abs(lo), math.Abs(hi)), a.Base)
	if a.ScaleFactor != nil {
		factor = *a.ScaleFactor
		if a.Suffix == nil {
			suffix = ""
		}
	}
	if a.Suffix != nil {
		suffix = *a.Suffix
	}
	decimals := 9
	if a.Decimals != nil {
		decimals = *a.Decimals
	} else {
		scaled := step / factor
		for d := 0; d < 10; d++ {
			p := math.Pow10(d)
			r := math.RoundToEven(scaled*p) / p
			if math.Abs(scaled-r) <= math.Max(1e-10, math.Abs(scaled)*1e-9) {
				decimals = d
				break
			}
		}
	}
	return yState{lo, hi, step, factor, suffix, decimals, maj, minor}, nil
}
func (s yState) label(v float64, zeroSuffix bool) string {
	n := v / s.factor
	if math.Abs(n) < .5*math.Pow10(-s.decimals) {
		n = 0
	}
	suffix := s.suffix
	if v == 0 && !zeroSuffix {
		suffix = ""
	}
	result := fmt.Sprintf("%.*f", s.decimals, n)
	if suffix != "" {
		result += " " + suffix
	}
	return result
}
func legendNumber(v *float64, a YAxis, s yState) string {
	if v == nil || math.IsNaN(*v) {
		return "NaN"
	}
	r := fmt.Sprintf("%.*f", a.LegendDecimals, *v/s.factor)
	if s.suffix != "" {
		r += " " + s.suffix
	}
	return r
}

type xState struct {
	start, end   float64
	major, minor []float64
	labels       []Tick
	mode         string
}

func fromEpoch(v float64) time.Time {
	whole := math.Floor(v)
	return time.Unix(int64(whole), int64(math.Round((v-whole)*1e9))).UTC()
}
func wallAsUTC(t time.Time) time.Time {
	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), time.UTC)
}

// localCandidates round-trips candidate offsets, preserving both folds and
// rejecting nonexistent wall times. Offset transitions are walked using ZoneBounds.
func zoneOffsets(start, end float64, loc *time.Location) []int {
	begin := fromEpoch(start - 3*86400)
	stop := fromEpoch(end + 3*86400)
	seen := map[int]bool{}
	for t, steps := begin, 0; !t.After(stop) && steps < MaxTicks; t, steps = t.Add(time.Second), steps+1 {
		local := t.In(loc)
		_, off := local.Zone()
		seen[off] = true
		_, bound := local.ZoneBounds()
		if bound.IsZero() || bound.After(stop) {
			break
		}
		if bound.After(t) {
			t = bound.Add(-time.Second)
		}
	}
	out := make([]int, 0, len(seen))
	for v := range seen {
		out = append(out, v)
	}
	sort.Ints(out)
	return out
}
func localCandidates(wall time.Time, loc *time.Location, offsets []int) []float64 {
	out := []float64{}
	for _, off := range offsets {
		t := wall.Add(-time.Duration(off) * time.Second)
		if wallAsUTC(t.In(loc)).Equal(wall) {
			out = append(out, Epoch(t))
		}
	}
	sort.Float64s(out)
	return out
}
func wallTicks(start, end, seconds float64, loc *time.Location) ([]float64, error) {
	if seconds <= 0 || !finite(seconds) {
		return nil, fmt.Errorf("invalid wall tick interval")
	}
	aa, bb := Epoch(wallAsUTC(fromEpoch(start).In(loc))), Epoch(wallAsUTC(fromEpoch(end).In(loc)))
	a, b := math.Floor(math.Min(aa, bb)/seconds)-2, math.Ceil(math.Max(aa, bb)/seconds)+2
	if !finite(b-a) || b-a > MaxTicks || math.Abs(a) > 9e15 || math.Abs(b) > 9e15 {
		return nil, fmt.Errorf("too many time ticks; choose a larger interval")
	}
	offsets := zoneOffsets(start, end, loc)
	seen := map[float64]bool{}
	for i := 0; i <= int(b-a); i++ {
		wall := fromEpoch((a + float64(i)) * seconds)
		for _, v := range localCandidates(wall, loc, offsets) {
			if v >= start && v <= end {
				seen[v] = true
			}
		}
	}
	out := make([]float64, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	sort.Float64s(out)
	return out, nil
}
func monthTicks(start, end float64, loc *time.Location, stride int) ([]float64, error) {
	a, b := fromEpoch(start).In(loc), fromEpoch(end).In(loc)
	first := (a.Year()*12 + int(a.Month()) - 1) / stride * stride
	last := b.Year()*12 + int(b.Month()) - 1 + stride
	if (last-first)/stride > MaxTicks {
		return nil, fmt.Errorf("too many calendar ticks")
	}
	offs := zoneOffsets(start, end, loc)
	var out []float64
	for i := first; i <= last; i += stride {
		year, month := i/12, i%12+1
		if year < 1 || year > 9999 {
			continue
		}
		w := time.Date(year, time.Month(month), 1, 0, 0, 0, 0, time.UTC)
		for _, t := range localCandidates(w, loc, offs) {
			if t >= start && t <= end {
				out = append(out, t)
			}
		}
	}
	sort.Float64s(out)
	return out, nil
}

// formatTime implements a fixed, locale-independent subset of strftime.
func formatTime(v float64, loc *time.Location, format string) (string, error) {
	t := fromEpoch(v).In(loc)
	var b strings.Builder
	for i := 0; i < len(format); i++ {
		if format[i] != '%' {
			b.WriteByte(format[i])
			continue
		}
		i++
		if i >= len(format) {
			return "", fmt.Errorf("trailing %% in label format")
		}
		var s string
		switch format[i] {
		case '%':
			s = "%"
		case 'H':
			s = t.Format("15")
		case 'I':
			s = t.Format("03")
		case 'M':
			s = t.Format("04")
		case 'S':
			s = t.Format("05")
		case 'd':
			s = t.Format("02")
		case 'e':
			s = fmt.Sprintf("%2d", t.Day())
		case 'm':
			s = t.Format("01")
		case 'Y':
			s = t.Format("2006")
		case 'y':
			s = t.Format("06")
		case 'a':
			s = t.Format("Mon")
		case 'A':
			s = t.Format("Monday")
		case 'b', 'h':
			s = t.Format("Jan")
		case 'B':
			s = t.Format("January")
		case 'z':
			s = t.Format("-0700")
		case 'Z':
			s = t.Format("MST")
		case 'p':
			s = t.Format("PM")
		case 'j':
			s = fmt.Sprintf("%03d", t.YearDay())
		case 'w':
			s = fmt.Sprint(int(t.Weekday()))
		case 'u':
			d := int(t.Weekday())
			if d == 0 {
				d = 7
			}
			s = fmt.Sprint(d)
		case 'F':
			s = t.Format("2006-01-02")
		case 'T':
			s = t.Format("15:04:05")
		case 'R':
			s = t.Format("15:04")
		default:
			return "", fmt.Errorf("unsupported strftime directive %%%c", format[i])
		}
		b.WriteString(s)
	}
	return b.String(), nil
}
func filterTicks(ts []float64, start, end float64) []float64 {
	out := make([]float64, 0)
	for _, t := range ts {
		if t >= start && t <= end {
			out = append(out, t)
		}
	}
	sort.Float64s(out)
	return out
}
func resolveX(a TimeAxis, start, end float64, width int) (xState, error) {
	st := xState{start: start, end: end}
	loc, e := time.LoadLocation(a.Timezone)
	if e != nil {
		return st, e
	}
	span := end - start
	mode := a.Mode
	if mode == "auto" {
		switch {
		case span <= 2*86400:
			mode = "daily"
		case span <= 10*86400:
			mode = "weekly"
		case span <= 62*86400:
			mode = "monthly"
		default:
			mode = "yearly"
		}
	}
	st.mode = mode
	var labelTimes []float64
	format := "%H:%M"
	if mode == "yearly" {
		stride := 1
		if span > 550*86400 {
			stride = max(1, int(math.Ceil(span/(365.25*86400))))
			format = "%b %Y"
		} else {
			format = "%b"
		}
		if a.MajorTicks == nil || a.Ticks == nil {
			st.major, e = monthTicks(start, end, loc, stride)
			if e != nil {
				return st, e
			}
			labelTimes = st.major
		}
		if a.MinorTicks == nil {
			st.minor, e = monthTicks(start, end, loc, 1)
			if e != nil {
				return st, e
			}
		}
	} else {
		mi, ma, ls := 1800., 7200., 7200.
		switch mode {
		case "weekly":
			mi, ma, ls, format = 21600, 86400, 86400, "%d"
		case "monthly":
			mi, ma, ls, format = 86400, 604800, 604800, "%d %b"
		}
		if a.Mode == "auto" && span < 12*3600 {
			target := span / float64(max(2, width/48))
			for _, i := range []float64{1, 5, 10, 15, 30, 60, 120, 300, 600, 900, 1800, 3600, 7200} {
				ls = i
				if i >= target {
					break
				}
			}
			mi, ma = math.Max(1, ls/4), ls
			if ls < 60 {
				format = "%H:%M:%S"
			}
		}
		mi = math.Max(mi, span/2000)
		if a.MinorSeconds != nil {
			mi = *a.MinorSeconds
		}
		if a.MajorSeconds != nil {
			ma = *a.MajorSeconds
		}
		if a.LabelSeconds != nil {
			ls = *a.LabelSeconds
		}
		if a.MinorTicks == nil {
			st.minor, e = wallTicks(start, end, mi, loc)
			if e != nil {
				return st, e
			}
		}
		if a.MajorTicks == nil {
			st.major, e = wallTicks(start, end, ma, loc)
			if e != nil {
				return st, e
			}
		}
		if a.Ticks == nil {
			labelTimes, e = wallTicks(start, end, ls, loc)
			if e != nil {
				return st, e
			}
		}
	}
	if mode == "yearly" {
		if a.MinorTicks == nil && a.MinorSeconds != nil {
			st.minor, e = wallTicks(start, end, *a.MinorSeconds, loc)
			if e != nil {
				return st, e
			}
		}
		if a.MajorTicks == nil && a.MajorSeconds != nil {
			st.major, e = wallTicks(start, end, *a.MajorSeconds, loc)
			if e != nil {
				return st, e
			}
		}
		if a.Ticks == nil && a.LabelSeconds != nil {
			labelTimes, e = wallTicks(start, end, *a.LabelSeconds, loc)
			if e != nil {
				return st, e
			}
		}
	}
	if a.MajorTicks != nil {
		st.major = filterTicks(a.MajorTicks, start, end)
	}
	if a.MinorTicks != nil {
		st.minor = filterTicks(a.MinorTicks, start, end)
	}
	if a.LabelFormat != "" {
		format = a.LabelFormat
	}
	if a.Ticks != nil {
		for _, t := range a.Ticks {
			if t.Time >= start && t.Time <= end {
				st.labels = append(st.labels, t)
			}
		}
	} else if mode == "weekly" && a.LabelOffsetSeconds == 0 && a.LabelSeconds == nil {
		bounds, err := wallTicks(start-2*86400, end, 86400, loc)
		if err != nil {
			return st, err
		}
		offs := zoneOffsets(start-2*86400, end+2*86400, loc)
		for _, t := range bounds {
			mid := wallAsUTC(fromEpoch(t).In(loc))
			next := localCandidates(mid.AddDate(0, 0, 1), loc, offs)
			if t < start || len(next) == 0 || next[len(next)-1] > end {
				continue
			}
			for _, noon := range localCandidates(mid.Add(12*time.Hour), loc, offs) {
				if noon >= start && noon <= end {
					lab, err := formatTime(t, loc, format)
					if err != nil {
						return st, err
					}
					st.labels = append(st.labels, Tick{noon, lab})
				}
			}
		}
	} else {
		for _, t := range labelTimes {
			pos := t + a.LabelOffsetSeconds
			if pos >= start && pos <= end {
				lab, err := formatTime(t, loc, format)
				if err != nil {
					return st, err
				}
				st.labels = append(st.labels, Tick{pos, lab})
			}
		}
	}
	sort.SliceStable(st.labels, func(i, j int) bool { return st.labels[i].Time < st.labels[j].Time })
	if a.Ticks == nil && mode == "daily" && span <= 86400*1.1 {
		offsets := map[string]map[int]bool{}
		for _, t := range st.labels {
			_, off := fromEpoch(t.Time).In(loc).Zone()
			if offsets[t.Label] == nil {
				offsets[t.Label] = map[int]bool{}
			}
			offsets[t.Label][off] = true
		}
		for i := range st.labels {
			if len(offsets[st.labels[i].Label]) > 1 {
				off, _ := formatTime(st.labels[i].Time, loc, "%z")
				st.labels[i].Label += " " + off
			}
		}
	}
	majors := map[float64]bool{}
	for _, t := range st.major {
		majors[t] = true
	}
	m := st.minor[:0]
	for _, t := range st.minor {
		if !majors[t] {
			m = append(m, t)
		}
	}
	st.minor = m
	return st, nil
}
func (g *Graph) timeRange() (float64, float64, error) {
	start, end := math.Inf(1), math.Inf(-1)
	for _, s := range g.Series {
		if len(s.Timestamps) > 0 {
			start = math.Min(start, s.Timestamps[0])
			end = math.Max(end, s.Timestamps[len(s.Timestamps)-1])
		}
	}
	if g.TimeAxis.Start != nil {
		start = *g.TimeAxis.Start
	}
	if g.TimeAxis.End != nil {
		end = *g.TimeAxis.End
	}
	if start == end && g.TimeAxis.Start == nil && g.TimeAxis.End == nil {
		start -= 150
		end += 150
	}
	if !validEpoch(start) || !validEpoch(end) || end <= start {
		return 0, 0, fmt.Errorf("empty data requires explicit start/end; resolved range must be nonzero within years 1..9999")
	}
	return start, end, nil
}
