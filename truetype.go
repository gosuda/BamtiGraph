// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

// TrueType-outline reader and unhinted quadratic rasterizer. This implementation
// follows the public OpenType table specifications; it is not a copy of FreeType
// or x/image. All binary reads, recursion and allocations are bounded.

import (
	"encoding/binary"
	"fmt"
	"image"
	"io"
	"math"
	"os"
	"sort"
)

// TrueTypeBackend is dependency-free and deterministic; it does not execute
// font bytecode. It accepts TrueType glyf outlines, not CFF/CFF2, WOFF or WOFF2.
type TrueTypeBackend struct{}

func (TrueTypeBackend) ID() string { return "bamtigraph-go/truetype-unhinted-v1/coverage8" }
func (TrueTypeBackend) Open(path string, index int, size float64) (FontFace, error) {
	if !finite(size) || size <= 0 || size > 2048 {
		return nil, fmt.Errorf("invalid font size")
	}
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	st, e := f.Stat()
	if e != nil {
		return nil, e
	}
	if st.Size() < 12 || st.Size() > 64<<20 {
		return nil, fmt.Errorf("font file must be 12 bytes..64 MiB")
	}
	data, e := io.ReadAll(io.LimitReader(f, (64<<20)+1))
	if e != nil {
		return nil, e
	}
	if len(data) > 64<<20 {
		return nil, fmt.Errorf("font exceeds 64 MiB")
	}
	return parseTrueType(data, index, size)
}

type fontPoint struct {
	x, y float64
	on   bool
}
type ttFace struct {
	data                                    []byte
	tables                                  map[string][]byte
	upem, numGlyphs, numMetrics, locaFormat int
	size                                    float64
	cmap                                    []byte
	cache                                   map[rune]Glyph
}

