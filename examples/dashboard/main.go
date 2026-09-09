package main

import (
	"flag"
	"fmt"
	rrd "github.com/gosuda/BamtiGraph"
	"log"
	"os"
	"path/filepath"
	_ "time/tzdata"
)

func main() {
	out := flag.String("o", "examples/output/dashboard.png", "output PNG")
	flag.Parse()
	a, e := rrd.DemoTraffic(false)
	if e != nil {
		log.Fatal(e)
	}
	b, e := rrd.DemoTraffic(true)
	if e != nil {
		log.Fatal(e)
	}
	im, e := rrd.Dashboard([]rrd.Panel{{Graph: a, Caption: "Daily (synthetic 5-minute samples)"}, {Graph: b, Caption: "Weekly (synthetic 30-minute samples)"}}, rrd.DefaultDashboardOptions())
	if e != nil {
		log.Fatal(e)
	}
	if e = os.MkdirAll(filepath.Dir(*out), 0755); e != nil {
		log.Fatal(e)
	}
	if e = rrd.SaveImagePNG(*out, im); e != nil {
		log.Fatal(e)
	}
	fmt.Println(*out)
}
