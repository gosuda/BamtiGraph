// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"image"
	"image/color"
	"math"
	"sort"
)

func solidImage(w, h int, c color.NRGBA) *image.NRGBA {
	im := image.NewNRGBA(image.Rect(0, 0, w, h))
	rectFill(im, im.Bounds(), c)
	return im
}
func rectFill(im *image.NRGBA, r image.Rectangle, c color.NRGBA) {
	r = r.Intersect(im.Bounds())
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			im.SetNRGBA(x, y, c)
		}
	}
}

// rawPixel replaces a pixel in a temporary layer, avoiding alpha accumulation at
// self-intersections. Final layers are subsequently source-over composited.
func rawPixel(im *image.NRGBA, x, y int, c color.NRGBA) {
	if image.Pt(x, y).In(im.Bounds()) {
		im.SetNRGBA(x, y, c)
	}
}
func blendPixel(im *image.NRGBA, x, y int, c color.NRGBA) {
	if c.A == 0 || !image.Pt(x, y).In(im.Bounds()) {
		return
	}
	if c.A == 255 {
		im.SetNRGBA(x, y, c)
		return
	}
	d := im.NRGBAAt(x, y)
	sa, da := uint64(c.A), uint64(d.A)
	alpha := sa*255 + da*(255-sa)
	if alpha == 0 {
		return
	}
	ch := func(s, d uint8) uint8 { return uint8((uint64(s)*sa*255 + uint64(d)*da*(255-sa) + alpha/2) / alpha) }
	im.SetNRGBA(x, y, RGBA(ch(c.R, d.R), ch(c.G, d.G), ch(c.B, d.B), uint8((alpha+127)/255)))
}
func composite(dst, src *image.NRGBA, at image.Point) {
	r := src.Bounds()
	for y := r.Min.Y; y < r.Max.Y; y++ {
		for x := r.Min.X; x < r.Max.X; x++ {
			c := src.NRGBAAt(x, y)
			if c.A != 0 {
				blendPixel(dst, at.X+x-r.Min.X, at.Y+y-r.Min.Y, c)
			}
		}
	}
}
func lineRaster(im *image.NRGBA, a, b point, c color.NRGBA, width int) {
	x0, y0, x1, y1 := iround(a.x), iround(a.y), iround(b.x), iround(b.y)
	if width <= 1 {
		dx, dy := absInt(x1-x0), -absInt(y1-y0)
		sx, sy := -1, -1
		if x0 < x1 {
			sx = 1
		}
		if y0 < y1 {
			sy = 1
		}
		err := dx + dy
		for {
			rawPixel(im, x0, y0, c)
			if x0 == x1 && y0 == y1 {
				break
			}
			e := 2 * err
			if e >= dy {
				err += dy
				x0 += sx
			}
			if e <= dx {
				err += dx
				y0 += sy
			}
		}
		return
	}
	if x0 == x1 {
		rectFill(im, image.Rect(x0-width/2, min(y0, y1), x0+(width-1)/2+1, max(y0, y1)+1), c)
		return
	}
	if y0 == y1 {
		rectFill(im, image.Rect(min(x0, x1), y0-width/2, max(x0, x1)+1, y0+(width-1)/2+1), c)
		return
	}
	dx, dy := float64(x1-x0), float64(y1-y0)
	length := math.Hypot(dx, dy)
	radius := float64(width-1) / 2
	ox, oy := -dy/length*radius, dx/length*radius
	p := []point{{float64(x0) + ox, float64(y0) + oy}, {float64(x1) + ox, float64(y1) + oy}, {float64(x1) - ox, float64(y1) - oy}, {float64(x0) - ox, float64(y0) - oy}}
	for i := range p {
		p[i].x = float64(iround(p[i].x))
		p[i].y = float64(iround(p[i].y))
	}
	polygonRaster(im, p, c)
}
func absInt(x int) int {
	if x < 0 {
		return -x
	}
	return x
}
func dashed(im *image.NRGBA, a, b point, c color.NRGBA, dash *[2]int, width int) {
	if dash == nil {
		lineRaster(im, a, b, c, width)
		return
	}
	length := math.Hypot(b.x-a.x, b.y-a.y)
	if length == 0 {
		rawPixel(im, iround(a.x), iround(a.y), c)
		return
	}
	dx, dy := (b.x-a.x)/length, (b.y-a.y)/length
	for s := 0; s <= int(math.Ceil(length)); s += dash[0] + dash[1] {
		e := math.Min(length, float64(s+dash[0]-1))
		lineRaster(im, point{a.x + dx*float64(s), a.y + dy*float64(s)}, point{a.x + dx*e, a.y + dy*e}, c, width)
	}
}

