// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

// bamtigraph renders traffic CSV files or deterministic demo data to PNG.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	_ "time/tzdata"

	rrd "github.com/gosuda/BamtiGraph"
)

type numberFlag struct {
	set bool
	v   float64
}

func (n *numberFlag) String() string {
	if !n.set {
		return "auto"
	}
	return strconv.FormatFloat(n.v, 'g', -1, 64)
}
func (n *numberFlag) Set(s string) error {
	v, e := strconv.ParseFloat(s, 64)
	if e != nil {
		return e
	}
	n.set = true
	n.v = v
	return nil
}
func (n *numberFlag) ptr() *float64 {
	if !n.set {
		return nil
	}
	return rrd.Float(n.v)
}

type columnFlag struct {
	columns *[]rrd.CSVColumn
	kind    rrd.Kind
}

func (c columnFlag) String() string { return "" }
func (c columnFlag) Set(s string) error {
	column, name, found := strings.Cut(s, "=")
	if column == "" {
		return fmt.Errorf("column name is empty")
	}
	if !found {
		name = column
	}
	*c.columns = append(*c.columns, rrd.CSVColumn{Column: column, Name: name, Kind: c.kind})
	return nil
}

func run(args []string, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet("bamtigraph", flag.ContinueOnError)
	fs.SetOutput(stderr)
	output := fs.String("o", "traffic.png", "output PNG path (parent directory must exist)")
	input := fs.String("input", "", "input CSV; use '-' for stdin")
	demo := fs.Bool("demo", false, "render deterministic synthetic data")
	weekly := fs.Bool("weekly", false, "use weekly synthetic demo data")
	mode := fs.String("mode", "auto", "auto|daily|weekly|monthly|yearly|custom")
	zone := fs.String("timezone", "UTC", "IANA timezone for labels")
	title := fs.String("title", "Traffic - ether1", "panel title")
	unit := fs.String("vertical-label", "bits per second", "vertical unit label")
	watermark := fs.String("watermark", "", "right-side text; empty disables")
	font := fs.String("font", "", "explicit installed TTF/TTC font path (strict)")
	width := fs.Int("width", 595, "full panel width")
	height := fs.Int("plot-height", 122, "plot height")
	scale := fs.Int("scale", 1, "nearest-neighbor integer pixel scale, 1..8")
	aa := fs.Int("antialias", 4, "data-path supersampling, 1..8")
	legend := fs.String("legend", "reference", "reference|aligned|none")
	base := fs.Int("base", 1000, "SI=1000, IEC=1024")
	dec := fs.Int("legend-decimals", 2, "legend decimals")
	metadata := fs.String("metadata", "", "optional manifest JSON path")
	timecol := fs.String("timestamp-column", "timestamp", "CSV timestamp header")
	gap := fs.Float64("gap-after", 0, "maximum elapsed seconds connected by a path")
	maxRows := fs.Int("max-rows", 2_000_000, "maximum CSV rows")
	version := fs.Bool("version", false, "print version")
	var ymax, ymin, ystep numberFlag
	var columns []rrd.CSVColumn
	fs.Var(&ymax, "y-max", "explicit raw-unit maximum")
	fs.Var(&ymin, "y-min", "explicit raw-unit minimum")
	fs.Var(&ystep, "y-step", "raw-unit major grid step")
	fs.Var(columnFlag{&columns, rrd.Area}, "area", "column=Label (repeatable)")
	fs.Var(columnFlag{&columns, rrd.Line}, "line", "column=Label (repeatable)")
	if e := fs.Parse(args); e != nil {
		if e == flag.ErrHelp {
			return 0
		}
		return 2
	}
	fail := func(e error) int { fmt.Fprintln(stderr, "bamtigraph:", e); return 2 }
	if *version {
		fmt.Fprintln(stdout, "bamtigraph-go", rrd.Version)
		return 0
	}
	if fs.NArg() > 1 {
		return fail(fmt.Errorf("only one positional CSV path is supported"))
	}
	if fs.NArg() == 1 {
		if *input != "" {
			return fail(fmt.Errorf("use either -input or positional CSV, not both"))
		}
		*input = fs.Arg(0)
	}
	if *demo && *input != "" {
		return fail(fmt.Errorf("-demo and CSV are mutually exclusive"))
	}
	var g *rrd.Graph
	var e error
	if *demo {
		g, e = rrd.DemoTraffic(*weekly)
		if e != nil {
			return fail(e)
		}
	} else {
		if *input == "" {
			return fail(fmt.Errorf("use -demo or -input traffic.csv; see -help"))
		}
		var r io.Reader = os.Stdin
		if *input != "-" {
			f, err := os.Open(*input)
			if err != nil {
				return fail(err)
			}
			defer f.Close()
			r = f
		}
		series, err := rrd.ReadCSV(r, rrd.CSVOptions{TimestampColumn: *timecol, Columns: columns, MaxRows: *maxRows})
		if err != nil {
			return fail(err)
		}
		g = rrd.NewGraph(series...)
	}
	g.Title, g.VerticalLabel, g.Watermark = *title, *unit, *watermark
	if *mode != "auto" || !*demo {
		g.TimeAxis.Mode = *mode
	}
	g.TimeAxis.Timezone = *zone
	g.Layout.Width, g.Layout.PlotHeight, g.Layout.PixelScale, g.Layout.Antialias, g.Layout.Legend = *width, *height, *scale, *aa, *legend
	g.YAxis.Base, g.YAxis.LegendDecimals = *base, *dec
	if ymin.set {
		g.YAxis.Minimum = ymin.ptr()
	}
	if ymax.set {
		g.YAxis.Maximum = ymax.ptr()
	}
	if ystep.set {
		g.YAxis.MajorStep = ystep.ptr()
	}
	if *font != "" {
		g.Fonts = rrd.FontConfig{Mono: *font, Strict: true}
	}
	for i := range g.Series {
		g.Series[i].GapAfter = *gap
	}
	samePath := func(a, b string) bool {
		aa, ea := filepath.Abs(a)
		bb, eb := filepath.Abs(b)
		if ea == nil && eb == nil && aa == bb {
			return true
		}
		sa, e1 := os.Stat(a)
		sb, e2 := os.Stat(b)
		return e1 == nil && e2 == nil && os.SameFile(sa, sb)
	}
	if *metadata != "" && samePath(*metadata, *output) {
		return fail(fmt.Errorf("PNG and metadata paths must differ"))
	}
	result, err := g.Save(*output)
	if err != nil {
		return fail(err)
	}
	if *metadata != "" {
		data, err := json.MarshalIndent(result.Metadata, "", "  ")
		if err != nil {
			return fail(err)
		}
		if err = os.WriteFile(*metadata, append(data, '\n'), 0644); err != nil {
			return fail(err)
		}
	}
	fmt.Fprintln(stdout, *output)
	for _, w := range result.Metadata.Environment.Warnings {
		fmt.Fprintln(stderr, "warning:", w)
	}
	return 0
}
func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }
