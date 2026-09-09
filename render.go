// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"image"
	"image/color"
	"math"
)

type DrawnTick struct {
	Time  float64 `json:"time"`
	Label string  `json:"label"`
	X     float64 `json:"x"`
}
type SeriesStatistics struct {
	Name string `json:"name"`
	Statistics
	DisplayOverride *DisplayValues `json:"display_override"`
}
type DisplayValues struct {
	Current *float64 `json:"current"`
	Average *float64 `json:"average"`
	Maximum *float64 `json:"maximum"`
}
type Unit struct {
	Factor float64 `json:"factor"`
	Suffix string  `json:"suffix"`
}
type Metadata struct {
	Renderer         string             `json:"renderer"`
	Version          string             `json:"version"`
	Environment      Environment        `json:"environment"`
	ImageSize        [2]int             `json:"image_size"`
	LogicalSize      [2]int             `json:"logical_size"`
	PlotBox          [4]int             `json:"plot_box"`
	PixelScale       int                `json:"pixel_scale"`
	Layout           Layout             `json:"layout"`
	Theme            Theme              `json:"theme"`
	Title            string             `json:"title"`
	VerticalLabel    string             `json:"vertical_label"`
	Watermark        string             `json:"watermark"`
	TimeRange        [2]float64         `json:"time_range"`
	Timezone         string             `json:"timezone"`
	YRange           [2]float64         `json:"y_range"`
	YStep            float64            `json:"y_step"`
	YUnit            Unit               `json:"y_unit"`
	XLabels          []DrawnTick        `json:"x_labels"`
	StatisticsPolicy string             `json:"statistics_policy"`
	Statistics       []SeriesStatistics `json:"statistics"`
}
type RenderResult struct {
	Image    *image.NRGBA
	Metadata Metadata
}