// polygonRaster fills inclusive integer-coordinate polygons using scanlines.
func polygonRaster(im *image.NRGBA, ps []point, c color.NRGBA) {
	if len(ps) < 3 {
		return
	}
	lo, hi := ps[0].y, ps[0].y
	for _, p := range ps {
		lo = math.Min(lo, p.y)
		hi = math.Max(hi, p.y)
	}
	ya, yb := max(im.Rect.Min.Y, int(math.Ceil(lo))), min(im.Rect.Max.Y-1, int(math.Floor(hi)))
	xs := make([]float64, 0, len(ps))
	for y := ya; y <= yb; y++ {
		xs = xs[:0]
		fy := float64(y)
		prev := ps[len(ps)-1]
		for _, cur := range ps {
			if prev.y == cur.y {
				if fy == cur.y {
					rectFill(im, image.Rect(int(math.Ceil(math.Min(prev.x, cur.x))), y, int(math.Floor(math.Max(prev.x, cur.x)))+1, y+1), c)
				}
			} else if (prev.y <= fy && fy < cur.y) || (cur.y <= fy && fy < prev.y) {
				xs = append(xs, prev.x+(fy-prev.y)/(cur.y-prev.y)*(cur.x-prev.x))
			}
			prev = cur
		}
		sort.Float64s(xs)
		for i := 0; i+1 < len(xs); i += 2 {
			a, b := max(im.Rect.Min.X, int(math.Ceil(xs[i]-1e-9))), min(im.Rect.Max.X-1, int(math.Floor(xs[i+1]+1e-9)))
			for x := a; x <= b; x++ {
				im.SetNRGBA(x, y, c)
			}
		}
	}
}
func circleRaster(im *image.NRGBA, cx, cy, r float64, c color.NRGBA) {
	for y := max(im.Rect.Min.Y, int(math.Floor(cy-r))); y <= min(im.Rect.Max.Y-1, int(math.Ceil(cy+r))); y++ {
		for x := max(im.Rect.Min.X, int(math.Floor(cx-r))); x <= min(im.Rect.Max.X-1, int(math.Ceil(cx+r))); x++ {
			dx, dy := float64(x)-cx, float64(y)-cy
			if dx*dx+dy*dy <= r*r {
				im.SetNRGBA(x, y, c)
			}
		}
	}
}
func boxDown(src *image.NRGBA, scale int) *image.NRGBA {
	if scale == 1 {
		return src
	}
	w, h := src.Rect.Dx()/scale, src.Rect.Dy()/scale
	dst := image.NewNRGBA(image.Rect(0, 0, w, h))
	n := uint64(scale * scale)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			var a, r, g, b uint64
			for sy := 0; sy < scale; sy++ {
				for sx := 0; sx < scale; sx++ {
					c := src.NRGBAAt(x*scale+sx, y*scale+sy)
					ca := uint64(c.A)
					a += ca
					r += uint64(c.R) * ca
					g += uint64(c.G) * ca
					b += uint64(c.B) * ca
				}
			}
			if a > 0 {
				dst.SetNRGBA(x, y, RGBA(uint8((r+a/2)/a), uint8((g+a/2)/a), uint8((b+a/2)/a), uint8((a+n/2)/n)))
			}
		}
	}
	return dst
}
func nearest(src *image.NRGBA, scale int) *image.NRGBA {
	if scale == 1 {
		return src
	}
	dst := image.NewNRGBA(image.Rect(0, 0, src.Rect.Dx()*scale, src.Rect.Dy()*scale))
	for y := 0; y < dst.Rect.Dy(); y++ {
		for x := 0; x < dst.Rect.Dx(); x++ {
			dst.SetNRGBA(x, y, src.NRGBAAt(src.Rect.Min.X+x/scale, src.Rect.Min.Y+y/scale))
		}
	}
	return dst
}
func border(im *image.NRGBA, t Theme) {
	w, h := im.Rect.Dx(), im.Rect.Dy()
	for _, l := range [][4]int{{0, 0, w - 1, 0}, {1, 1, w - 2, 1}, {0, 0, 0, h - 1}, {1, 1, 1, h - 2}} {
		lineRaster(im, point{float64(l[0]), float64(l[1])}, point{float64(l[2]), float64(l[3])}, t.ShadeLight, 1)
	}
	for _, l := range [][4]int{{w - 2, 1, w - 2, h - 1}, {w - 1, 0, w - 1, h - 1}, {1, h - 2, w - 1, h - 2}, {0, h - 1, w - 1, h - 1}} {
		lineRaster(im, point{float64(l[0]), float64(l[1])}, point{float64(l[2]), float64(l[3])}, t.ShadeDark, 1)
	}
}
