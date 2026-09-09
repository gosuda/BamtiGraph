// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fontPath(t testing.TB) string {
	t.Helper()
	if p := os.Getenv("BAMTIGRAPH_FONT"); p != "" {
		return p
	}
	for _, p := range monoCandidates {
		if _, e := os.Stat(p); e == nil {
			return p
		}
	}
	t.Fatal("font tests require an installed TrueType-outline font; set BAMTIGRAPH_FONT")
	return ""
}
func TestFontGlyphs(t *testing.T) {
	face, e := (TrueTypeBackend{}).Open(fontPath(t), 0, 14)
	if e != nil {
		t.Fatal(e)
	}
	defer face.Close()
	for _, r := range []rune{'A', 'g', '0', 'é', 'Å', ' ', '\U0010ffff'} {
		t.Run(string(r), func(t *testing.T) {
			g, e := face.Glyph(r)
			if e != nil {
				t.Fatal(e)
			}
			if !finite(g.Advance) || g.Advance <= 0 {
				t.Fatal("invalid advance")
			}
			again, e := face.Glyph(r)
			if e != nil || g.Mask != again.Mask {
				t.Fatal("glyph not cached")
			}
			if r != ' ' && g.Mask == nil {
				t.Fatal("missing mask")
			}
		})
	}
}
func TestFontErrors(t *testing.T) {
	path := fontPath(t)
	for _, sz := range []float64{0, -1, math.NaN(), math.Inf(1), 4096} {
		if _, e := (TrueTypeBackend{}).Open(path, 0, sz); e == nil {
			t.Fatal("invalid font size accepted")
		}
	}
	for _, index := range []int{-1, 100000} {
		if _, e := (TrueTypeBackend{}).Open(path, index, 12); e == nil {
			t.Fatal("invalid index")
		}
	}
	for _, c := range []FontConfig{{Strict: true}, {Mono: "/absent/font.ttf"}, {Mono: path, FaceIndex: -1}, {Mono: path, Caption: "/absent/font.ttf"}} {
		if _, e := FontEnvironment(c, DefaultTheme()); e == nil {
			t.Fatal("invalid font configuration")
		}
	}
}
func TestFontManifest(t *testing.T) {
	p := fontPath(t)
	e, err := FontEnvironment(FontConfig{Mono: p, Strict: true}, DefaultTheme())
	if err != nil {
		t.Fatal(err)
	}
	if len(e.Fonts) != 6 || e.FontBackend != (TrueTypeBackend{}).ID() {
		t.Fatal(e)
	}
	for _, fp := range e.Fonts {
		if strings.Contains(fp.Filename, "/") || len(fp.SHA256) != 64 {
			t.Fatal("bad fingerprint")
		}
	}
	if e.Fonts["caption"].SHA256 != e.Fonts["axis"].SHA256 {
		t.Fatal("strict font changed")
	}
}
func TestTruncatedFonts(t *testing.T) {
	data, e := os.ReadFile(fontPath(t))
	if e != nil {
		t.Fatal(e)
	}
	for _, n := range []int{0, 1, 11, 12, 20, 100, 1000} {
		if _, e := parseTrueType(data[:n], 0, 11); e == nil {
			t.Fatalf("truncated font %d accepted", n)
		}
	}
	cff := make([]byte, 16)
	copy(cff, "OTTO")
	if _, e := parseTrueType(cff, 0, 11); e == nil {
		t.Fatal("CFF accepted")
	}
	if _, e := parseTrueType([]byte("ttcf000000000000"), 1, 11); e == nil {
		t.Fatal("bad TTC accepted")
	}
}
func TestTrueTypeCollection(t *testing.T) {
	data, e := os.ReadFile(fontPath(t))
	if e != nil {
		t.Fatal(e)
	}
	if bytes.Equal(data[:4], []byte("ttcf")) {
		t.Skip("TTC supplied as seed")
	}
	n := int(binary.BigEndian.Uint16(data[4:6]))
	collection := make([]byte, len(data)+16)
	copy(collection, "ttcf")
	binary.BigEndian.PutUint32(collection[4:], 0x00010000)
	binary.BigEndian.PutUint32(collection[8:], 1)
	binary.BigEndian.PutUint32(collection[12:], 16)
	copy(collection[16:], data)
	for i := 0; i < n; i++ {
		p := 16 + 12 + i*16 + 8
		off := binary.BigEndian.Uint32(collection[p:])
		binary.BigEndian.PutUint32(collection[p:], off+16)
	}
	f, e := parseTrueType(collection, 0, 11)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	g, e := f.Glyph('A')
	if e != nil || g.Mask == nil {
		t.Fatal(g, e)
	}
	if _, e := parseTrueType(collection, 1, 11); e == nil {
		t.Fatal("missing TTC face accepted")
	}
}
func TestMutatedFontsDoNotPanic(t *testing.T) {
	seed, e := os.ReadFile(fontPath(t))
	if e != nil {
		t.Fatal(e)
	}
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 80; i++ {
		d := bytes.Clone(seed)
		for k := 0; k < 8; k++ {
			j := rng.Intn(len(d))
			d[j] = byte(rng.Intn(256))
		}
		f, e := parseTrueType(d, 0, 11)
		if e == nil {
			_, _ = f.Glyph('é')
			_ = f.Close()
		}
	}
}
func TestFontRotationAndWideCells(t *testing.T) {
	f, e := (TrueTypeBackend{}).Open(fontPath(t), 0, 11)
	if e != nil {
		t.Fatal(e)
	}
	defer f.Close()
	if cellAdvance('\uD55C', 6) != 12 || cellAdvance('A', 6) != 6 {
		t.Fatal("wide cells")
	}
	a, e := rotatedText("bits", f, RGB(0, 0, 0), false)
	if e != nil {
		t.Fatal(e)
	}
	b, e := rotatedText("bits", f, RGB(0, 0, 0), true)
	if e != nil {
		t.Fatal(e)
	}
	if a.Rect.Size() != b.Rect.Size() {
		t.Fatal("rotation bounds")
	}
}
func TestFontBackendDispatch(t *testing.T) {
	path := fontPath(t)
	be := &countedBackend{}
	g := testGraph(t)
	g.Fonts = FontConfig{Mono: path, Strict: true, Backend: be}
	r, e := g.RenderResult()
	if e != nil {
		t.Fatal(e)
	}
	if be.opens == 0 || r.Metadata.Environment.FontBackend != "custom/test" {
		t.Fatal("custom backend unused")
	}
}

type countedBackend struct{ opens int }

func (b *countedBackend) ID() string { return "custom/test" }
func (b *countedBackend) Open(p string, i int, s float64) (FontFace, error) {
	b.opens++
	return (TrueTypeBackend{}).Open(p, i, s)
}
func FuzzTrueTypeParser(f *testing.F) {
	f.Add([]byte{})
	f.Add([]byte("OTTO unsupported font"))
	f.Add([]byte("ttcf000000000000000000000"))
	if p := os.Getenv("BAMTIGRAPH_FONT"); p != "" {
		if data, e := os.ReadFile(p); e == nil {
			f.Add(data)
		}
	}
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		face, e := parseTrueType(data, 0, 11)
		if e == nil {
			defer face.Close()
			_, _ = face.Glyph('A')
			_, _ = face.Glyph('é')
		}
	})
}
func TestInvalidFontFile(t *testing.T) {
	p := filepath.Join(t.TempDir(), "bad.ttf")
	if e := os.WriteFile(p, []byte("not a font........"), 0600); e != nil {
		t.Fatal(e)
	}
	if _, e := (TrueTypeBackend{}).Open(p, 0, 12); e == nil {
		t.Fatal("invalid file")
	}
}
