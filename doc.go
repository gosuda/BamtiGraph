// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

// Package bamtigraph renders classic RRD-style time-series charts using explicit
// pixel geometry. It is a native Go time-series library, not an RRDtool binding.
//
// Construct graphs with NewGraph or Traffic. Start configuration from the
// Default* constructors: configuration zero values are not implicit defaults.
// Values are raw units; math.NaN marks missing data. Epoch timestamps are seconds
// and must be finite and strictly increasing. No implicit sorting, resampling,
// unit conversion, counter interpretation, or extrapolation is performed.
//
// New graphs are white-label: Watermark is empty. Set Graph.Watermark to draw
// caller-provided text vertically at the existing right-side margin, or leave
// it empty to omit the mark entirely. PNG manifests use the generic "chart" key.
//
// The built-in font backend reads caller-installed TrueType-outline TTF/TTC
// fonts. It does not execute TrueType hinting or perform complex-text shaping.
// No font files, third-party modules, Python interpreter, or C libraries are
// required in the distribution. Font files must be supplied by the caller.
//
// Pixel layouts are deterministic. Byte-identical output to Pillow/FreeType is
// NOT promised: the native rasterizers differ. Pixel comparison utilities and
// reproducible fixtures quantify the difference instead of concealing it.
//
// Concurrent Render calls are safe while the graph and its input slices are not
// mutated. Faces and raster buffers are owned by each render, not shared globally.
package bamtigraph
