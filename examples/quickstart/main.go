package main

import (
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"time"
	_ "time/tzdata"

	rrd "github.com/gosuda/BamtiGraph"
)

func main() {
	output := flag.String("o", "examples/output/quickstart.png", "output PNG")
	font := flag.String("font", "", "path to an installed TrueType-outline TTF/TTC font")
	watermark := flag.String("watermark", "", "optional custom text in the right-hand margin")
	flag.Parse()
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Fatal(err)
	}
	start := time.Date(2026, 9, 7, 12, 0, 0, 0, loc)
	ts, inbound, outbound := make([]float64, 289), make([]float64, 289), make([]float64, 289)
	for i := range ts {
		x := float64(i)
		ts[i] = rrd.Epoch(start.Add(time.Duration(i) * 5 * time.Minute))
		inbound[i] = (130 + 75*math.Sin(x/50) + 10*math.Sin(x*2.17) + 5*math.Cos(x*.73)) * 1e6
		outbound[i] = (42 + 15*math.Sin(x/43) + 3*math.Sin(x*1.3)) * 1e6
	}
	graph, err := rrd.Traffic(ts, inbound, outbound)
	if err != nil {
		log.Fatal(err)
	}
	graph.TimeAxis = rrd.DailyAxis("Asia/Seoul")
	graph.YAxis.Maximum = rrd.Float(240e6)
	graph.YAxis.MajorStep = rrd.Float(50e6)
	graph.Watermark = *watermark
	if *font != "" {
		graph.Fonts = rrd.FontConfig{Mono: *font, Strict: true}
	}
	if err = os.MkdirAll(filepath.Dir(*output), 0755); err != nil {
		log.Fatal(err)
	}
	result, err := graph.Save(*output)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(*output)
	for _, s := range result.Metadata.Statistics {
		fmt.Printf("%s: samples=%d missing=%d\n", s.Name, s.Count, s.Missing)
	}
}
