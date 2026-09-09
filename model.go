// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
	"time"
)

const Version = "1.0.0"
const MaxPixels = 40_000_000
const MaxTicks = 5000

// Float and String construct optional configuration values, including explicit zero.
func Float(v float64) *float64  { return &v }
func String(v string) *string   { return &v }
func Int(v int) *int            { return &v }
func Epoch(t time.Time) float64 { return float64(t.Unix()) + float64(t.Nanosecond())/1e9 }
func Epochs(ts []time.Time) []float64 {
	out := make([]float64, len(ts))
	for i, t := range ts {
		out[i] = Epoch(t)
	}
	return out
}
func finite(v float64) bool { return !math.IsNaN(v) && !math.IsInf(v, 0) }
func oneLine(s string) bool { return !strings.ContainsAny(s, "\r\n\x00") && len(s) <= 16384 }
func iround(v float64) int  { return int(math.Floor(v + 0.5)) }
func missing() float64      { return math.NaN() }
func clone[T any](s []T) []T {
	if s == nil {
		return nil
	}
	r := make([]T, len(s))
	copy(r, s)
	return r
}

// RGB and RGBA use straight (not premultiplied) alpha.
func RGB(r, g, b uint8) color.NRGBA     { return color.NRGBA{R: r, G: g, B: b, A: 255} }
func RGBA(r, g, b, a uint8) color.NRGBA { return color.NRGBA{R: r, G: g, B: b, A: a} }
func ParseColor(s string) (color.NRGBA, error) {
	names := map[string]string{"black": "#000000", "white": "#ffffff", "red": "#ff0000", "green": "#008000", "blue": "#0000ff", "transparent": "#00000000"}
	if n, ok := names[strings.ToLower(s)]; ok {
		s = n
	}
	if !strings.HasPrefix(s, "#") {
		return color.NRGBA{}, fmt.Errorf("color must be #RGB, #RRGGBB or #RRGGBBAA: %q", s)
	}
	h := s[1:]
	if len(h) == 3 {
		h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
	}
	if len(h) != 6 && len(h) != 8 {
		return color.NRGBA{}, fmt.Errorf("invalid color %q", s)
	}
	b := []uint8{0, 0, 0, 255}
	for i := 0; i < len(h)/2; i++ {
		var n uint8
		for j := 0; j < 2; j++ {
			c := h[i*2+j]
			var v uint8
			switch {
			case c >= '0' && c <= '9':
				v = c - '0'
			case c >= 'a' && c <= 'f':
				v = c - 'a' + 10
			case c >= 'A' && c <= 'F':
				v = c - 'A' + 10
			default:
				return color.NRGBA{}, fmt.Errorf("invalid color %q", s)
			}
			n = n*16 + v
		}
		b[i] = n
	}
	return RGBA(b[0], b[1], b[2], b[3]), nil
}

type Kind string

const (
	Line Kind = "line"
	Area Kind = "area"
)

type Interpolation string

const (
	Linear   Interpolation = "linear"
	StepPost Interpolation = "step-post"
)

type LegendValues struct{ Current, Average, Maximum float64 }

type Series struct {
	Name          string
	Timestamps    []float64
	Values        []float64
	Kind          Kind
	Color         color.NRGBA
	LineWidth     float64
	Outline       *color.NRGBA
	Baseline      float64
	Interpolation Interpolation
	// GapAfter is elapsed seconds; zero disables the implicit-gap threshold.
	GapAfter     float64
	LegendValues *LegendValues
}

