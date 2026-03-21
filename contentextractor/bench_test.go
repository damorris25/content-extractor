package contentextractor

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"testing"
)

// benchFixture holds a test file for benchmarking.
type benchFixture struct {
	name string
	file string
	mime string
}

var benchFixtures = []benchFixture{
	{"Text/1KB", "", "text/plain"},
	{"DOCX", "../testdata/simple.docx", ""},
	{"XLSX", "../testdata/simple.xlsx", ""},
	{"PPTX", "../testdata/simple.pptx", ""},
	{"PDF", "../testdata/simple.pdf", ""},
	{"ODT", "../testdata/simple.odt", ""},
	{"ODS", "../testdata/simple.ods", ""},
	{"ODP", "../testdata/simple.odp", ""},
	{"JPEG", "../testdata/simple.jpg", ""},
	{"PNG", "../testdata/text.png", ""},
	{"MP3", "../testdata/simple.mp3", "audio/mpeg"},
	{"OGG", "../testdata/simple.ogg", "audio/ogg"},
	{"FLAC", "../testdata/simple.flac", "audio/flac"},
	{"WAV", "../testdata/simple.wav", "audio/wav"},
	{"DXF", "../testdata/simple.dxf", "image/vnd.dxf"},
}

func BenchmarkExtract(b *testing.B) {
	reg := NewRegistry()
	ctx := context.Background()

	for _, fix := range benchFixtures {
		var data []byte
		if fix.file == "" {
			// Generate synthetic text data.
			data = bytes.Repeat([]byte("The quick brown fox jumps over the lazy dog. "), 23)
		} else {
			var err error
			data, err = os.ReadFile(fix.file)
			if err != nil {
				b.Logf("skipping %s: %v", fix.name, err)
				continue
			}
		}

		b.Run(fix.name, func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				r := bytes.NewReader(data)
				_, err := reg.Extract(ctx, r, fix.mime)
				if err != nil {
					b.Fatalf("extract error: %v", err)
				}
			}
		})
	}
}

func BenchmarkExtract_TextScaling(b *testing.B) {
	reg := NewRegistry()
	ctx := context.Background()

	for _, size := range []int{1024, 10240, 102400, 1048576} {
		data := bytes.Repeat([]byte("Hello world. This is a test of text extraction throughput. "), size/58+1)
		data = data[:size]

		b.Run(fmt.Sprintf("%dB", size), func(b *testing.B) {
			b.SetBytes(int64(len(data)))
			b.ReportAllocs()
			for b.Loop() {
				r := bytes.NewReader(data)
				_, err := reg.Extract(ctx, r, "text/plain")
				if err != nil {
					b.Fatalf("extract error: %v", err)
				}
			}
		})
	}
}

func BenchmarkNewRegistry(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		_ = NewRegistry()
	}
}

func BenchmarkDetection(b *testing.B) {
	reg := NewRegistry()
	ctx := context.Background()

	// Benchmark auto-detection overhead by comparing explicit MIME vs empty MIME.
	data, err := os.ReadFile("../testdata/simple.docx")
	if err != nil {
		b.Skip("testdata/simple.docx not found")
	}

	b.Run("ExplicitMIME", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		for b.Loop() {
			r := bytes.NewReader(data)
			_, err := reg.Extract(ctx, r, "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
			if err != nil {
				b.Fatalf("extract error: %v", err)
			}
		}
	})

	b.Run("AutoDetect", func(b *testing.B) {
		b.SetBytes(int64(len(data)))
		b.ReportAllocs()
		for b.Loop() {
			r := bytes.NewReader(data)
			_, err := reg.Extract(ctx, r, "")
			if err != nil {
				b.Fatalf("extract error: %v", err)
			}
		}
	})
}