func u16(b []byte, o int) (uint16, error) {
	if o < 0 || o+2 > len(b) {
		return 0, fmt.Errorf("truncated TrueType table")
	}
	return binary.BigEndian.Uint16(b[o : o+2]), nil
}
func u32(b []byte, o int) (uint32, error) {
	if o < 0 || o+4 > len(b) {
		return 0, fmt.Errorf("truncated TrueType table")
	}
	return binary.BigEndian.Uint32(b[o : o+4]), nil
}
func parseTrueType(data []byte, index int, size float64) (*ttFace, error) {
	if index < 0 || len(data) < 12 || len(data) > 64<<20 {
		return nil, fmt.Errorf("invalid font data/index")
	}
	offset := 0
	if string(data[:4]) == "ttcf" {
		n, e := u32(data, 8)
		if e != nil || n > 65535 || index >= int(n) {
			return nil, fmt.Errorf("invalid TTC face index")
		}
		off, e := u32(data, 12+index*4)
		if e != nil || uint64(off)+12 > uint64(len(data)) {
			return nil, fmt.Errorf("invalid TTC offset")
		}
		offset = int(off)
	} else if index != 0 {
		return nil, fmt.Errorf("face index requires TTC")
	}
	if string(data[offset:offset+4]) == "OTTO" {
		return nil, fmt.Errorf("CFF/CFF2 outlines are unsupported; use a TrueType-outline TTF/TTC or FontBackend")
	}
	magic := binary.BigEndian.Uint32(data[offset : offset+4])
	if magic != 0x00010000 && string(data[offset:offset+4]) != "true" {
		return nil, fmt.Errorf("not a TrueType-outline font")
	}
	nt, e := u16(data, offset+4)
	if e != nil || nt > 512 || offset+12+int(nt)*16 > len(data) {
		return nil, fmt.Errorf("invalid TrueType directory")
	}
	face := &ttFace{data: data, tables: map[string][]byte{}, size: size, cache: map[rune]Glyph{}}
	for i := 0; i < int(nt); i++ {
		p := offset + 12 + i*16
		tag := string(data[p : p+4])
		off, _ := u32(data, p+8)
		length, _ := u32(data, p+12)
		if uint64(off)+uint64(length) > uint64(len(data)) {
			return nil, fmt.Errorf("invalid %s table range", tag)
		}
		if _, ok := face.tables[tag]; ok {
			return nil, fmt.Errorf("duplicate table %s", tag)
		}
		face.tables[tag] = data[int(off) : int(off)+int(length)]
	}
	for tag, n := range map[string]int{"head": 54, "maxp": 6, "hhea": 36, "hmtx": 4, "loca": 2, "glyf": 0, "cmap": 4} {
		b, ok := face.tables[tag]
		if !ok || len(b) < n {
			return nil, fmt.Errorf("missing/truncated %s table", tag)
		}
	}
	units, _ := u16(face.tables["head"], 18)
	ng, _ := u16(face.tables["maxp"], 4)
	nm, _ := u16(face.tables["hhea"], 34)
	lf, _ := u16(face.tables["head"], 50)
	face.upem, face.numGlyphs, face.numMetrics, face.locaFormat = int(units), int(ng), int(nm), int(lf)
	if units < 16 || units > 16384 || ng == 0 || nm == 0 || nm > ng || lf > 1 {
		return nil, fmt.Errorf("invalid TrueType metrics")
	}
	locSize := (int(ng) + 1) * 2
	if lf == 1 {
		locSize *= 2
	}
	if len(face.tables["loca"]) < locSize || len(face.tables["hmtx"]) < int(nm)*4+2*(int(ng)-int(nm)) {
		return nil, fmt.Errorf("truncated TrueType offsets/metrics")
	}
	cmap := face.tables["cmap"]
	count, e := u16(cmap, 2)
	if e != nil || count > 512 || 4+int(count)*8 > len(cmap) {
		return nil, fmt.Errorf("invalid cmap directory")
	}
	best := -1
	for i := 0; i < int(count); i++ {
		p := 4 + i*8
		platform, _ := u16(cmap, p)
		enc, _ := u16(cmap, p+2)
		if platform != 0 && !(platform == 3 && (enc == 1 || enc == 10)) {
			continue
		}
		off, _ := u32(cmap, p+4)
		if uint64(off)+2 > uint64(len(cmap)) {
			return nil, fmt.Errorf("invalid cmap offset")
		}
		b := cmap[int(off):]
		format, _ := u16(b, 0)
		score := 0
		var length int
		switch format {
		case 4:
			l, er := u16(b, 2)
			if er != nil {
				return nil, er
			}
			length = int(l)
			score = 1
		case 12:
			l, er := u32(b, 4)
			if er != nil || uint64(l) > uint64(len(b)) {
				return nil, fmt.Errorf("invalid cmap12 length")
			}
			length = int(l)
			score = 2
		default:
			continue
		}
		if length > len(b) || length < 16 {
			return nil, fmt.Errorf("invalid cmap length")
		}
		if score > best {
			best = score
			face.cmap = b[:length]
		}
	}
	if face.cmap == nil {
		return nil, fmt.Errorf("font lacks Unicode cmap format 4/12")
	}
	return face, nil
}
func (f *ttFace) Close() error { f.cache = nil; f.data = nil; f.tables = nil; return nil }
func (f *ttFace) glyphIndex(r rune) (int, error) {
	if r < 0 || r > 0x10ffff {
		return 0, nil
	}
	b := f.cmap
	fmtID, _ := u16(b, 0)
	if fmtID == 12 {
		n, e := u32(b, 12)
		if e != nil || uint64(n) > uint64((len(b)-16)/12) {
			return 0, fmt.Errorf("truncated cmap12")
		}
		i := sort.Search(int(n), func(i int) bool { v, _ := u32(b, 16+i*12+4); return v >= uint32(r) })
		if i == int(n) {
			return 0, nil
		}
		start, _ := u32(b, 16+i*12)
		if uint32(r) < start {
			return 0, nil
		}
		g, _ := u32(b, 16+i*12+8)
		idx := uint64(g) + uint64(uint32(r)-start)
		if idx >= uint64(f.numGlyphs) {
			return 0, fmt.Errorf("invalid cmap glyph")
		}
		return int(idx), nil
	}
	if r > 65535 {
		return 0, nil
	}
	nn, e := u16(b, 6)
	n := int(nn) / 2
	if e != nil || n == 0 || nn%2 != 0 || 16+n*8 > len(b) {
		return 0, fmt.Errorf("truncated cmap4")
	}
	i := sort.Search(n, func(i int) bool { end, _ := u16(b, 14+2*i); return end >= uint16(r) })
	if i == n {
		return 0, nil
	}
	start, _ := u16(b, 16+2*n+2*i)
	if uint16(r) < start {
		return 0, nil
	}
	delta, _ := u16(b, 16+4*n+2*i)
	rp := 16 + 6*n + 2*i
	ro, _ := u16(b, rp)
	var g uint16
	if ro == 0 {
		g = uint16(r) + delta
	} else {
		g, e = u16(b, rp+int(ro)+2*(int(r)-int(start)))
		if e != nil {
			return 0, e
		}
		if g != 0 {
			g += delta
		}
	}
	if int(g) >= f.numGlyphs {
		return 0, fmt.Errorf("cmap4 glyph outside range")
	}
	return int(g), nil
}
func (f *ttFace) glyphData(index int) ([]byte, error) {
	if index < 0 || index >= f.numGlyphs {
		return nil, fmt.Errorf("glyph index outside range")
	}
	loc := f.tables["loca"]
	var a, b uint32
	if f.locaFormat == 0 {
		aa, e := u16(loc, index*2)
		if e != nil {
			return nil, e
		}
		bb, e := u16(loc, (index+1)*2)
		if e != nil {
			return nil, e
		}
		a, b = uint32(aa)*2, uint32(bb)*2
	} else {
		var e error
		a, e = u32(loc, index*4)
		if e != nil {
			return nil, e
		}
		b, e = u32(loc, (index+1)*4)
		if e != nil {
			return nil, e
		}
	}
	glyf := f.tables["glyf"]
	if b < a || uint64(b) > uint64(len(glyf)) {
		return nil, fmt.Errorf("invalid glyph location")
	}
	return glyf[int(a):int(b)], nil
}

