//go:build cgo && ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"testing"
	"time"
)

func BenchmarkOCR_RecognizePNG(b *testing.B) {
	data := benchData(b, "ocr-text.png")
	engine, err := NewOCREngine("eng")
	if err != nil {
		b.Fatalf("NewOCREngine: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = engine.RecognizeBytes(ctx, data)
	}
}

func BenchmarkOCR_RecognizeJPEG(b *testing.B) {
	data := benchData(b, "ocr-text.jpg")
	engine, err := NewOCREngine("eng")
	if err != nil {
		b.Fatalf("NewOCREngine: %v", err)
	}
	defer engine.Close()

	ctx := context.Background()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _, _ = engine.RecognizeBytes(ctx, data)
	}
}

func BenchmarkOCR_ExtractWithOCR_PNG(b *testing.B) {
	data := benchData(b, "ocr-text.png")
	engine, err := NewOCREngine("eng")
	if err != nil {
		b.Fatalf("NewOCREngine: %v", err)
	}
	defer engine.Close()

	ext := &Extractor{OCR: engine, OCRLanguage: "eng"}
	r := bytes.NewReader(data)
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		r.Reset(data)
		_, _ = ext.Extract(ctx, r, "image/png")
	}
}
