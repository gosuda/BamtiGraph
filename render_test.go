// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"errors"
	"hash/crc32"
	"image"
	"image/png"
	"math"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
)

func testGraph(t *testing.T) *Graph {
	t.Helper()
	g, e := DemoTraffic(false)
	if e != nil {
		t.Fatal(e)
	}
	return g
}
func TestRenderDeterministic(t *testing.T) {
	g := testGraph(t)
	a, e := g.PNGBytes()
	if e != nil {
		t.Fatal(e)
	}
	b, e := g.PNGBytes()
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(a, b) {
		t.Fatal("repeat PNG is not deterministic")
	}
}

func TestWatermarkIsOptInAndRightAligned(t *testing.T) {
	g := testGraph(t)
	plain, err := g.RenderResult()
	if err != nil {
		t.Fatal(err)
	}
	if plain.Metadata.Watermark != "" {
		t.Fatal("default render must be white-label")
	}
	g.Watermark = "BAMTIGRAPH"
	marked, err := g.RenderResult()
	if err != nil {
		t.Fatal(err)
	}
	if marked.Metadata.Watermark != g.Watermark {
		t.Fatal("custom watermark absent from export metadata")
	}
	difference, err := CompareImages(plain.Image, marked.Image, CompareOptions{})
	if err != nil {
		t.Fatal(err)
	}
	if difference.DifferenceBox == nil {
		t.Fatal("custom watermark did not draw any pixels")
	}
	box := *difference.DifferenceBox
	if box[0] <= plain.Metadata.PlotBox[2] || box[2] >= g.Layout.Width-2 || box[1] < 4 || box[3] > plain.Image.Rect.Dy()-4 {
		t.Fatalf("watermark changed pixels outside the right margin: %v", box)
	}
	g.Watermark = ""
	cleared, err := g.Render()
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(plain.Image.Pix, cleared.Pix) {
		t.Fatal("clearing watermark did not restore unmarked pixels")
	}
}
func TestExactIntegerScale(t *testing.T) {
	g := testGraph(t)
	a, e := g.Render()
	if e != nil {
		t.Fatal(e)
	}
	g.Layout.PixelScale = 2
	b, e := g.Render()
	if e != nil {
		t.Fatal(e)
	}
	if b.Bounds().Size() != a.Bounds().Size().Mul(2) {
		t.Fatal("dimensions")
	}
	for y := 0; y < b.Rect.Dy(); y++ {
		for x := 0; x < b.Rect.Dx(); x++ {
			if b.NRGBAAt(x, y) != a.NRGBAAt(x/2, y/2) {
				t.Fatalf("changed pixel %d,%d", x, y)
			}
		}
	}
}
func TestRenderConcurrent(t *testing.T) {
	g := testGraph(t)
	want, e := g.PNGBytes()
	if e != nil {
		t.Fatal(e)
	}
	var wg sync.WaitGroup
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := g.PNGBytes()
			if e != nil {
				t.Error(e)
				return
			}
			if !bytes.Equal(got, want) {
				t.Error("concurrent image differs")
			}
		}()
	}
	wg.Wait()
}
func TestVisibleOverrideIsSeparate(t *testing.T) {
	g := testGraph(t)
	g.Series[0].LegendValues = &LegendValues{1, 2, 3}
	r, e := g.RenderResult()
	if e != nil {
		t.Fatal(e)
	}
	s := r.Metadata.Statistics[0]
	if s.DisplayOverride == nil || *s.DisplayOverride.Current != 1 || *s.Current == 1 {
		t.Fatal("override hid real statistics")
	}
}
func TestGridFrontAndLegendModes(t *testing.T) {
	for _, mode := range []string{"reference", "aligned", "none"} {
		for _, front := range []bool{true, false} {
			t.Run(mode+map[bool]string{true: "_front", false: "_back"}[front], func(t *testing.T) {
				g := testGraph(t)
				g.Layout.Legend = mode
				g.Theme.GridFront = front
				r, e := g.RenderResult()
				if e != nil {
					t.Fatal(e)
				}
				if mode == "none" && r.Image.Rect.Dy() != 174 {
					t.Fatal("hidden legend size")
				}
				if r.Image.NRGBAAt(0, 0) != g.Theme.ShadeLight {
					t.Fatal("top bevel")
				}
				for i := 3; i < len(r.Image.Pix); i += 4 {
					if r.Image.Pix[i] != 255 {
						t.Fatal("output is not opaque")
					}
				}
			})
		}
	}
}
func TestRulesAreClipped(t *testing.T) {
	g := testGraph(t)
	g.HRules = []HRule{HorizontalRule(100e6), HorizontalRule(-100e6), HorizontalRule(1e12)}
	g.VRules = []VRule{VerticalRule(g.Series[0].Timestamps[10]), VerticalRule(-100)}
	if _, e := g.Render(); e != nil {
		t.Fatal(e)
	}
}
func TestTextFitsOrErrors(t *testing.T) {
	for _, c := range []struct {
		name string
		edit func(*Graph)
		bad  bool
	}{
		{"title_ellipsized", func(g *Graph) { g.Title = strings.Repeat("Very long title ", 100) }, false},
		{"name_ellipsized", func(g *Graph) { g.Series[0].Name = strings.Repeat("Name", 100) }, false},
		{"unit_overflow", func(g *Graph) { g.VerticalLabel = strings.Repeat("unit", 100) }, true},
		{"watermark_overflow", func(g *Graph) { g.Watermark = strings.Repeat("watermark", 100) }, true},
		{"statistic_overflow", func(g *Graph) { g.YAxis.ScaleFactor = Float(1); g.YAxis.LegendDecimals = 12 }, true},
		{"small_panel", func(g *Graph) { g.Layout.Width = 400 }, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			g := testGraph(t)
			c.edit(g)
			_, e := g.Render()
			if (e != nil) != c.bad {
				t.Fatalf("error=%v, expected failure %v", e, c.bad)
			}
		})
	}
}
func TestPNGMetadataChunk(t *testing.T) {
	g := testGraph(t)
	r, e := g.RenderResult()
	if e != nil {
		t.Fatal(e)
	}
	r.Metadata.Title = "Unicode: \uD2B8\uB798\uD53D \U0001F98A"
	data, e := r.PNGBytes()
	if e != nil {
		t.Fatal(e)
	}
	if _, e = png.Decode(bytes.NewReader(data)); e != nil {
		t.Fatal(e)
	}
	found := false
	for p := 8; p < len(data); {
		if p+12 > len(data) {
			t.Fatal("truncated chunk")
		}
		n := int(binary.BigEndian.Uint32(data[p:]))
		if p+12+n > len(data) {
			t.Fatal("bad length")
		}
		typ := string(data[p+4 : p+8])
		payload := data[p+8 : p+8+n]
		want := binary.BigEndian.Uint32(data[p+8+n:])
		if crc32.ChecksumIEEE(data[p+4:p+8+n]) != want {
			t.Fatal("bad CRC")
		}
		if typ == "tEXt" {
			found = true
			if !bytes.HasPrefix(payload, []byte("chart\x00")) {
				t.Fatal("keyword")
			}
			for _, x := range payload {
				if x > 127 {
					t.Fatal("non-ASCII PNG metadata")
				}
			}
			var meta Metadata
			if e = json.Unmarshal(payload[len("chart\x00"):], &meta); e != nil {
				t.Fatal(e)
			}
			if meta.Title != r.Metadata.Title {
				t.Fatal("Unicode did not roundtrip")
			}
		}
		p += 12 + n
	}
	if !found {
		t.Fatal("no manifest")
	}
	var plain bytes.Buffer
	if e = r.EncodePNG(&plain, false); e != nil {
		t.Fatal(e)
	}
	if bytes.Contains(plain.Bytes(), []byte("chart\x00")) {
		t.Fatal("metadata not disabled")
	}
}

