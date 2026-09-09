package main

import (
	"encoding/json"
	"flag"
	"fmt"
	rrd "github.com/gosuda/BamtiGraph"
	"github.com/gosuda/BamtiGraph/internal/fixture"
	"log"
	"os"
	"path/filepath"
	_ "time/tzdata"
)

func main() {
	fixtures := flag.String("fixtures", "testdata/python_cases.json", "shared fixture JSON")
	out := flag.String("out", "examples/output", "output directory")
	referenceDir := flag.String("reference-dir", "", "optional directory containing python_<case>.png reference images")
	flag.Parse()
	cases, e := fixture.Load(*fixtures)
	if e != nil {
		log.Fatal(e)
	}
	if e = os.MkdirAll(*out, 0755); e != nil {
		log.Fatal(e)
	}
	type report struct {
		Name      string `json:"name"`
		Compared  bool   `json:"compared"`
		Actual    string `json:"actual"`
		Reference string `json:"reference,omitempty"`
		*rrd.PixelDifference
		GoEnvironment rrd.Environment `json:"go_environment"`
	}
	reports := make([]report, 0, len(cases))
	for _, c := range cases {
		g, e := c.Graph()
		if e != nil {
			log.Fatal(e)
		}
		path := filepath.Join(*out, "go_"+c.Name+".png")
		r, e := g.Save(path)
		if e != nil {
			log.Fatal(e)
		}
		meta, e := r.MetadataJSON()
		if e != nil {
			log.Fatal(e)
		}
		if e = os.WriteFile(filepath.Join(*out, "go_"+c.Name+".json"), meta, 0644); e != nil {
			log.Fatal(e)
		}
		entry := report{Name: c.Name, Actual: path, GoEnvironment: r.Metadata.Environment}
		if *referenceDir == "" {
			reports = append(reports, entry)
			fmt.Printf("%-12s rendered %s (no reference comparison requested)\n", c.Name, path)
			continue
		}
		py := filepath.Join(*referenceDir, "python_"+c.Name+".png")
		d, e := rrd.CompareFiles(py, path, rrd.CompareOptions{})
		if e != nil {
			log.Fatal(e)
		}
		entry.Compared, entry.Reference, entry.PixelDifference = true, py, &d
		reports = append(reports, entry)
		fmt.Printf("%-12s exact=%6.2f%% MAE=%8.5f/255\n", c.Name, 100*d.ExactRatio, d.MeanAbsoluteError)
		if c.Name == "traffic" {
			a, e := rrd.LoadImage(py)
			if e != nil {
				log.Fatal(e)
			}
			diff, e := rrd.DifferenceImage(a, r.Image, 4)
			if e != nil {
				log.Fatal(e)
			}
			if e = rrd.SaveImagePNG(filepath.Join(*out, "python_go_difference.png"), diff); e != nil {
				log.Fatal(e)
			}
		}
	}
	data, e := json.MarshalIndent(reports, "", "  ")
	if e != nil {
		log.Fatal(e)
	}
	if e = os.WriteFile(filepath.Join(*out, "python_go_comparison.json"), append(data, '\n'), 0644); e != nil {
		log.Fatal(e)
	}
}
