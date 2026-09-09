// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"image"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPixelDifference(t *testing.T) {
	a := solidImage(4, 3, RGB(10, 20, 30))
	b := solidImage(4, 3, RGB(10, 20, 30))
	b.SetNRGBA(2, 1, RGBA(13, 24, 30, 12))
	d, e := CompareImages(a, b, CompareOptions{Tolerance: 4})
	if e != nil {
		t.Fatal(e)
	}
	if d.TotalPixels != 12 || d.ExactPixels != 11 || d.WithinTolerancePixels != 12 || d.MaximumChannelError != 4 || d.MeanAbsoluteError != 7./36 || d.RootMeanSquareError != math.Sqrt(25./36) || *d.DifferenceBox != [4]int{2, 1, 3, 2} {
		t.Fatalf("%+v", d)
	}
	box := image.Rect(1, 1, 4, 3)
	d, e = CompareImages(a, b, CompareOptions{Box: &box})
	if e != nil || *d.DifferenceBox != [4]int{1, 0, 2, 1} {
		t.Fatal(d, e)
	}
	diff, e := DifferenceImage(a, b, 4)
	if e != nil || diff.NRGBAAt(2, 1) != RGB(12, 16, 0) {
		t.Fatal(diff, e)
	}
}
func TestPixelIgnoreAlphaAndOrigins(t *testing.T) {
	a := image.NewNRGBA(image.Rect(5, 6, 7, 8))
	b := image.NewNRGBA(image.Rect(0, 0, 2, 2))
	for y := 0; y < 2; y++ {
		for x := 0; x < 2; x++ {
			a.SetNRGBA(5+x, 6+y, RGBA(2, 3, 4, 0))
			b.SetNRGBA(x, y, RGBA(2, 3, 4, 255))
		}
	}
	d, e := CompareImages(a, b, CompareOptions{})
	if e != nil || d.ExactRatio != 1 || d.DifferenceBox != nil {
		t.Fatal(d, e)
	}
}
func TestCompareFailures(t *testing.T) {
	a := solidImage(2, 2, RGB(0, 0, 0))
	for _, o := range []CompareOptions{{Tolerance: -1}, {Tolerance: 256}, {Box: ptrRect(image.Rect(-1, 0, 1, 1))}, {Box: ptrRect(image.Rectangle{})}} {
		if _, e := CompareImages(a, a, o); e == nil {
			t.Fatal("invalid options")
		}
	}
	if _, e := CompareImages(a, solidImage(3, 3, RGB(0, 0, 0)), CompareOptions{}); e == nil {
		t.Fatal("different size accepted")
	}
	if _, e := CompareImages(nil, a, CompareOptions{}); e == nil {
		t.Fatal("nil accepted")
	}
	if _, e := DifferenceImage(a, a, 0); e == nil {
		t.Fatal("zero amplification")
	}
}
func ptrRect(r image.Rectangle) *image.Rectangle { return &r }
func TestCompareFiles(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "a.png")
	a := solidImage(2, 2, RGB(8, 9, 10))
	if e := SaveImagePNG(p, a); e != nil {
		t.Fatal(e)
	}
	d, e := CompareFiles(p, p, CompareOptions{})
	if e != nil || d.ExactRatio != 1 {
		t.Fatal(d, e)
	}
	bad := filepath.Join(dir, "bad")
	os.WriteFile(bad, []byte("not image"), 0600)
	if _, e := LoadImage(bad); e == nil {
		t.Fatal("invalid file accepted")
	}
}
func TestCSVTraffic(t *testing.T) {
	s, e := ReadCSV(strings.NewReader("\ufefftimestamp,inbound,outbound\n2026-09-07T12:00:00+09:00,100,1\n2026-09-07T12:05:00+09:00,,null\n2026-09-07T12:10:00+09:00,NaN,None\n"), DefaultCSVOptions())
	if e != nil {
		t.Fatal(e)
	}
	if len(s) != 2 || s[0].Kind != Area || s[1].Kind != Line || s[0].Values[0] != 100 || !math.IsNaN(s[0].Values[1]) || s[0].Timestamps[1]-s[0].Timestamps[0] != 300 {
		t.Fatal(s)
	}
}
func TestCSVCustom(t *testing.T) {
	o := DefaultCSVOptions()
	o.TimestampColumn = "t"
	o.Columns = []CSVColumn{{"cpu", "CPU", Area}, {"io", "Wait", Line}}
	s, e := ReadCSV(strings.NewReader("t,cpu,io\n0,10,2\n1,20,3\n"), o)
	if e != nil || len(s) != 2 || s[0].Name != "CPU" || s[1].Name != "Wait" {
		t.Fatal(s, e)
	}
}
func TestCSVErrors(t *testing.T) {
	for _, c := range []struct{ name, csv string }{
		{"empty", ""}, {"dupe_header", "timestamp,inbound,inbound\n"}, {"empty_header", "timestamp,,outbound\n"}, {"no_time", "x,inbound,outbound\n"}, {"no_series", "timestamp,x,y\n"}, {"bad_value", "timestamp,inbound,outbound\n0,Inf,0\n"}, {"bad_number", "timestamp,inbound,outbound\n0,nothing,0\n"}, {"naive_time", "timestamp,inbound,outbound\n2026-01-01T00:00:00,1,2\n"}, {"nan_time", "timestamp,inbound,outbound\nNaN,1,2\n"}, {"duplicate_time", "timestamp,inbound,outbound\n0,1,2\n0,2,3\n"}, {"ragged", "timestamp,inbound,outbound\n0,1\n"}} {
		t.Run(c.name, func(t *testing.T) {
			if _, e := ReadCSV(strings.NewReader(c.csv), DefaultCSVOptions()); e == nil {
				t.Fatal("invalid CSV accepted")
			}
		})
	}
	o := DefaultCSVOptions()
	o.MaxRows = 1
	if _, e := ReadCSV(strings.NewReader("timestamp,inbound,outbound\n0,1,2\n1,2,3\n"), o); e == nil {
		t.Fatal("row guard ignored")
	}
}