type failingWriter struct{}

func (failingWriter) Write([]byte) (int, error) { return 0, errors.New("write failed") }
func TestPNGWriteErrors(t *testing.T) {
	g := testGraph(t)
	r, e := g.RenderResult()
	if e != nil {
		t.Fatal(e)
	}
	if e = r.EncodePNG(failingWriter{}, true); e == nil {
		t.Fatal("write error lost")
	}
	var nilResult *RenderResult
	if _, e = nilResult.PNGBytes(); e == nil {
		t.Fatal("nil accepted")
	}
	r.Metadata.YStep = math.NaN()
	if _, e = r.PNGBytes(); e == nil {
		t.Fatal("invalid JSON accepted")
	}
}
func TestAtomicSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "traffic.png")
	original := []byte("keep me")
	if e := os.WriteFile(path, original, 0600); e != nil {
		t.Fatal(e)
	}
	g := testGraph(t)
	g.Layout.Width = 1
	if _, e := g.Save(path); e == nil {
		t.Fatal("invalid graph accepted")
	}
	got, _ := os.ReadFile(path)
	if !bytes.Equal(got, original) {
		t.Fatal("original changed")
	}
	g.Layout.Width = 595
	if _, e := g.Save(path); e != nil {
		t.Fatal(e)
	}
	if _, e := LoadImage(path); e != nil {
		t.Fatal(e)
	}
	if _, e := g.Save(filepath.Join(dir, "absent", "x.png")); e == nil {
		t.Fatal("created missing parent")
	}
	list, e := os.ReadDir(dir)
	if e != nil || len(list) != 1 {
		t.Fatal("temporary files leaked")
	}
}
func TestDashboard(t *testing.T) {
	g := testGraph(t)
	o := DefaultDashboardOptions()
	im, e := Dashboard([]Panel{{g, "Daily"}, {g, "Weekly"}}, o)
	if e != nil {
		t.Fatal(e)
	}
	if im.Rect.Size() != image.Pt(605, 479) {
		t.Fatal(im.Rect)
	}
	o.CropHeight = 455
	crop, e := Dashboard([]Panel{{g, "Daily"}, {g, "Weekly"}}, o)
	if e != nil {
		t.Fatal(e)
	}
	if crop.Rect.Dy() != 455 || !bytes.Equal(crop.Pix, im.Pix[:len(crop.Pix)]) {
		t.Fatal("crop resampled")
	}
}
func TestDashboardFailures(t *testing.T) {
	for _, c := range []string{"empty", "nil", "mixedscale", "gap", "crop", "caption", "padding"} {
		t.Run(c, func(t *testing.T) {
			g, h := testGraph(t), testGraph(t)
			ps := []Panel{{g, "Caption"}, {h, ""}}
			o := DefaultDashboardOptions()
			switch c {
			case "empty":
				ps = nil
			case "nil":
				ps[0].Graph = nil
			case "mixedscale":
				h.Layout.PixelScale = 2
			case "gap":
				o.Gap = 0
			case "crop":
				o.CropHeight = 10000
			case "caption":
				ps[0].Caption = "line\nbreak"
			case "padding":
				o.Padding[0] = -1
			}
			if _, e := Dashboard(ps, o); e == nil {
				t.Fatal("invalid dashboard accepted")
			}
		})
	}
}
func TestGoldenPixels(t *testing.T) {
	g := testGraph(t)
	// Reference pixels come from the supplied native renderer with this explicit text.
	g.Watermark = "BAMTIGRAPH"
	r, e := g.RenderResult()
	if e != nil {
		t.Fatal(e)
	}
	data, e := os.ReadFile("testdata/go_golden.json")
	if e != nil {
		t.Fatal(e)
	}
	var golden struct {
		Width      int                        `json:"width"`
		Height     int                        `json:"height"`
		RGBASHA256 string                     `json:"rgba_sha256"`
		Fonts      map[string]FontFingerprint `json:"fonts"`
	}
	if e = json.Unmarshal(data, &golden); e != nil {
		t.Fatal(e)
	}
	if !reflect.DeepEqual(golden.Fonts, r.Metadata.Environment.Fonts) {
		t.Skip("golden font fingerprint differs")
	}
	bounds := r.Image.Bounds()
	if bounds.Dx() != golden.Width || bounds.Dy() != golden.Height {
		t.Fatalf("golden geometry changed: got %dx%d, want %dx%d", bounds.Dx(), bounds.Dy(), golden.Width, golden.Height)
	}
	// Hash every straight RGBA channel in row-major order, excluding stride padding.
	h := sha256.New()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		offset := r.Image.PixOffset(bounds.Min.X, y)
		h.Write(r.Image.Pix[offset : offset+4*bounds.Dx()])
	}
	if actual := hex.EncodeToString(h.Sum(nil)); actual != golden.RGBASHA256 {
		t.Fatalf("golden pixels changed: got SHA-256 %s, want %s", actual, golden.RGBASHA256)
	}
}