func (g *Graph) Render() (*image.NRGBA, error) {
	r, e := g.RenderResult()
	if e != nil {
		return nil, e
	}
	return r.Image, nil
}
func (g *Graph) RenderResult() (*RenderResult, error) {
	if e := g.Validate(); e != nil {
		return nil, e
	}
	l, t := g.Layout, g.Theme
	left, top, right, bottom := l.Left, l.Top, l.Width-l.Right, l.Top+l.PlotHeight
	pw, ph := right-left, bottom-top
	height := l.Height(len(g.Series))
	start, end, e := g.timeRange()
	if e != nil {
		return nil, e
	}
	runs := make([][][]point, len(g.Series))
	var vals []float64
	stats := make([]Statistics, len(g.Series))
	for i, s := range g.Series {
		runs[i] = visibleRuns(s, start, end)
		for _, r := range runs[i] {
			for _, p := range r {
				vals = append(vals, p.y)
			}
		}
		if s.Kind == Area {
			vals = append(vals, s.Baseline)
		}
		stats[i] = statsFor(s, start, end)
	}
	ys, e := resolveY(g.YAxis, vals)
	if e != nil {
		return nil, e
	}
	xs, e := resolveX(g.TimeAxis, start, end, pw)
	if e != nil {
		return nil, e
	}
	fonts, e := newFonts(g.Fonts, t)
	if e != nil {
		return nil, e
	}
	defer fonts.close()
	im := solidImage(l.Width, height, t.Background)
	rectFill(im, image.Rect(left, top, right+1, bottom+1), t.Canvas)
	xx := func(v float64) float64 { return (v - start) / (end - start) * float64(pw) }
	yy := func(v float64) float64 {
		return float64(ph) * (1 - (v/(ys.maximum-ys.minimum) - ys.minimum/(ys.maximum-ys.minimum)))
	}
	projected := make([][][]point, len(runs))
	for i, rr := range runs {
		for _, r := range rr {
			ps := decimate(r, start, end, pw)
			pr := make([]point, len(ps))
			for k, p := range ps {
				x, y := xx(p.x), yy(p.y)
				if !finite(x) || !finite(y) || math.Abs(y) > 1e15 {
					return nil, fmt.Errorf("data too large relative to chosen y-axis")
				}
				pr[k] = point{x, y}
			}
			projected[i] = append(projected[i], pr)
		}
	}
	grid := func() {
		layer := image.NewNRGBA(im.Rect)
		for _, item := range []struct {
			vs []float64
			c  color.NRGBA
		}{{ys.minor, t.MinorGrid}, {ys.major, t.MajorGrid}} {
			for _, v := range item.vs {
				y := top + iround(yy(v))
				if y >= top && y <= bottom {
					dashed(layer, point{float64(left), float64(y)}, point{float64(right), float64(y)}, item.c, &t.GridDash, 1)
				}
			}
		}
		for _, item := range []struct {
			vs []float64
			c  color.NRGBA
		}{{xs.minor, t.MinorGrid}, {xs.major, t.MajorGrid}} {
			for _, v := range item.vs {
				x := left + iround(xx(v))
				dashed(layer, point{float64(x), float64(top)}, point{float64(x), float64(bottom)}, item.c, &t.GridDash, 1)
			}
		}
		composite(im, layer, image.Point{})
	}
	aa := l.Antialias
	layerRect := image.Rect(0, 0, (pw+1)*aa, (ph+1)*aa)
	if !t.GridFront {
		grid()
	}
	for i, s := range g.Series {
		if s.Kind != Area {
			continue
		}
		base := yy(s.Baseline)
		if !finite(base) || math.Abs(base) > 1e15 {
			return nil, fmt.Errorf("baseline too large for selected y-axis")
		}
		layer := image.NewNRGBA(layerRect)
		for _, ps := range projected[i] {
			if len(ps) < 2 {
				continue
			}
			poly := make([]point, 0, len(ps)+2)
			poly = append(poly, point{ps[0].x, base})
			poly = append(poly, ps...)
			poly = append(poly, point{ps[len(ps)-1].x, base})
			poly = clipPolygon(poly, float64(pw), float64(ph))
			for k := range poly {
				poly[k] = point{float64(iround(poly[k].x * float64(aa))), float64(iround(poly[k].y * float64(aa)))}
			}
			polygonRaster(layer, poly, s.Color)
		}
		composite(im, boxDown(layer, aa), image.Pt(left, top))
	}
	if t.GridFront {
		grid()
	}
	for i, s := range g.Series {
		var c color.NRGBA
		if s.Kind == Line {
			c = s.Color
		} else {
			if s.Outline == nil {
				continue
			}
			c = *s.Outline
		}
		layer := image.NewNRGBA(layerRect)
		width := max(1, iround(s.LineWidth*float64(aa)))
		for _, ps := range projected[i] {
			if len(ps) == 1 {
				p := ps[0]
				if p.x >= 0 && p.x <= float64(pw) && p.y >= 0 && p.y <= float64(ph) {
					circleRaster(layer, p.x*float64(aa), p.y*float64(aa), math.Max(float64(aa)/2, float64(width)/2), c)
				}
			}
			for k := 1; k < len(ps); k++ {
				a, b, ok := clipLine(ps[k-1], ps[k], float64(pw), float64(ph))
				if ok {
					a = point{float64(iround(a.x * float64(aa))), float64(iround(a.y * float64(aa)))}
					b = point{float64(iround(b.x * float64(aa))), float64(iround(b.y * float64(aa)))}
					lineRaster(layer, a, b, c, width)
				}
			}
		}
		composite(im, boxDown(layer, aa), image.Pt(left, top))
	}
	rules := image.NewNRGBA(image.Rect(0, 0, pw+1, ph+1))
	for _, r := range g.HRules {
		if r.Value >= ys.minimum && r.Value <= ys.maximum {
			y := float64(iround(yy(r.Value)))
			dashed(rules, point{0, y}, point{float64(pw), y}, r.Color, r.Dash, max(1, iround(r.Width)))
		}
	}
	for _, r := range g.VRules {
		if r.Time >= start && r.Time <= end {
			x := float64(iround(xx(r.Time)))
			dashed(rules, point{x, 0}, point{x, float64(ph)}, r.Color, r.Dash, max(1, iround(r.Width)))
		}
	}
	composite(im, rules, image.Pt(left, top))
	lineRaster(im, point{float64(left), float64(top - 3)}, point{float64(left), float64(bottom + 4)}, t.Axis, 1)
	lineRaster(im, point{float64(left - 4), float64(bottom)}, point{float64(right + 4), float64(bottom)}, t.Axis, 1)
	polygonRaster(im, []point{{float64(left), float64(top - 5)}, {float64(left - 3), float64(top)}, {float64(left + 3), float64(top)}}, t.Arrow)
	polygonRaster(im, []point{{float64(right + 7), float64(bottom)}, {float64(right + 2), float64(bottom - 3)}, {float64(right + 2), float64(bottom + 3)}}, t.Arrow)
	axisFont, e := fonts.get("axis", 1)
	if e != nil {
		return nil, e
	}
	for _, v := range ys.major {
		y := float64(top) + yy(v)
		label := ys.label(v, g.YAxis.ShowZeroSuffix)
		w, err := textWidth(label, axisFont, t.AxisAdvance)
		if err != nil {
			return nil, err
		}
		if w > float64(left-l.YLabelGap-20) {
			return nil, fmt.Errorf("layout: y labels overlap vertical label; increase Left or use coarser scale")
		}
		lineRaster(im, point{float64(left - 3), float64(iround(y))}, point{float64(left), float64(iround(y))}, t.Axis, 1)
		if err = drawText(im, float64(left-l.YLabelGap), y, label, axisFont, t.Text, t.AxisAdvance, "right", true); err != nil {
			return nil, err
		}
	}
	lastRight := math.Inf(-1)
	drawn := []DrawnTick{}
	for _, tick := range xs.labels {
		x := float64(left) + xx(tick.Time)
		tw, err := textWidth(tick.Label, axisFont, t.AxisAdvance)
		if err != nil {
			return nil, err
		}
		lx, rx := x-tw/2, x+tw/2
		if g.TimeAxis.Ticks == nil && lx < lastRight+3 {
			continue
		}
		if lx < 2 || rx > float64(l.Width-3) {
			continue
		}
		c := t.MajorGrid
		c.A = 255
		lineRaster(im, point{float64(iround(x)), float64(bottom)}, point{float64(iround(x)), float64(bottom + 3)}, c, 1)
		if err = drawText(im, x, float64(bottom+l.XLabelGap), tick.Label, axisFont, t.Text, t.AxisAdvance, "center", false); err != nil {
			return nil, err
		}
		drawn = append(drawn, DrawnTick{tick.Time, tick.Label, x})
		lastRight = rx
	}
	titleFont, e := fonts.get("title", 1)
	if e != nil {
		return nil, e
	}
	titleX := float64(left+right)/2 + l.TitleOffsetX
	maxTitle := 2 * math.Min(titleX-5, float64(l.Width-14)-titleX)
	title, e := fitText(g.Title, maxTitle, titleFont, t.TitleAdvance)
	if e != nil {
		return nil, e
	}
	if e = drawText(im, titleX, float64(l.TitleY), title, titleFont, t.Text, t.TitleAdvance, "center", false); e != nil {
		return nil, e
	}
	if g.VerticalLabel != "" {
		face, err := fonts.get("unit", 1)
		if err != nil {
			return nil, err
		}
		label, err := rotatedText(g.VerticalLabel, face, t.Text, false)
		if err != nil {
			return nil, err
		}
		if label.Rect.Dy() > ph+18 || l.UnitX+label.Rect.Dx() > left-l.YLabelGap {
			return nil, fmt.Errorf("layout: vertical label does not fit")
		}
		composite(im, label, image.Pt(l.UnitX, iround(float64(top+bottom-label.Rect.Dy())/2)))
	}
	if g.Watermark != "" {
		face, err := fonts.get("watermark", 1)
		if err != nil {
			return nil, err
		}
		mark, err := rotatedText(g.Watermark, face, t.Watermark, true)
		if err != nil {
			return nil, err
		}
		if mark.Rect.Dy() > height-8 || mark.Rect.Dx() > l.Right-9 {
			return nil, fmt.Errorf("layout: watermark does not fit")
		}
		composite(im, mark, image.Pt(l.Width-mark.Rect.Dx()-4, 4))
	}
	if l.Legend != "none" {
		if e = drawLegend(im, g, stats, ys, fonts); e != nil {
			return nil, e
		}
	}
	border(im, t)
	for i := 3; i < len(im.Pix); i += 4 {
		im.Pix[i] = 255
	}
	im = nearest(im, l.PixelScale)
	meta := Metadata{Renderer: "bamtigraph-go", Version: Version, Environment: fonts.environment, ImageSize: [2]int{im.Rect.Dx(), im.Rect.Dy()}, LogicalSize: [2]int{l.Width, height}, PlotBox: l.PlotBox(), PixelScale: l.PixelScale, Layout: l, Theme: t, Title: g.Title, VerticalLabel: g.VerticalLabel, Watermark: g.Watermark, TimeRange: [2]float64{start, end}, Timezone: g.TimeAxis.Timezone, YRange: [2]float64{ys.minimum, ys.maximum}, YStep: ys.step, YUnit: Unit{ys.factor, ys.suffix}, XLabels: drawn, StatisticsPolicy: "inclusive viewport; arithmetic sample mean; current is last sample including missing", Statistics: make([]SeriesStatistics, len(stats))}
	for i, s := range g.Series {
		meta.Statistics[i] = SeriesStatistics{Name: s.Name, Statistics: stats[i]}
		if s.LegendValues != nil {
			v := s.LegendValues
			meta.Statistics[i].DisplayOverride = &DisplayValues{nullable(v.Current), nullable(v.Average), nullable(v.Maximum)}
		}
	}
	return &RenderResult{im, meta}, nil
}
func drawLegend(im *image.NRGBA, g *Graph, stats []Statistics, ys yState, f *fontManager) error {
	l, t := g.Layout, g.Theme
	scale := math.Min(1, float64(l.Width-40)/555)
	face, e := f.get("legend", scale)
	if e != nil {
		return e
	}
	advance := t.LegendAdvance * scale
	y0 := l.Top + l.PlotHeight + l.LegendGap
	ll := l.LegendLayout
	factor := 1.
	if ll.AutoScaleColumns {
		factor = float64(l.Width-ll.NameX) / float64(ll.ReferenceWidth-ll.NameX)
	}
	xx := func(x float64) float64 { return float64(ll.NameX) + (x-float64(ll.NameX))*factor }
	for i, s := range g.Series {
		pairs := ll.Compact
		if l.Legend == "reference" && i == len(g.Series)-1 && len(g.Series) > 1 {
			pairs = ll.Expanded
		}
		if l.Legend == "aligned" {
			pairs = ll.Aligned
		}
		y := y0 + i*l.LegendRowHeight
		if ll.SwatchHeight > l.LegendRowHeight || ll.SwatchX+ll.SwatchWidth >= l.Width || y+ll.SwatchHeight > im.Rect.Dy()-2 {
			return fmt.Errorf("layout: legend swatch does not fit")
		}
		rectFill(im, image.Rect(ll.SwatchX, y, ll.SwatchX+ll.SwatchWidth, y+ll.SwatchHeight), t.Frame)
		rectFill(im, image.Rect(ll.SwatchX+1, y+1, ll.SwatchX+ll.SwatchWidth-1, y+ll.SwatchHeight-1), s.Color)
		name, e := fitText(s.Name, xx(pairs[0][0])-float64(ll.NameX)-12, face, advance)
		if e != nil {
			return e
		}
		if e = drawText(im, float64(ll.NameX), float64(y), name, face, t.Text, advance, "left", false); e != nil {
			return e
		}
		values := [3]*float64{stats[i].Current, stats[i].Average, stats[i].Maximum}
		if s.LegendValues != nil {
			v := s.LegendValues
			values = [3]*float64{nullable(v.Current), nullable(v.Average), nullable(v.Maximum)}
		}
		for j, label := range []string{"Current:", "Average:", "Maximum:"} {
			number := legendNumber(values[j], g.YAxis, ys)
			a, b := xx(pairs[j][0]), xx(pairs[j][1])
			if b >= float64(l.Width-3) {
				return fmt.Errorf("layout: legend column outside panel")
			}
			lw, e := textWidth(label, face, advance)
			if e != nil {
				return e
			}
			nw, e := textWidth(number, face, advance)
			if e != nil {
				return e
			}
			if lw+nw+7*scale > b-a+1 {
				return fmt.Errorf("layout: legend statistic overlaps; increase width or reduce LegendDecimals")
			}
			if e = drawText(im, a, float64(y), label, face, t.Text, advance, "left", false); e != nil {
				return e
			}
			if e = drawText(im, b, float64(y), number, face, t.Text, advance, "right", false); e != nil {
				return e
			}
		}
	}
	return nil
}
