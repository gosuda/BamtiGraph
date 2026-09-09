// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"image"
	"image/color"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"
)

type PixelDifference struct {
	TotalPixels           int     `json:"total_pixels"`
	ExactPixels           int     `json:"exact_pixels"`
	WithinTolerancePixels int     `json:"within_tolerance_pixels"`
	MeanAbsoluteError     float64 `json:"mean_absolute_error"`
	RootMeanSquareError   float64 `json:"root_mean_square_error"`
	MaximumChannelError   int     `json:"maximum_channel_error"`
	DifferenceBox         *[4]int `json:"difference_box"`
	ExactRatio            float64 `json:"exact_ratio"`
	ToleranceRatio        float64 `json:"tolerance_ratio"`
}
type CompareOptions struct {
	Tolerance int
	Box       *image.Rectangle
}

func comparisonBox(a, b image.Image, o CompareOptions) (image.Rectangle, error) {
	if a == nil || b == nil {
		return image.Rectangle{}, fmt.Errorf("nil comparison image")
	}
	if a.Bounds().Size() != b.Bounds().Size() {
		return image.Rectangle{}, fmt.Errorf("image sizes differ: %v != %v", a.Bounds().Size(), b.Bounds().Size())
	}
	if o.Tolerance < 0 || o.Tolerance > 255 {
		return image.Rectangle{}, fmt.Errorf("tolerance must be 0..255")
	}
	r := image.Rectangle{Max: a.Bounds().Size()}
	if o.Box != nil {
		if o.Box.Empty() || !o.Box.In(r) {
			return r, fmt.Errorf("comparison box outside image")
		}
		r = *o.Box
	}
	if r.Empty() || int64(r.Dx())*int64(r.Dy()) > MaxPixels {
		return r, fmt.Errorf("comparison exceeds pixel guard or is empty")
	}
	return r, nil
}
func rgbAt(im image.Image, x, y int) color.NRGBA {
	return color.NRGBAModel.Convert(im.At(x+im.Bounds().Min.X, y+im.Bounds().Min.Y)).(color.NRGBA)
}
func CompareImages(reference, actual image.Image, o CompareOptions) (PixelDifference, error) {
	box, e := comparisonBox(reference, actual, o)
	if e != nil {
		return PixelDifference{}, e
	}
	d := PixelDifference{TotalPixels: box.Dx() * box.Dy()}
	var total, squares uint64
	var diff image.Rectangle
	set := false
	for y := box.Min.Y; y < box.Max.Y; y++ {
		for x := box.Min.X; x < box.Max.X; x++ {
			a, b := rgbAt(reference, x, y), rgbAt(actual, x, y)
			mx := 0
			for _, v := range []int{absInt(int(a.R) - int(b.R)), absInt(int(a.G) - int(b.G)), absInt(int(a.B) - int(b.B))} {
				mx = max(mx, v)
				total += uint64(v)
				squares += uint64(v * v)
			}
			d.MaximumChannelError = max(d.MaximumChannelError, mx)
			if mx == 0 {
				d.ExactPixels++
			} else {
				p := image.Rect(x-box.Min.X, y-box.Min.Y, x-box.Min.X+1, y-box.Min.Y+1)
				if !set {
					diff = p
					set = true
				} else {
					diff = diff.Union(p)
				}
			}
			if mx <= o.Tolerance {
				d.WithinTolerancePixels++
			}
		}
	}
	if set {
		d.DifferenceBox = &[4]int{diff.Min.X, diff.Min.Y, diff.Max.X, diff.Max.Y}
	}
	d.MeanAbsoluteError = float64(total) / float64(d.TotalPixels*3)
	d.RootMeanSquareError = math.Sqrt(float64(squares) / float64(d.TotalPixels*3))
	d.ExactRatio = float64(d.ExactPixels) / float64(d.TotalPixels)
	d.ToleranceRatio = float64(d.WithinTolerancePixels) / float64(d.TotalPixels)
	return d, nil
}
func DifferenceImage(reference, actual image.Image, amplify float64) (*image.NRGBA, error) {
	if !finite(amplify) || amplify <= 0 {
		return nil, fmt.Errorf("amplify must be positive and finite")
	}
	box, e := comparisonBox(reference, actual, CompareOptions{})
	if e != nil {
		return nil, e
	}
	out := image.NewNRGBA(box)
	ch := func(a, b uint8) uint8 {
		return uint8(math.Min(255, math.RoundToEven(float64(absInt(int(a)-int(b)))*amplify)))
	}
	for y := 0; y < box.Dy(); y++ {
		for x := 0; x < box.Dx(); x++ {
			a, b := rgbAt(reference, x, y), rgbAt(actual, x, y)
			out.SetNRGBA(x, y, RGB(ch(a.R, b.R), ch(a.G, b.G), ch(a.B, b.B)))
		}
	}
	return out, nil
}

// LoadImage checks dimensions before decoding to avoid accidental huge allocation.
func LoadImage(path string) (image.Image, error) {
	f, e := os.Open(path)
	if e != nil {
		return nil, e
	}
	defer f.Close()
	cfg, _, e := image.DecodeConfig(f)
	if e != nil {
		return nil, e
	}
	if cfg.Width <= 0 || cfg.Height <= 0 || int64(cfg.Width) > MaxPixels/int64(cfg.Height) {
		return nil, fmt.Errorf("image exceeds 40 million pixels")
	}
	if _, e = f.Seek(0, 0); e != nil {
		return nil, e
	}
	im, _, e := image.Decode(f)
	return im, e
}
func CompareFiles(reference, actual string, o CompareOptions) (PixelDifference, error) {
	a, e := LoadImage(reference)
	if e != nil {
		return PixelDifference{}, e
	}
	b, e := LoadImage(actual)
	if e != nil {
		return PixelDifference{}, e
	}
	return CompareImages(a, b, o)
}