type ttReader struct {
	b   []byte
	p   int
	err error
}

func (r *ttReader) byte() byte {
	if r.p >= len(r.b) {
		r.err = fmt.Errorf("truncated glyph")
		return 0
	}
	b := r.b[r.p]
	r.p++
	return b
}
func (r *ttReader) word() uint16 { a, b := r.byte(), r.byte(); return uint16(a)<<8 | uint16(b) }
func (r *ttReader) signed() int  { return int(int16(r.word())) }
func (r *ttReader) skip(n int) {
	if n < 0 || n > len(r.b)-r.p {
		r.err = fmt.Errorf("truncated glyph instructions")
		r.p = len(r.b)
	} else {
		r.p += n
	}
}
func (f *ttFace) contours(index, depth int, budget *int) ([][]fontPoint, error) {
	if depth > 16 || *budget <= 0 {
		return nil, fmt.Errorf("composite glyph recursion/budget exceeded")
	}
	*budget--
	data, e := f.glyphData(index)
	if e != nil {
		return nil, e
	}
	if len(data) == 0 {
		return nil, nil
	}
	if len(data) < 10 {
		return nil, fmt.Errorf("truncated glyph header")
	}
	count := int(int16(binary.BigEndian.Uint16(data)))
	rd := ttReader{b: data, p: 10}
	if count >= 0 {
		if count == 0 {
			return nil, nil
		}
		if count > 4096 {
			return nil, fmt.Errorf("too many glyph contours")
		}
		ends := make([]int, count)
		prev := -1
		for i := range ends {
			ends[i] = int(rd.word())
			if ends[i] <= prev {
				return nil, fmt.Errorf("invalid glyph contour endpoints")
			}
			prev = ends[i]
		}
		n := ends[count-1] + 1
		if n > 65536 {
			return nil, fmt.Errorf("too many glyph points")
		}
		ilen := int(rd.word())
		rd.skip(ilen)
		flags := make([]byte, 0, n)
		for len(flags) < n && rd.err == nil {
			b := rd.byte()
			reps := 1
			if b&8 != 0 {
				reps += int(rd.byte())
			}
			if len(flags)+reps > n {
				return nil, fmt.Errorf("invalid glyph repeat count")
			}
			for k := 0; k < reps; k++ {
				flags = append(flags, b)
			}
		}
		if rd.err != nil {
			return nil, rd.err
		}
		ps := make([]fontPoint, n)
		x, y := 0, 0
		for i, b := range flags {
			d := 0
			if b&2 != 0 {
				d = int(rd.byte())
				if b&16 == 0 {
					d = -d
				}
			} else if b&16 == 0 {
				d = rd.signed()
			}
			x += d
			ps[i].x = float64(x)
			ps[i].on = b&1 != 0
		}
		for i, b := range flags {
			d := 0
			if b&4 != 0 {
				d = int(rd.byte())
				if b&32 == 0 {
					d = -d
				}
			} else if b&32 == 0 {
				d = rd.signed()
			}
			y += d
			ps[i].y = float64(y)
		}
		if rd.err != nil {
			return nil, rd.err
		}
		out := make([][]fontPoint, count)
		start := 0
		for i, end := range ends {
			out[i] = ps[start : end+1]
			start = end + 1
		}
		return out, nil
	}
	var out [][]fontPoint
	total := 0
	for components := 0; components < 1024; components++ {
		flags, child := rd.word(), int(rd.word())
		var a, b int
		if flags&1 != 0 {
			if flags&2 != 0 {
				a, b = rd.signed(), rd.signed()
			} else {
				a, b = int(rd.word()), int(rd.word())
			}
		} else {
			if flags&2 != 0 {
				a, b = int(int8(rd.byte())), int(int8(rd.byte()))
			} else {
				a, b = int(rd.byte()), int(rd.byte())
			}
		}
		xx, xy, yx, yy := 1., 0., 0., 1.
		f2 := func() float64 { return float64(int16(rd.word())) / 16384 }
		switch {
		case flags&8 != 0:
			xx = f2()
			yy = xx
		case flags&64 != 0:
			xx, yy = f2(), f2()
		case flags&128 != 0:
			xx, yx, xy, yy = f2(), f2(), f2(), f2()
		}
		if rd.err != nil {
			return nil, rd.err
		}
		cs, e := f.contours(child, depth+1, budget)
		if e != nil {
			return nil, e
		}
		for i := range cs {
			for j, p := range cs[i] {
				cs[i][j].x = xx*p.x + xy*p.y
				cs[i][j].y = yx*p.x + yy*p.y
			}
		}
		dx, dy := float64(a), float64(b)
		if flags&2 == 0 {
			lookup := func(contours [][]fontPoint, idx int) (fontPoint, bool) {
				for _, c := range contours {
					if idx < len(c) {
						return c[idx], true
					}
					idx -= len(c)
				}
				return fontPoint{}, false
			}
			p, ok1 := lookup(out, a)
			q, ok2 := lookup(cs, b)
			if !ok1 || !ok2 {
				return nil, fmt.Errorf("unsupported composite phantom-point attachment")
			}
			dx, dy = p.x-q.x, p.y-q.y
		} else if flags&0x0800 != 0 && flags&0x1000 == 0 {
			dx, dy = xx*dx+xy*dy, yx*dx+yy*dy
		}
		for i := range cs {
			total += len(cs[i])
			if total > 65536 {
				return nil, fmt.Errorf("composite point budget exceeded")
			}
			for j := range cs[i] {
				cs[i][j].x += dx
				cs[i][j].y += dy
			}
		}
		out = append(out, cs...)
		if flags&32 == 0 {
			if flags&256 != 0 {
				ilen := int(rd.word())
				rd.skip(ilen)
			}
			if rd.err != nil {
				return nil, rd.err
			}
			return out, nil
		}
	}
	return nil, fmt.Errorf("too many composite components")
}
func flattenContour(ps []fontPoint, scale float64) []point {
	if len(ps) == 0 {
		return nil
	}
	trans := func(p fontPoint) point { return point{p.x * scale, -p.y * scale} }
	// Insert implied on-curve midpoints between consecutive quadratic controls.
	seq := make([]fontPoint, 0, len(ps)*2)
	for i, p := range ps {
		seq = append(seq, p)
		q := ps[(i+1)%len(ps)]
		if !p.on && !q.on {
			seq = append(seq, fontPoint{(p.x + q.x) / 2, (p.y + q.y) / 2, true})
		}
	}
	first := -1
	for i, p := range seq {
		if p.on {
			first = i
			break
		}
	}
	if first < 0 {
		return nil
	}
	ordered := append(clone(seq[first:]), seq[:first]...)
	ordered = append(ordered, ordered[0])
	out := []point{trans(ordered[0])}
	for i := 1; i < len(ordered); {
		p := ordered[i]
		if p.on {
			out = append(out, trans(p))
			i++
			continue
		}
		if i+1 >= len(ordered) {
			break
		}
		a, b, c := out[len(out)-1], trans(p), trans(ordered[i+1])
		distance := math.Hypot(a.x-2*b.x+c.x, a.y-2*b.y+c.y)
		steps := min(128, max(2, int(math.Ceil(math.Sqrt(distance/0.08)))))
		for k := 1; k <= steps; k++ {
			t := float64(k) / float64(steps)
			u := 1 - t
			out = append(out, point{u*u*a.x + 2*u*t*b.x + t*t*c.x, u*u*a.y + 2*u*t*b.y + t*t*c.y})
		}
		i += 2
	}
	return out
}

