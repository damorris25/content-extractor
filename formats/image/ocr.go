//go:build cgo && ocr

// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/otiai10/gosseract/v2"
)

// OCREngine wraps Tesseract for OCR text extraction from images.
// NOT safe for concurrent use — callers must serialize access or pool engines.
// Create once and reuse across calls for best performance.
type OCREngine struct {
	client *gosseract.Client
	mu     sync.Mutex
}

// NewOCREngine creates a new OCR engine with the specified language.
// Language codes follow Tesseract conventions (e.g., "eng", "fra", "deu").
// Returns an error if Tesseract libraries are not available.
func NewOCREngine(lang string) (*OCREngine, error) {
	if lang == "" {
		lang = "eng"
	}
	client := gosseract.NewClient()
	if err := client.SetLanguage(lang); err != nil {
		client.Close()
		return nil, fmt.Errorf("image/ocr: set language %q: %w", lang, err)
	}
	return &OCREngine{client: client}, nil
}

// Close releases Tesseract resources. Must be called when the engine is no longer needed.
func (e *OCREngine) Close() error {
	e.mu.Lock()
	defer e.mu.Unlock()
	if e.client != nil {
		e.client.Close()
		e.client = nil
	}
	return nil
}

// RecognizeBytes performs OCR on image data and returns the extracted text
// and a confidence score (0-100). Respects context cancellation.
//
// The engine serializes access internally — only one recognition at a time.
// For concurrent OCR, create multiple OCREngine instances.
func (e *OCREngine) RecognizeBytes(ctx context.Context, data []byte) (text string, confidence float64, err error) {
	if err := ctx.Err(); err != nil {
		return "", 0, err
	}
	if len(data) == 0 {
		return "", 0, nil
	}

	// Run OCR in a goroutine to support context cancellation.
	type result struct {
		text       string
		confidence float64
		err        error
	}
	ch := make(chan result, 1)

	go func() {
		e.mu.Lock()
		defer e.mu.Unlock()

		if e.client == nil {
			ch <- result{err: fmt.Errorf("image/ocr: engine closed")}
			return
		}

		if err := e.client.SetImageFromBytes(data); err != nil {
			ch <- result{err: fmt.Errorf("image/ocr: set image: %w", err)}
			return
		}

		ocrText, err := e.client.Text()
		if err != nil {
			ch <- result{err: fmt.Errorf("image/ocr: recognize: %w", err)}
			return
		}

		// Compute mean confidence from word bounding boxes.
		var conf float64
		boxes, boxErr := e.client.GetBoundingBoxes(gosseract.RIL_WORD)
		if boxErr == nil && len(boxes) > 0 {
			var sum float64
			for _, box := range boxes {
				sum += box.Confidence
			}
			conf = sum / float64(len(boxes))
		}

		ch <- result{text: strings.TrimSpace(ocrText), confidence: conf}
	}()

	select {
	case <-ctx.Done():
		return "", 0, ctx.Err()
	case r := <-ch:
		return r.text, r.confidence, r.err
	}
}