func NewSeries(name string, timestamps, values []float64) (Series, error) {
	s := Series{Name: name, Timestamps: clone(timestamps), Values: clone(values), Kind: Line, Color: RGB(0, 0, 204), LineWidth: 0.8, Interpolation: Linear}
	return s, s.Validate()
}
func RegularSeries(name string, values []float64, start, step float64) (Series, error) {
	if !finite(start) || !finite(step) || step <= 0 {
		return Series{}, fmt.Errorf("start must be finite and step positive")
	}
	ts := make([]float64, len(values))
	for i := range ts {
		ts[i] = start + float64(i)*step
	}
	return NewSeries(name, ts, values)
}
func validateSamples(ts, vs []float64) error {
	if len(ts) != len(vs) {
		return fmt.Errorf("timestamps and values must have equal length")
	}
	for i, t := range ts {
		if !validEpoch(t) {
			return fmt.Errorf("timestamp %d is outside finite years 1..9999", i)
		}
		if i > 0 && t <= ts[i-1] {
			return fmt.Errorf("timestamps must be strictly increasing (index %d)", i)
		}
		if math.IsInf(vs[i], 0) {
			return fmt.Errorf("infinite value at index %d; use NaN for gaps", i)
		}
	}
	return nil
}
func (s Series) Validate() error {
	if !oneLine(s.Name) {
		return fmt.Errorf("series name must be a single line (max 16384 bytes)")
	}
	if err := validateSamples(s.Timestamps, s.Values); err != nil {
		return fmt.Errorf("%q: %w", s.Name, err)
	}
	if s.Kind != Line && s.Kind != Area {
		return fmt.Errorf("invalid series kind %q", s.Kind)
	}
	if s.Interpolation != Linear && s.Interpolation != StepPost {
		return fmt.Errorf("invalid interpolation")
	}
	if !finite(s.LineWidth) || s.LineWidth <= 0 || s.LineWidth > 4096 || !finite(s.Baseline) || !finite(s.GapAfter) || s.GapAfter < 0 {
		return fmt.Errorf("invalid line width, baseline or gap threshold")
	}
	if s.LegendValues != nil {
		for _, v := range []float64{s.LegendValues.Current, s.LegendValues.Average, s.LegendValues.Maximum} {
			if math.IsInf(v, 0) {
				return fmt.Errorf("infinite legend override")
			}
		}
	}
	return nil
}

type Tick struct {
	Time  float64 `json:"time"`
	Label string  `json:"label"`
}
type TimeAxis struct {
	Start, End                               *float64
	Mode                                     string
	Timezone                                 string
	MinorSeconds, MajorSeconds, LabelSeconds *float64
	// LabelFormat uses strftime directives, not Go's reference-time format.
	LabelFormat string
	// nil means automatic; an explicitly empty slice disables labels/grid.
	Ticks                  []Tick
	MajorTicks, MinorTicks []float64
	LabelOffsetSeconds     float64
}

func DefaultTimeAxis() TimeAxis { return TimeAxis{Mode: "auto", Timezone: "UTC"} }
func DailyAxis(zone string) TimeAxis {
	a := DefaultTimeAxis()
	a.Mode = "daily"
	if zone != "" {
		a.Timezone = zone
	}
	return a
}
func WeeklyAxis(zone string) TimeAxis  { a := DailyAxis(zone); a.Mode = "weekly"; return a }
func MonthlyAxis(zone string) TimeAxis { a := DailyAxis(zone); a.Mode = "monthly"; return a }
func YearlyAxis(zone string) TimeAxis  { a := DailyAxis(zone); a.Mode = "yearly"; return a }
func validEpoch(v float64) bool        { return finite(v) && v >= -62135596800 && v < 253402300800 }
func (a TimeAxis) Validate() error {
	switch a.Mode {
	case "auto", "daily", "weekly", "monthly", "yearly", "custom":
	default:
		return fmt.Errorf("invalid time axis mode")
	}
	if _, err := time.LoadLocation(a.Timezone); err != nil {
		return fmt.Errorf("timezone: %w", err)
	}
	for _, p := range []*float64{a.Start, a.End} {
		if p != nil && !validEpoch(*p) {
			return fmt.Errorf("invalid time bound")
		}
	}
	if a.Start != nil && a.End != nil && *a.Start >= *a.End {
		return fmt.Errorf("end must be later than start")
	}
	for _, p := range []*float64{a.MinorSeconds, a.MajorSeconds, a.LabelSeconds} {
		if p != nil && (!finite(*p) || *p <= 0) {
			return fmt.Errorf("tick interval must be positive and finite")
		}
	}
	if !finite(a.LabelOffsetSeconds) || !oneLine(a.LabelFormat) {
		return fmt.Errorf("invalid time label format/offset")
	}
	if len(a.Ticks) > MaxTicks || len(a.MajorTicks) > MaxTicks || len(a.MinorTicks) > MaxTicks {
		return fmt.Errorf("too many explicit ticks")
	}
	for _, t := range a.Ticks {
		if !validEpoch(t.Time) || !oneLine(t.Label) {
			return fmt.Errorf("invalid explicit tick")
		}
	}
	for _, ts := range [][]float64{a.MajorTicks, a.MinorTicks} {
		for _, t := range ts {
			if !validEpoch(t) {
				return fmt.Errorf("invalid explicit grid tick")
			}
		}
	}
	return nil
}

