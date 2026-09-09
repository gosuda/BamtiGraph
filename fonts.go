// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"crypto/sha256"
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

var monoCandidates = []string{
	"/usr/share/fonts/truetype/dejavu/DejaVuSansMono.ttf",
	"/usr/share/fonts/dejavu-sans-mono-fonts/DejaVuSansMono.ttf",
	"/usr/local/share/fonts/DejaVuSansMono.ttf", "DejaVuSansMono.ttf",
	"/usr/share/fonts/truetype/liberation2/LiberationMono-Regular.ttf",
	"C:/Windows/Fonts/consola.ttf", "C:/Windows/Fonts/cour.ttf",
	"/System/Library/Fonts/Menlo.ttc", "/System/Library/Fonts/Monaco.ttf",
}
var captionCandidates = []string{"/usr/share/fonts/truetype/dejavu/DejaVuSans-Bold.ttf", "DejaVuSans-Bold.ttf", "C:/Windows/Fonts/arialbd.ttf", "/System/Library/Fonts/Supplemental/Arial Bold.ttf"}

type FontFingerprint struct {
	Filename  string  `json:"filename"`
	SHA256    string  `json:"sha256"`
	Size      float64 `json:"size"`
	FaceIndex int     `json:"face_index"`
}
type Environment struct {
	Go          string                     `json:"go"`
	GOOS        string                     `json:"goos"`
	GOARCH      string                     `json:"goarch"`
	FontBackend string                     `json:"font_backend"`
	Fonts       map[string]FontFingerprint `json:"fonts"`
	Warnings    []string                   `json:"warnings,omitempty"`
}
type fontManager struct {
	paths       map[string]string
	sizes       map[string]float64
	faces       map[string]FontFace
	backend     FontBackend
	index       int
	environment Environment
}

