//go:build !cgo || !ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"context"
	"errors"
)

// ErrOCRUnavailable is returned when OCR is not available in this build.
// Build with CGO_ENABLED=1 and -tags ocr to enable OCR support.
var ErrOCRUnavailable = errors.New("image: OCR not available (build without cgo+ocr tags)")

// OCREngine is a stub when OCR is not compiled in.
type OCREngine struct{}

// NewOCREngine returns ErrOCRUnavailable when OCR is not compiled in.
func NewOCREngine(_ string) (*OCREngine, error) {
	return nil, ErrOCRUnavailable
}

// Close is a no-op for the stub engine.
func (e *OCREngine) Close() error {
	return nil
}

// RecognizeBytes returns ErrOCRUnavailable when OCR is not compiled in.
func (e *OCREngine) RecognizeBytes(_ context.Context, _ []byte) (string, float64, error) {
	return "", 0, ErrOCRUnavailable
}
