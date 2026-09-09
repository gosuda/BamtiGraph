// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"math"
	"time"
)

// DemoTraffic creates synthetic, not measured, data. The named time preset only
// formats the axis; no implicit aggregation is performed.
func DemoTraffic(weekly bool) (*Graph, error) {
	n, step := 289, 300.
	start := Epoch(time.Date(2026, 9, 7, 12, 0, 0, 0, time.FixedZone("KST", 9*3600)))
	if weekly {
		n, step = 337, 1800
		start -= 6 * 86400
	}
	ts, a, b := make([]float64, n), make([]float64, n), make([]float64, n)
	for i := range ts {
		x := float64(i)
		ts[i] = start + x*step
		if weekly {
			phase := float64(i%48) / 48
			cycle := math.Max(.06, math.Sin(math.Pi*phase))
			a[i] = (25 + 180*cycle + 12*math.Sin(x*1.43) + 7*math.Sin(x*3.1)) * 1e6
			b[i] = (15 + 35*cycle + 3*math.Sin(x*1.1)) * 1e6
		} else {
			a[i] = (130 + 75*math.Sin(x/50) + 10*math.Sin(x*2.17) + 5*math.Cos(x*.73)) * 1e6
			b[i] = (42 + 15*math.Sin(x/43) + 3*math.Sin(x*1.3)) * 1e6
		}
	}
	g, e := Traffic(ts, a, b)
	if e != nil {
		return nil, e
	}
	g.TimeAxis = DailyAxis("Asia/Seoul")
	if weekly {
		g.TimeAxis = WeeklyAxis("Asia/Seoul")
	}
	g.YAxis.Maximum = Float(240e6)
	g.YAxis.MajorStep = Float(50e6)
	return g, nil
}
