// SPDX-License-Identifier: BSD-3-Clause
// Copyright (c) 2026 GoSuda. All rights reserved.
// See LICENSE for the project license.

package bamtigraph

import (
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"image"
	"image/png"
	"io"
	"os"
	"path/filepath"
)

// EncodePNG writes an already rendered image. Metadata is a standard PNG tEXt
// chunk named "chart". The encoded JSON is ASCII-escaped for PNG Latin-1 text
// compliance. Font bytes and timestamps of the render operation are never stored.
func (r *RenderResult) EncodePNG(w io.Writer, includeMetadata bool) error {
	if r == nil || r.Image == nil {
		return fmt.Errorf("nil render result/image")
	}
	var raw bytes.Buffer
	enc := png.Encoder{CompressionLevel: png.BestCompression}
	if e := enc.Encode(&raw, r.Image); e != nil {
		return e
	}
	data := raw.Bytes()
	if !includeMetadata {
		_, e := io.Copy(w, bytes.NewReader(data))
		return e
	}
	meta, e := json.Marshal(r.Metadata)
	if e != nil {
		return e
	}
	meta = asciiJSON(meta)
	if len(meta) > 8<<20 {
		return fmt.Errorf("PNG metadata exceeds 8 MiB")
	}
	payload := append([]byte("chart\x00"), meta...)
	var chunk bytes.Buffer
	_ = binary.Write(&chunk, binary.BigEndian, uint32(len(payload)))
	chunk.WriteString("tEXt")
	chunk.Write(payload)
	checksum := crc32.ChecksumIEEE(chunk.Bytes()[4:])
	_ = binary.Write(&chunk, binary.BigEndian, checksum)
	// PNG signature plus the mandatory 25-byte IHDR chunk.
	if len(data) < 33 {
		return fmt.Errorf("unexpected PNG encoder result")
	}
	for _, p := range [][]byte{data[:33], chunk.Bytes(), data[33:]} {
		if _, e := io.Copy(w, bytes.NewReader(p)); e != nil {
			return e
		}
	}
	return nil
}
func asciiJSON(p []byte) []byte {
	var out bytes.Buffer
	for _, r := range string(p) {
		if r < 128 {
			out.WriteByte(byte(r))
		} else if r <= 65535 {
			fmt.Fprintf(&out, "\\u%04x", r)
		} else {
			r -= 0x10000
			fmt.Fprintf(&out, "\\u%04x\\u%04x", 0xd800+(r>>10), 0xdc00+(r&1023))
		}
	}
	return out.Bytes()
}
func (r *RenderResult) PNGBytes() ([]byte, error) {
	var b bytes.Buffer
	e := r.EncodePNG(&b, true)
	return b.Bytes(), e
}
func (g *Graph) PNGBytes() ([]byte, error) {
	r, e := g.RenderResult()
	if e != nil {
		return nil, e
	}
	return r.PNGBytes()
}
func atomicWrite(path string, write func(io.Writer) error) error {
	f, e := os.CreateTemp(filepath.Dir(path), ".bamtigraph-*.tmp")
	if e != nil {
		return e
	}
	name := f.Name()
	defer os.Remove(name)
	if e = write(f); e != nil {
		_ = f.Close()
		return e
	}
	if e = f.Close(); e != nil {
		return e
	}
	return os.Rename(name, path)
}
func (r *RenderResult) Save(path string) error {
	return atomicWrite(path, func(w io.Writer) error { return r.EncodePNG(w, true) })
}
func (g *Graph) Save(path string) (*RenderResult, error) {
	r, e := g.RenderResult()
	if e != nil {
		return nil, e
	}
	if e = r.Save(path); e != nil {
		return nil, e
	}
	return r, nil
}

// SaveImagePNG saves a dashboard or an ordinary image, without a graph manifest.
func SaveImagePNG(path string, im image.Image) error {
	if im == nil {
		return fmt.Errorf("nil image")
	}
	return atomicWrite(path, func(w io.Writer) error {
		enc := png.Encoder{CompressionLevel: png.BestCompression}
		return enc.Encode(w, im)
	})
}
func (r *RenderResult) MetadataJSON() ([]byte, error) {
	return json.MarshalIndent(r.Metadata, "", "  ")
}