type scanCross struct {
	x     float64
	delta int
}

func glyphMask(contours [][]point, box image.Rectangle) *image.Alpha {
	mask := image.NewAlpha(image.Rect(0, 0, box.Dx(), box.Dy()))
	const aa = 8
	w, h := box.Dx()*aa, box.Dy()*aa
	coverage := make([]uint16, box.Dx()*box.Dy())
	cross := []scanCross{}
	for sy := 0; sy < h; sy++ {
		fy := float64(box.Min.Y) + (float64(sy)+.5)/aa
		cross = cross[:0]
		for _, ps := range contours {
			if len(ps) < 2 {
				continue
			}
			prev := ps[len(ps)-1]
			for _, cur := range ps {
				delta := 0
				if prev.y <= fy && fy < cur.y {
					delta = 1
				} else if cur.y <= fy && fy < prev.y {
					delta = -1
				}
				if delta != 0 {
					cross = append(cross, scanCross{prev.x + (fy-prev.y)/(cur.y-prev.y)*(cur.x-prev.x), delta})
				}
				prev = cur
			}
		}
		sort.Slice(cross, func(i, j int) bool { return cross[i].x < cross[j].x })
		winding := 0
		last := 0.
		for _, c := range cross {
			if winding != 0 {
				a := max(0, int(math.Ceil((last-float64(box.Min.X))*aa-.5)))
				b := min(w, int(math.Ceil((c.x-float64(box.Min.X))*aa-.5)))
				for sx := a; sx < b; sx++ {
					coverage[(sy/aa)*box.Dx()+sx/aa]++
				}
			}
			winding += c.delta
			last = c.x
		}
	}
	for i, c := range coverage {
		mask.Pix[i] = uint8((uint32(c)*255 + 32) / 64)
	}
	return mask
}
func (f *ttFace) Glyph(r rune) (Glyph, error) {
	if g, ok := f.cache[r]; ok {
		return g, nil
	}
	idx, e := f.glyphIndex(r)
	if e != nil {
		return Glyph{}, e
	}
	metric := min(idx, f.numMetrics-1)
	adv, e := u16(f.tables["hmtx"], metric*4)
	if e != nil {
		return Glyph{}, e
	}
	scale := f.size / float64(f.upem)
	g := Glyph{Advance: float64(iround(float64(adv) * scale))}
	budget := 1024
	cs, e := f.contours(idx, 0, &budget)
	if e != nil {
		return g, e
	}
	if len(cs) == 0 {
		g.Mask = image.NewAlpha(image.Rect(0, 0, 0, 0))
		f.cache[r] = g
		return g, nil
	}
	var flats [][]point
	loX, loY, hiX, hiY := math.Inf(1), math.Inf(1), math.Inf(-1), math.Inf(-1)
	for _, c := range cs {
		flat := flattenContour(c, scale)
		for _, p := range flat {
			loX = math.Min(loX, p.x)
			loY = math.Min(loY, p.y)
			hiX = math.Max(hiX, p.x)
			hiY = math.Max(hiY, p.y)
		}
		flats = append(flats, flat)
	}
	if !finite(loX) || !finite(loY) || !finite(hiX) || !finite(hiY) || math.Max(math.Max(math.Abs(loX), math.Abs(loY)), math.Max(math.Abs(hiX), math.Abs(hiY))) > 32768 {
		return g, fmt.Errorf("invalid glyph bounds")
	}
	box := image.Rect(int(math.Floor(loX)), int(math.Floor(loY)), int(math.Ceil(hiX)), int(math.Ceil(hiY)))
	if box.Dx() > 4096 || box.Dy() > 4096 || int64(box.Dx())*int64(box.Dy()) > 1_000_000 {
		return g, fmt.Errorf("glyph allocation guard exceeded")
	}
	g.X, g.Y = box.Min.X, box.Min.Y
	g.Mask = glyphMask(flats, box)
	f.cache[r] = g
	return g, nil
}