type YAxis struct {
	Minimum, Maximum, MajorStep *float64
	MinorDivisions              int
	Base                        int
	ScaleFactor                 *float64
	Suffix                      *string
	Decimals                    *int
	LegendDecimals              int
	ShowZeroSuffix              bool
}

func DefaultYAxis() YAxis {
	return YAxis{Minimum: Float(0), MinorDivisions: 5, Base: 1000, LegendDecimals: 2}
}
func (a YAxis) Validate() error {
	for _, p := range []*float64{a.Minimum, a.Maximum, a.MajorStep, a.ScaleFactor} {
		if p != nil && !finite(*p) {
			return fmt.Errorf("y-axis parameters must be finite")
		}
	}
	if a.Minimum != nil && a.Maximum != nil && *a.Minimum >= *a.Maximum {
		return fmt.Errorf("y maximum must exceed minimum")
	}
	for _, p := range []*float64{a.MajorStep, a.ScaleFactor} {
		if p != nil && *p <= 0 {
			return fmt.Errorf("y step and scale must be positive")
		}
	}
	if a.Base != 1000 && a.Base != 1024 {
		return fmt.Errorf("unit base must be 1000 or 1024")
	}
	if a.MinorDivisions < 1 || a.MinorDivisions > 100 {
		return fmt.Errorf("minor divisions must be 1..100")
	}
	if a.LegendDecimals < 0 || a.LegendDecimals > 12 || (a.Decimals != nil && (*a.Decimals < 0 || *a.Decimals > 12)) {
		return fmt.Errorf("decimals must be 0..12")
	}
	if a.Suffix != nil && !oneLine(*a.Suffix) {
		return fmt.Errorf("invalid unit suffix")
	}
	return nil
}

type LegendLayout struct {
	NameX, SwatchX, SwatchWidth, SwatchHeight, ReferenceWidth int
	Compact, Expanded, Aligned                                [3][2]float64
	AutoScaleColumns                                          bool
}

func DefaultLegendLayout() LegendLayout {
	return LegendLayout{30, 15, 9, 10, 595, [3][2]float64{{102, 228}, {244, 370}, {386, 512}}, [3][2]float64{{124, 250}, {289, 415}, {454, 580}}, [3][2]float64{{102, 250}, {267, 415}, {432, 580}}, true}
}

type Layout struct {
	Width, PlotHeight, Left, Right, Top                                   int
	TitleY                                                                int
	TitleOffsetX                                                          float64
	UnitX, XLabelGap, YLabelGap, LegendGap, LegendRowHeight, LegendBottom int
	Legend                                                                string
	LegendLayout                                                          LegendLayout
	Antialias, PixelScale                                                 int
}

