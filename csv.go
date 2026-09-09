// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"encoding/csv"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"
	"time"
)

type CSVColumn struct {
	Column, Name string
	Kind         Kind
}
type CSVOptions struct {
	TimestampColumn string
	Columns         []CSVColumn
	MaxRows         int
}

func DefaultCSVOptions() CSVOptions {
	return CSVOptions{TimestampColumn: "timestamp", MaxRows: 2_000_000}
}

// ReadCSV accepts numeric epoch seconds or RFC3339/RFC3339Nano timestamps.
// Blank/NaN/None/null cells are gaps. Duplicate headers or timestamps are errors.
func ReadCSV(r io.Reader, o CSVOptions) ([]Series, error) {
	if o.TimestampColumn == "" || o.MaxRows < 1 || o.MaxRows > 10_000_000 {
		return nil, fmt.Errorf("invalid CSV options")
	}
	reader := csv.NewReader(r)
	header, e := reader.Read()
	if e != nil {
		return nil, fmt.Errorf("CSV header: %w", e)
	}
	names := map[string]int{}
	for i, n := range header {
		n = strings.TrimSpace(strings.TrimPrefix(n, "\ufeff"))
		if n == "" {
			return nil, fmt.Errorf("empty CSV header")
		}
		if _, ok := names[n]; ok {
			return nil, fmt.Errorf("duplicate CSV column %q", n)
		}
		names[n] = i
	}
	ti, ok := names[o.TimestampColumn]
	if !ok {
		return nil, fmt.Errorf("timestamp column %q missing", o.TimestampColumn)
	}
	cols := clone(o.Columns)
	trafficMode := len(cols) == 0
	if trafficMode {
		cols = []CSVColumn{{"inbound", "Inbound", Area}, {"outbound", "Outbound", Line}}
	}
	if len(cols) > 128 {
		return nil, fmt.Errorf("too many CSV series")
	}
	idx := make([]int, len(cols))
	vals := make([][]float64, len(cols))
	for i, c := range cols {
		index, ok := names[c.Column]
		if !ok {
			return nil, fmt.Errorf("CSV column %q missing", c.Column)
		}
		if c.Kind != Line && c.Kind != Area {
			return nil, fmt.Errorf("invalid CSV series kind")
		}
		idx[i] = index
	}
	var ts []float64
	for row := 2; ; row++ {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("CSV row %d: %w", row, err)
		}
		if len(ts) >= o.MaxRows {
			return nil, fmt.Errorf("CSV exceeds %d rows", o.MaxRows)
		}
		text := strings.TrimSpace(record[ti])
		t, err := strconv.ParseFloat(text, 64)
		if err != nil {
			dt, parseErr := time.Parse(time.RFC3339Nano, text)
			if parseErr != nil {
				return nil, fmt.Errorf("CSV row %d: timestamp must be finite epoch seconds or RFC3339 with offset", row)
			}
			t = Epoch(dt)
		}
		if !validEpoch(t) || len(ts) > 0 && t <= ts[len(ts)-1] {
			return nil, fmt.Errorf("CSV row %d: timestamps must be finite and strictly increasing", row)
		}
		ts = append(ts, t)
		for j, index := range idx {
			s := strings.TrimSpace(record[index])
			v := math.NaN()
			switch strings.ToLower(s) {
			case "", "nan", "none", "null":
			default:
				v, err = strconv.ParseFloat(s, 64)
				if err != nil || math.IsInf(v, 0) {
					return nil, fmt.Errorf("CSV row %d, column %q: invalid finite value", row, cols[j].Column)
				}
			}
			vals[j] = append(vals[j], v)
		}
	}
	if trafficMode {
		g, err := Traffic(ts, vals[0], vals[1])
		if err != nil {
			return nil, err
		}
		return g.Series, nil
	}
	palette := []struct{ line, fill [3]uint8 }{{[3]uint8{0, 48, 0}, [3]uint8{0, 204, 0}}, {[3]uint8{0, 0, 204}, [3]uint8{153, 187, 255}}, {[3]uint8{153, 0, 0}, [3]uint8{255, 153, 153}}, {[3]uint8{128, 64, 0}, [3]uint8{255, 204, 102}}}
	out := make([]Series, len(cols))
	for i, c := range cols {
		name := c.Name
		if name == "" {
			name = c.Column
		}
		s, err := NewSeries(name, ts, vals[i])
		if err != nil {
			return nil, err
		}
		s.Kind = c.Kind
		p := palette[i%len(palette)]
		s.Color = RGB(p.line[0], p.line[1], p.line[2])
		if c.Kind == Area {
			line := s.Color
			s.Outline = &line
			s.Color = RGB(p.fill[0], p.fill[1], p.fill[2])
		}
		out[i] = s
	}
	return out, nil
}
