package main

import (
	"fmt"
	rrd "github.com/gosuda/BamtiGraph"
	"log"
	"math"
)

func main() {
	options := rrd.DefaultCounterOptions()
	options.Factor = 8
	options.CounterBits = 64
	base := uint64(1) << 60
	rates, e := rrd.CounterRate([]float64{0, 1, 2, 3}, []uint64{base, base + 10, base + 30, base + 60}, options)
	if e != nil {
		log.Fatal(e)
	}
	for i, v := range rates.Values {
		if math.IsNaN(v) {
			fmt.Printf("t=%.0f: missing\n", rates.Timestamps[i])
		} else {
			fmt.Printf("t=%.0f: %.0f bits/s\n", rates.Timestamps[i], v)
		}
	}
}