func ReferenceLayout() Layout {
	return Layout{595, 122, 64, 31, 34, 8, 27, 5, 5, 6, 20, 14, 7, "reference", DefaultLegendLayout(), 4, 1}
}
func (l Layout) PlotBox() [4]int {
	return [4]int{l.Left, l.Top, l.Width - l.Right, l.Top + l.PlotHeight}
}
func (l Layout) Height(n int) int {
	if l.Legend == "none" {
		return l.Top + l.PlotHeight + 18
	}
	return l.Top + l.PlotHeight + l.LegendGap + max(1, n)*l.LegendRowHeight + l.LegendBottom
}
func (l Layout) Validate() error {
	if l.Width < 400 || l.Width > 8192 || l.PlotHeight < 30 || l.PlotHeight > 4096 {
		return fmt.Errorf("layout: width 400..8192, plot height 30..4096 required")
	}
	if l.Left < 20 || l.Right < 12 || l.Top < 12 || l.Width-l.Left-l.Right < 100 {
		return fmt.Errorf("layout: insufficient plot margins")
	}
	for _, v := range []int{l.Left, l.Right, l.Top, l.TitleY, l.UnitX, l.XLabelGap, l.YLabelGap, l.LegendGap, l.LegendRowHeight, l.LegendBottom} {
		if v < 0 || v > 16384 {
			return fmt.Errorf("layout: invalid coordinate")
		}
	}
	if !finite(l.TitleOffsetX) || math.Abs(l.TitleOffsetX) > 8192 {
		return fmt.Errorf("invalid title offset")
	}
	if l.Legend != "reference" && l.Legend != "aligned" && l.Legend != "none" {
		return fmt.Errorf("invalid legend mode")
	}
	if l.Legend != "none" && (l.LegendGap < 14 || l.LegendRowHeight < 10) {
		return fmt.Errorf("legend needs gap >=14, row height >=10")
	}
	if l.Antialias < 1 || l.Antialias > 8 || l.PixelScale < 1 || l.PixelScale > 8 {
		return fmt.Errorf("antialias and pixel scale must be 1..8")
	}
	ll := l.LegendLayout
	if ll.NameX < 0 || ll.SwatchX < 0 || ll.ReferenceWidth <= ll.NameX || ll.SwatchWidth < 3 || ll.SwatchHeight < 3 {
		return fmt.Errorf("invalid legend geometry")
	}
	for _, pairs := range [][3][2]float64{ll.Compact, ll.Expanded, ll.Aligned} {
		last := float64(ll.NameX)
		for _, p := range pairs {
			if !finite(p[0]) || !finite(p[1]) || !(last < p[0] && p[0] < p[1] && p[1] < float64(ll.ReferenceWidth)) {
				return fmt.Errorf("invalid legend columns")
			}
			last = p[1]
		}
	}
	return nil
}

type Theme struct {
	Background, Canvas, ShadeLight, ShadeDark, Text, MinorGrid, MajorGrid, Axis, Arrow, Watermark, Frame color.NRGBA
	GridFront                                                                                            bool
	GridDash                                                                                             [2]int
	TitleSize, AxisSize, UnitSize, LegendSize, WatermarkSize, CaptionSize                                float64
	TitleAdvance, AxisAdvance, LegendAdvance                                                             float64
}

func DefaultTheme() Theme {
	return Theme{RGB(243, 243, 243), RGB(255, 255, 255), RGB(207, 207, 207), RGB(158, 158, 158), RGB(0, 0, 0), RGBA(143, 143, 143, 60), RGBA(223, 79, 79, 60), RGB(119, 119, 119), RGB(127, 31, 31), RGB(170, 170, 170), RGB(0, 0, 0), true, [2]int{1, 1}, 14, 11, 10, 11, 8, 11, 8, 6, 7}
}
func (t Theme) Validate() error {
	if t.GridDash[0] < 1 || t.GridDash[1] < 1 || t.GridDash[0] > 16384 || t.GridDash[1] > 16384 {
		return fmt.Errorf("invalid grid dash")
	}
	for _, s := range []float64{t.TitleSize, t.AxisSize, t.UnitSize, t.LegendSize, t.WatermarkSize, t.CaptionSize, t.TitleAdvance, t.AxisAdvance, t.LegendAdvance} {
		if !finite(s) || s <= 0 || s > 256 {
			return fmt.Errorf("font sizes/advances must be >0 and <=256")
		}
	}
	return nil
}

type FontConfig struct {
	Mono, Title, Unit, Caption string
	Strict                     bool
	FaceIndex                  int
	// Backend may provide a custom rasterizer. nil uses the dependency-free,
	// unhinted TrueType backend. Each Open call must return an independently owned face.
	Backend FontBackend `json:"-"`
}

// Glyph is a coverage mask and its placement relative to an integer baseline.
type Glyph struct {
	Mask    *image.Alpha
	X, Y    int
	Advance float64
}
type FontFace interface {
	Glyph(rune) (Glyph, error)
	Close() error
}
type FontBackend interface {
	Open(path string, index int, size float64) (FontFace, error)
	ID() string
}

type HRule struct {
	Value float64
	Color color.NRGBA
	Width float64
	Dash  *[2]int
}
type VRule struct {
	Time  float64
	Color color.NRGBA
	Width float64
	Dash  *[2]int
}

