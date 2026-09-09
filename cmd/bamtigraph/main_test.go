// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package main

import (
	"bytes"
	"encoding/json"
	rrd "github.com/gosuda/BamtiGraph"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func invoke(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, err bytes.Buffer
	code := run(args, &out, &err)
	return code, out.String(), err.String()
}
func TestCLIDemo(t *testing.T) {
	dir := t.TempDir()
	png, meta := filepath.Join(dir, "out.png"), filepath.Join(dir, "out.json")
	code, _, err := invoke(t, "-demo", "-o", png, "-metadata", meta, "-timezone", "Asia/Seoul")
	if code != 0 {
		t.Fatal(err)
	}
	if _, e := rrd.LoadImage(png); e != nil {
		t.Fatal(e)
	}
	p, e := os.ReadFile(meta)
	if e != nil {
		t.Fatal(e)
	}
	var m rrd.Metadata
	if e = json.Unmarshal(p, &m); e != nil {
		t.Fatal(e)
	}
	if m.Timezone != "Asia/Seoul" {
		t.Fatal(m.Timezone)
	}
	if m.Watermark != "" {
		t.Fatal("CLI default must not stamp a provider watermark")
	}
}
func TestCLICSV(t *testing.T) {
	dir := t.TempDir()
	in, out := filepath.Join(dir, "in.csv"), filepath.Join(dir, "out.png")
	if e := os.WriteFile(in, []byte("timestamp,cpu,wait\n0,30,1\n300,80,4\n"), 0600); e != nil {
		t.Fatal(e)
	}
	code, _, err := invoke(t, "-input", in, "-area", "cpu=CPU", "-line", "wait=Wait", "-vertical-label", "percent", "-y-max", "100", "-o", out)
	if code != 0 {
		t.Fatal(err)
	}
	if _, e := rrd.LoadImage(out); e != nil {
		t.Fatal(e)
	}
}
func TestCLIInfo(t *testing.T) {
	for _, a := range []string{"-help", "-version"} {
		code, out, err := invoke(t, a)
		if code != 0 || out+err == "" {
			t.Fatal(code, out, err)
		}
	}
}
func TestCLIErrors(t *testing.T) {
	for _, args := range [][]string{nil, {"-bad"}, {"-demo", "-input", "x"}, {"-input", "x", "y"}, {"x", "y"}, {"-demo", "-width", "0"}, {"-demo", "-y-max", "NaN"}, {"-demo", "-timezone", "unknown"}, {"-input", "/absent/csv"}, {"-demo", "-font", "/absent/font"}, {"-demo", "-y-max", "bad"}, {"-area", "=bad"}} {
		t.Run(strings.Join(args, "_"), func(t *testing.T) {
			code, _, err := invoke(t, args...)
			if code != 2 || err == "" {
				t.Fatal(code, err)
			}
		})
	}
	dir := t.TempDir()
	out := filepath.Join(dir, "x.png")
	alias := filepath.Join(dir, ".", "x.png")
	code, _, err := invoke(t, "-demo", "-o", out, "-metadata", alias)
	if code != 2 || !strings.Contains(err, "must differ") {
		t.Fatal(code, err)
	}
}
