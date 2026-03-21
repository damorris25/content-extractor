// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"os"
	"testing"
)

func benchData(b *testing.B, name string) []byte {
	b.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		b.Fatalf("failed to load %s: %v", name, err)
	}
	return data
}

func BenchmarkExtract_JPEG_MetadataOnly(b *testing.B) {
	data := benchData(b, "simple.jpg")
	ext := New()
	r := bytes.NewReader(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		_, _ = ext.Extract(context.Background(), r, "image/jpeg")
	}
}

func BenchmarkExtract_PNG_MetadataOnly(b *testing.B) {
	data := benchData(b, "simple.png")
	ext := New()
	r := bytes.NewReader(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		_, _ = ext.Extract(context.Background(), r, "image/png")
	}
}

func BenchmarkExtract_GeoTIFF(b *testing.B) {
	data := benchData(b, "geotiff-sample.tif")
	ext := New()
	r := bytes.NewReader(data)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		_, _ = ext.Extract(context.Background(), r, "image/tiff")
	}
}
