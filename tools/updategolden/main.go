// updategolden is an explicit developer operation. Review rendered pixels
// before accepting their digest; updating a golden is not a substitute for fixing bugs.
package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"flag"
	rrd "github.com/gosuda/BamtiGraph"
	"log"
	"os"
)

func main() {
	update := flag.Bool("update", false, "explicitly accept current demo pixels as the Go regression baseline")
	flag.Parse()
	if !*update {
		log.Fatal("requires -update; inspect the image before accepting it")
	}
	g, e := rrd.DemoTraffic(false)
	if e != nil {
		log.Fatal(e)
	}
	g.Watermark = "BAMTIGRAPH"
	r, e := g.RenderResult()
	if e != nil {
		log.Fatal(e)
	}
	bounds := r.Image.Bounds()
	h := sha256.New()
	for y := bounds.Min.Y; y < bounds.Max.Y; y++ {
		offset := r.Image.PixOffset(bounds.Min.X, y)
		h.Write(r.Image.Pix[offset : offset+4*bounds.Dx()])
	}
	golden := struct {
		Width      int                            `json:"width"`
		Height     int                            `json:"height"`
		RGBASHA256 string                         `json:"rgba_sha256"`
		Fonts      map[string]rrd.FontFingerprint `json:"fonts"`
	}{bounds.Dx(), bounds.Dy(), hex.EncodeToString(h.Sum(nil)), r.Metadata.Environment.Fonts}
	p, e := json.MarshalIndent(golden, "", "  ")
	if e != nil {
		log.Fatal(e)
	}
	if e = os.WriteFile("testdata/go_golden.json", append(p, '\n'), 0644); e != nil {
		log.Fatal(e)
	}
}
