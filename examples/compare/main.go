// Compare two decoded images without alignment or resampling.
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	rrd "github.com/gosuda/BamtiGraph"
	"log"
	"os"
)

func main() {
	a := flag.String("reference", "", "required reference image path")
	b := flag.String("actual", "", "required actual image path")
	diff := flag.String("diff", "examples/output/comparison-difference.png", "amplified diff path")
	flag.Parse()
	if *a == "" || *b == "" {
		fmt.Fprintln(os.Stderr, "Usage: compare -reference reference.png -actual actual.png [-diff difference.png]")
		flag.PrintDefaults()
		os.Exit(2)
	}
	report, e := rrd.CompareFiles(*a, *b, rrd.CompareOptions{})
	if e != nil {
		log.Fatal(e)
	}
	data, e := json.MarshalIndent(report, "", "  ")
	if e != nil {
		log.Fatal(e)
	}
	fmt.Println(string(data))
	ia, e := rrd.LoadImage(*a)
	if e != nil {
		log.Fatal(e)
	}
	ib, e := rrd.LoadImage(*b)
	if e != nil {
		log.Fatal(e)
	}
	im, e := rrd.DifferenceImage(ia, ib, 4)
	if e != nil {
		log.Fatal(e)
	}
	if e = rrd.SaveImagePNG(*diff, im); e != nil {
		log.Fatal(e)
	}
}