func newFonts(c FontConfig, t Theme) (*fontManager, error) {
	if c.FaceIndex < 0 || c.FaceIndex > 65535 || (c.Strict && c.Mono == "") {
		return nil, fmt.Errorf("invalid FontConfig: strict mode requires Mono; face index must be 0..65535")
	}
	if e := t.Validate(); e != nil {
		return nil, e
	}
	backend := c.Backend
	if backend == nil {
		backend = TrueTypeBackend{}
	}
	f := &fontManager{paths: map[string]string{}, sizes: map[string]float64{"axis": t.AxisSize, "legend": t.LegendSize, "watermark": t.WatermarkSize, "title": t.TitleSize, "unit": t.UnitSize, "caption": t.CaptionSize}, faces: map[string]FontFace{}, backend: backend, index: c.FaceIndex}
	f.environment = Environment{Go: runtime.Version(), GOOS: runtime.GOOS, GOARCH: runtime.GOARCH, FontBackend: backend.ID(), Fonts: map[string]FontFingerprint{}}
	resolve := func(explicit string, candidates []string) (string, error) {
		if explicit != "" {
			candidates = []string{explicit}
		}
		var last error
		for _, path := range candidates {
			face, e := backend.Open(path, c.FaceIndex, 11)
			if e == nil {
				_ = face.Close()
				return path, nil
			}
			last = e
		}
		if explicit != "" {
			return "", fmt.Errorf("open font %q: %w", explicit, last)
		}
		return "", fmt.Errorf("no usable system TrueType font; set FontConfig.Mono or BAMTIGRAPH_FONT (fonts are not bundled): %w", last)
	}
	monoName := c.Mono
	if monoName == "" {
		monoName = os.Getenv("BAMTIGRAPH_FONT")
	}
	mono, e := resolve(monoName, monoCandidates)
	if e != nil {
		return nil, e
	}
	if monoName == "" && !strings.Contains(filepath.Base(mono), "DejaVuSansMono") {
		f.environment.Warnings = append(f.environment.Warnings, "DejaVu Sans Mono unavailable; font substitution changes text pixels")
	}
	for _, role := range []string{"axis", "legend", "watermark"} {
		f.paths[role] = mono
	}
	f.paths["title"], e = resolve(c.Title, []string{mono})
	if e != nil {
		return nil, e
	}
	f.paths["unit"], e = resolve(c.Unit, []string{mono})
	if e != nil {
		return nil, e
	}
	if c.Strict && c.Caption == "" {
		f.paths["caption"] = mono
	} else {
		caption, err := resolve(c.Caption, captionCandidates)
		if err != nil {
			if c.Caption != "" {
				return nil, err
			}
			caption = mono
		}
		f.paths["caption"] = caption
	}
	hashes := map[string]string{}
	for role, path := range f.paths {
		hash, ok := hashes[path]
		if !ok {
			data, e := os.ReadFile(path)
			if e != nil {
				return nil, e
			}
			sum := sha256.Sum256(data)
			hash = fmt.Sprintf("%x", sum)
			hashes[path] = hash
		}
		f.environment.Fonts[role] = FontFingerprint{filepath.Base(path), hash, f.sizes[role], c.FaceIndex}
	}
	return f, nil
}
func (f *fontManager) get(role string, scale float64) (FontFace, error) {
	key := fmt.Sprintf("%s/%.12g", role, scale)
	if face, ok := f.faces[key]; ok {
		return face, nil
	}
	face, e := f.backend.Open(f.paths[role], f.index, f.sizes[role]*scale)
	if e != nil {
		return nil, e
	}
	f.faces[key] = face
	return face, nil
}
func (f *fontManager) close() {
	for _, face := range f.faces {
		_ = face.Close()
	}
}
func wideRune(r rune) bool {
	return r >= 0x1100 && (r <= 0x115f || r == 0x2329 || r == 0x232a || (r >= 0x2e80 && r <= 0xa4cf && r != 0x303f) || (r >= 0xac00 && r <= 0xd7a3) || (r >= 0xf900 && r <= 0xfaff) || (r >= 0xfe10 && r <= 0xfe19) || (r >= 0xfe30 && r <= 0xfe6f) || (r >= 0xff00 && r <= 0xff60) || (r >= 0xffe0 && r <= 0xffe6) || (r >= 0x1f300 && r <= 0x1faff) || (r >= 0x20000 && r <= 0x3fffd))
}
func cellAdvance(r rune, a float64) float64 {
	if wideRune(r) {
		return 2 * a
	}
	return a
}
func textWidth(text string, face FontFace, advance float64) (float64, error) {
	w := 0.
	for _, r := range text {
		if advance > 0 {
			w += cellAdvance(r, advance)
		} else {
			g, e := face.Glyph(r)
			if e != nil {
				return 0, e
			}
			w += g.Advance
		}
	}
	return w, nil
}
func fitText(text string, maxWidth float64, face FontFace, advance float64) (string, error) {
	w, e := textWidth(text, face, advance)
	if e != nil {
		return "", e
	}
	if w <= maxWidth {
		return text, nil
	}
	end := "..."
	ew, e := textWidth(end, face, advance)
	if e != nil {
		return "", e
	}
	if ew > maxWidth {
		return "", nil
	}
	rs := []rune(text)
	for len(rs) > 0 {
		w, e = textWidth(string(rs), face, advance)
		if e != nil {
			return "", e
		}
		if w+ew <= maxWidth {
			break
		}
		rs = rs[:len(rs)-1]
	}
	return string(rs) + end, nil
}
func textBounds(text string, face FontFace) (image.Rectangle, error) {
	x := 0.
	var box image.Rectangle
	set := false
	for _, r := range text {
		g, e := face.Glyph(r)
		if e != nil {
			return box, e
		}
		if g.Mask != nil && !g.Mask.Rect.Empty() {
			b := g.Mask.Bounds().Sub(g.Mask.Bounds().Min).Add(image.Pt(iround(x)+g.X, g.Y))
			if !set {
				box = b
				set = true
			} else {
				box = box.Union(b)
			}
		}
		x += g.Advance
	}
	return box, nil
}
func drawGlyph(im *image.NRGBA, x, y int, g Glyph, fill color.NRGBA) {
	if g.Mask == nil {
		return
	}
	b := g.Mask.Bounds()
	for sy := b.Min.Y; sy < b.Max.Y; sy++ {
		for sx := b.Min.X; sx < b.Max.X; sx++ {
			alpha := uint8((uint16(g.Mask.AlphaAt(sx, sy).A)*uint16(fill.A) + 127) / 255)
			if alpha > 0 {
				c := fill
				c.A = alpha
				blendPixel(im, x+g.X+sx-b.Min.X, y+g.Y+sy-b.Min.Y, c)
			}
		}
	}
}
func drawText(im *image.NRGBA, x, y float64, text string, face FontFace, fill color.NRGBA, advance float64, align string, centerY bool) error {
	width, e := textWidth(text, face, advance)
	if e != nil {
		return e
	}
	switch align {
	case "center":
		x -= width / 2
	case "right":
		x -= width
	}
	measure := text
	if measure == "" {
		measure = "0"
	}
	box, e := textBounds(measure, face)
	if e != nil {
		return e
	}
	if centerY {
		y -= float64(box.Dy()) / 2
	}
	baseline := y - float64(box.Min.Y)
	for _, r := range text {
		g, e := face.Glyph(r)
		if e != nil {
			return e
		}
		drawGlyph(im, iround(x), iround(baseline), g, fill)
		if advance > 0 {
			x += cellAdvance(r, advance)
		} else {
			x += g.Advance
		}
	}
	return nil
}
func rotatedText(text string, face FontFace, fill color.NRGBA, clockwise bool) (*image.NRGBA, error) {
	box, e := textBounds(text, face)
	if e != nil {
		return nil, e
	}
	if box.Dx() > 32768 || box.Dy() > 32768 || int64(max(1, box.Dx()+2))*int64(max(1, box.Dy()+2)) > MaxPixels {
		return nil, fmt.Errorf("rotated text exceeds allocation guard")
	}
	im := image.NewNRGBA(image.Rect(0, 0, max(1, box.Dx()+2), max(1, box.Dy()+2)))
	x, base := float64(1-box.Min.X), 1-box.Min.Y
	for _, r := range text {
		g, err := face.Glyph(r)
		if err != nil {
			return nil, err
		}
		drawGlyph(im, iround(x), base, g, fill)
		x += g.Advance
	}
	w, h := im.Rect.Dx(), im.Rect.Dy()
	out := image.NewNRGBA(image.Rect(0, 0, h, w))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if clockwise {
				out.SetNRGBA(h-1-y, x, im.NRGBAAt(x, y))
			} else {
				out.SetNRGBA(y, w-1-x, im.NRGBAAt(x, y))
			}
		}
	}
	return out, nil
}

// FontEnvironment records the font backend and SHA-256 hashes without exposing
// font data or absolute system paths.
func FontEnvironment(c FontConfig, t Theme) (Environment, error) {
	f, e := newFonts(c, t)
	if e != nil {
		return Environment{}, e
	}
	defer f.close()
	return f.environment, nil
}
