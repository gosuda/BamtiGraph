// Package fixture loads the shared Python/Go compatibility data. It is test and
// example support, not an alternate renderer or a source of embedded font data.
package fixture

import (
	"encoding/json"
	rrd "github.com/gosuda/BamtiGraph"
	"math"
	"os"
)

type Case struct {
	Name          string      `json:"name"`
	Timestamps    []float64   `json:"timestamps"`
	Inbound       []*float64  `json:"inbound"`
	Outbound      []*float64  `json:"outbound"`
	Mode          string      `json:"mode"`
	Zone          string      `json:"zone"`
	Start         *float64    `json:"start"`
	End           *float64    `json:"end"`
	YMin          *float64    `json:"y_min"`
	YMax          *float64    `json:"y_max"`
	YStep         *float64    `json:"y_step"`
	Base          int         `json:"base"`
	Factor        *float64    `json:"factor"`
	Suffix        *string     `json:"suffix"`
	Minor         *float64    `json:"minor_seconds"`
	Major         *float64    `json:"major_seconds"`
	Label         *float64    `json:"label_seconds"`
	Gap           float64     `json:"gap_after"`
	Interpolation string      `json:"interpolation"`
	Expected      Expectation `json:"expected"`
}

func Load(path string) ([]Case, error) {
	data, e := os.ReadFile(path)
	if e != nil {
		return nil, e
	}
	var cases []Case
	e = json.Unmarshal(data, &cases)
	return cases, e
}
func (c Case) Graph() (*rrd.Graph, error) {
	conv := func(in []*float64) []float64 {
		out := make([]float64, len(in))
		for i, p := range in {
			out[i] = math.NaN()
			if p != nil {
				out[i] = *p
			}
		}
		return out
	}
	g, e := rrd.Traffic(c.Timestamps, conv(c.Inbound), conv(c.Outbound))
	if e != nil {
		return nil, e
	}
	a := rrd.DefaultTimeAxis()
	a.Mode, a.Timezone, a.Start, a.End = c.Mode, c.Zone, c.Start, c.End
	a.MinorSeconds, a.MajorSeconds, a.LabelSeconds = c.Minor, c.Major, c.Label
	g.TimeAxis = a
	g.YAxis.Minimum, g.YAxis.Maximum, g.YAxis.MajorStep, g.YAxis.Base, g.YAxis.ScaleFactor, g.YAxis.Suffix = c.YMin, c.YMax, c.YStep, c.Base, c.Factor, c.Suffix
	for i := range g.Series {
		g.Series[i].GapAfter = c.Gap
		g.Series[i].Interpolation = rrd.Interpolation(c.Interpolation)
	}
	return g, nil
}

type Expectation struct {
	ImageSize   [2]int                 `json:"image_size"`
	LogicalSize [2]int                 `json:"logical_size"`
	PlotBox     [4]int                 `json:"plot_box"`
	TimeRange   [2]float64             `json:"time_range"`
	YRange      [2]float64             `json:"y_range"`
	YStep       float64                `json:"y_step"`
	YUnit       rrd.Unit               `json:"y_unit"`
	XLabels     []rrd.DrawnTick        `json:"x_labels"`
	Statistics  []rrd.SeriesStatistics `json:"statistics"`
}
