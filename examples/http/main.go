// Render once, then serve immutable PNG bytes. Bounded server timeouts are set.
package main

import (
	"flag"
	rrd "github.com/gosuda/BamtiGraph"
	"log"
	"net/http"
	"time"
	_ "time/tzdata"
)

func main() {
	addr := flag.String("addr", "127.0.0.1:8080", "listen address")
	flag.Parse()
	g, e := rrd.DemoTraffic(false)
	if e != nil {
		log.Fatal(e)
	}
	png, e := g.PNGBytes()
	if e != nil {
		log.Fatal(e)
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/traffic.png", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=60")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		if r.Method != http.MethodHead {
			_, _ = w.Write(png)
		}
	})
	server := http.Server{Addr: *addr, Handler: mux, ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, IdleTimeout: 60 * time.Second}
	log.Printf("listening at %s/traffic.png", *addr)
	log.Fatal(server.ListenAndServe())
}
