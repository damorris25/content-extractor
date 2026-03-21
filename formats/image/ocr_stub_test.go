//go:build !cgo || !ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"testing"
)

func TestOCR_StubUnavailable(t *testing.T) {
	_, err := NewOCREngine("eng")
	if err != ErrOCRUnavailable {
		t.Errorf("NewOCREngine() = %v, want ErrOCRUnavailable", err)
	}
}

func TestOCR_StubRecognize(t *testing.T) {
	e := &OCREngine{}
	_, _, err := e.RecognizeBytes(context.Background(), []byte("data"))
	if err != ErrOCRUnavailable {
		t.Errorf("RecognizeBytes() = %v, want ErrOCRUnavailable", err)
	}
}

func TestOCR_ExtractorWithNilOCR(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.jpg")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/jpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// With nil OCR, no ocr: metadata keys should be set.
	if result.Metadata.Get("ocr:confidence") != "" {
		t.Error("expected no ocr:confidence with nil OCR engine")
	}
	if result.Metadata.Get("ocr:language") != "" {
		t.Error("expected no ocr:language with nil OCR engine")
	}
}
