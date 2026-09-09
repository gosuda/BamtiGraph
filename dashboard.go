// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"fmt"
	"image"
	"image/color"
)

type Panel struct {
	Graph   *Graph
	Caption string
}
type DashboardOptions struct {
	Gap int
	// Padding is left, top, right, bottom, in logical pixels.
	Padding    [4]int
	Background color.NRGBA
	// Zero preserves full height. Positive values explicitly crop logical pixels.
	CropHeight int
}

func DefaultDashboardOptions() DashboardOptions {
	return DashboardOptions{Gap: 24, Padding: [4]int{4, 3, 6, 6}, Background: RGB(243, 243, 243)}
}
func Dashboard(panels []Panel, options DashboardOptions) (*image.NRGBA, error) {
	if len(panels) == 0 || len(panels) > 128 {
		return nil, fmt.Errorf("dashboard requires 1..128 panels")
	}
	if options.Gap < 0 || options.Gap > 16384 || options.CropHeight < 0 {
		return nil, fmt.Errorf("invalid dashboard gap/crop")
	}
	for _, p := range options.Padding {
		if p < 0 || p > 16384 {
			return nil, fmt.Errorf("invalid padding")
		}
	}
	if panels[0].Graph == nil {
		return nil, fmt.Errorf("nil graph")
	}
	scale := panels[0].Graph.Layout.PixelScale
	width, total := 0, 0
	// Validate complete output size BEFORE rendering or allocating all panels.
	for _, p := range panels {
		if p.Graph == nil {
			return nil, fmt.Errorf("nil graph")
		}
		if e := p.Graph.Validate(); e != nil {
			return nil, e
		}
		if !oneLine(p.Caption) {
			return nil, fmt.Errorf("caption must be single-line")
		}
		if p.Graph.Layout.PixelScale != scale {
			return nil, fmt.Errorf("dashboard panels must share PixelScale")
		}
		width = max(width, p.Graph.Layout.Width*scale)
		total += p.Graph.Layout.Height(len(p.Graph.Series)) * scale
	}
	pl, pt, pr, pb := options.Padding[0]*scale, options.Padding[1]*scale, options.Padding[2]*scale, options.Padding[3]*scale
	gap := options.Gap * scale
	gaps := len(panels) - 1
	if panels[len(panels)-1].Caption != "" {
		gaps++
	}
	width += pl + pr
	total += pt + pb + gap*gaps
	if int64(width)*int64(total) > MaxPixels {
		return nil, fmt.Errorf("dashboard exceeds 40 million pixels")
	}
	if options.CropHeight > 0 && options.CropHeight*scale > total {
		return nil, fmt.Errorf("crop exceeds dashboard height")
	}
	canvas := solidImage(width, total, options.Background)
	y := pt
	for _, p := range panels {
		im, e := p.Graph.Render()
		if e != nil {
			return nil, e
		}
		composite(canvas, im, image.Pt(pl, y))
		if p.Caption != "" {
			f, e := newFonts(p.Graph.Fonts, p.Graph.Theme)
			if e != nil {
				return nil, e
			}
			face, e := f.get("caption", float64(scale))
			if e != nil {
				f.close()
				return nil, e
			}
			box, e := textBounds(p.Caption, face)
			if e != nil {
				f.close()
				return nil, e
			}
			tw, e := textWidth(p.Caption, face, 0)
			if e != nil {
				f.close()
				return nil, e
			}
			if box.Dy()+6*scale > gap || tw > float64(im.Rect.Dx()-4*scale) {
				f.close()
				return nil, fmt.Errorf("dashboard caption does not fit; increase gap/width or reduce caption size")
			}
			e = drawText(canvas, float64(pl)+float64(im.Rect.Dx())/2, float64(y+im.Rect.Dy()+4*scale), p.Caption, face, p.Graph.Theme.Text, 0, "center", false)
			f.close()
			if e != nil {
				return nil, e
			}
		}
		y += im.Rect.Dy() + gap
	}
	if options.CropHeight > 0 {
		h := options.CropHeight * scale
		crop := image.NewNRGBA(image.Rect(0, 0, width, h))
		copy(crop.Pix, canvas.Pix[:len(crop.Pix)])
		canvas = crop
	}
	return canvas, nil
}