func HorizontalRule(v float64) HRule { return HRule{v, RGB(153, 0, 0), 1, &[2]int{3, 2}} }
func VerticalRule(t float64) VRule   { return VRule{t, RGB(153, 0, 0), 1, &[2]int{3, 2}} }
func validateRule(v, w float64, d *[2]int) error {
	if !finite(v) || !finite(w) || w <= 0 || w > 4096 {
		return fmt.Errorf("invalid rule value/width")
	}
	if d != nil && (d[0] < 1 || d[1] < 1 || d[0] > 16384 || d[1] > 16384) {
		return fmt.Errorf("invalid rule dash")
	}
	return nil
}

// Graph is a configurable native chart. NewGraph copies incoming samples;
// subsequent configuration is explicit and mutable. Do not mutate a graph while
// rendering it concurrently. Watermark defaults to empty; custom text is drawn
// vertically in the right margin without modifying plot geometry.
type Graph struct {
	Series                          []Series
	Title, VerticalLabel, Watermark string
	TimeAxis                        TimeAxis
	YAxis                           YAxis
	Layout                          Layout
	Theme                           Theme
	Fonts                           FontConfig
	HRules                          []HRule
	VRules                          []VRule
}

func NewGraph(series ...Series) *Graph {
	copied := clone(series)
	for i := range copied {
		copied[i].Timestamps = clone(copied[i].Timestamps)
		copied[i].Values = clone(copied[i].Values)
		if copied[i].Outline != nil {
			c := *copied[i].Outline
			copied[i].Outline = &c
		}
		if copied[i].LegendValues != nil {
			v := *copied[i].LegendValues
			copied[i].LegendValues = &v
		}
	}
	return &Graph{Series: copied, Title: "Traffic - ether1", VerticalLabel: "bits per second", Watermark: "", TimeAxis: DefaultTimeAxis(), YAxis: DefaultYAxis(), Layout: ReferenceLayout(), Theme: DefaultTheme()}
}
func Traffic(timestamps, inbound, outbound []float64) (*Graph, error) {
	a, err := NewSeries("Inbound", timestamps, inbound)
	if err != nil {
		return nil, err
	}
	a.Kind = Area
	a.Color = RGB(0, 204, 0)
	c := RGB(0, 48, 0)
	a.Outline = &c
	a.LineWidth = .7
	b, err := NewSeries("Outbound", timestamps, outbound)
	if err != nil {
		return nil, err
	}
	b.LineWidth = .7
	return NewGraph(a, b), nil
}
func (g *Graph) Validate() error {
	if g == nil {
		return fmt.Errorf("nil graph")
	}
	if len(g.Series) > 128 {
		return fmt.Errorf("at most 128 series are supported")
	}
	for _, s := range g.Series {
		if err := s.Validate(); err != nil {
			return err
		}
	}
	for _, s := range []string{g.Title, g.VerticalLabel, g.Watermark} {
		if !oneLine(s) {
			return fmt.Errorf("graph text must be single-line and <=16384 bytes")
		}
	}
	for _, fn := range []func() error{g.TimeAxis.Validate, g.YAxis.Validate, g.Layout.Validate, g.Theme.Validate} {
		if e := fn(); e != nil {
			return e
		}
	}
	if g.Fonts.FaceIndex < 0 || g.Fonts.FaceIndex > 65535 {
		return fmt.Errorf("invalid font face index")
	}
	if g.Fonts.Strict && g.Fonts.Mono == "" {
		return fmt.Errorf("strict font mode requires Mono path")
	}
	if len(g.HRules)+len(g.VRules) > MaxTicks {
		return fmt.Errorf("too many rules")
	}
	for _, r := range g.HRules {
		if err := validateRule(r.Value, r.Width, r.Dash); err != nil {
			return err
		}
	}
	for _, r := range g.VRules {
		if !validEpoch(r.Time) {
			return fmt.Errorf("invalid rule timestamp")
		}
		if err := validateRule(r.Time, r.Width, r.Dash); err != nil {
			return err
		}
	}
	l := g.Layout
	pw := l.Width - l.Left - l.Right
	if int64(pw+1)*int64(l.PlotHeight+1)*int64(l.Antialias*l.Antialias) > MaxPixels || int64(l.Width)*int64(l.Height(len(g.Series)))*int64(l.PixelScale*l.PixelScale) > MaxPixels {
		return fmt.Errorf("layout: render exceeds 40 million pixel allocation guard")
	}
	return nil
}
