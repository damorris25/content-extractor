// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"fmt"
	"image"
	_ "image/gif"  // register GIF decoder
	_ "image/jpeg" // register JPEG decoder
	_ "image/png"  // register PNG decoder
	"io"
	"strings"

	goexif "github.com/rwcarlsen/goexif/exif"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
)

const (
	// DefaultMaxImageSize is the maximum size for an image file.
	DefaultMaxImageSize = 100 * 1024 * 1024 // 100MB
)

var supportedMIMEs = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
	"image/tiff": true,
	"image/bmp":  true,
	"image/gif":  true,
	"image/webp": true,
	"image/heic": true,
	"image/heif": true,
	"image/avif": true,
}

// Extractor extracts metadata and optionally text (via OCR) from image files.
// When OCR is nil (default), only metadata is extracted.
// When OCR is set, text within images is recognized via Tesseract.
// Safe for concurrent use when OCR is nil; with OCR, the engine serializes internally.
type Extractor struct {
	MaxSize     int64
	OCR         *OCREngine // nil = no OCR (default). Set via NewOCREngine().
	OCRLanguage string     // Tesseract language code (default "eng").
}

// New creates a new image metadata extractor.
func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxImageSize
}


// Extract reads an image file from r and returns extracted metadata.
func (e *Extractor) Extract(ctx context.Context, r io.Reader, mime string) (*extractor.Result, error) {
	if r == nil {
		return nil, extractor.ErrNilReader
	}
	if !supportedMIMEs[mime] {
		return nil, extractor.ErrUnsupportedFormat
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	// Buffer input for multiple passes (dimensions + EXIF).
	limited := io.LimitReader(r, e.maxSize()+1)
	data, err := io.ReadAll(limited)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", extractor.ErrCorruptInput, err)
	}
	if int64(len(data)) > e.maxSize() {
		return nil, extractor.ErrFileTooLarge
	}
	if len(data) == 0 {
		return &extractor.Result{ContentType: mime, Metadata: make(metadata.Metadata)}, nil
	}

	if err := ctx.Err(); err != nil {
		return nil, err
	}

	result := &extractor.Result{
		ContentType: mime,
		Metadata:    make(metadata.Metadata),
	}
	result.Metadata.Set(metadata.KeyFormat, mime)

	// Format-specific metadata extraction.
	switch mime {
	case "image/heic", "image/heif", "image/avif":
		// ISOBMFF container — dimensions from ispe, EXIF from embedded item.
		// OCR not supported for HEIC/AVIF (no pure-Go decoder; Tesseract/Leptonica can't decode).
		extractISOBMFF(data, result)
	default:
		// Extract image dimensions via stdlib.
		extractDimensions(data, result)

		// Extract EXIF metadata (JPEG and TIFF only).
		if mime == "image/jpeg" || mime == "image/tiff" {
			extractEXIF(data, result)
		}

		// Extract PNG text chunks.
		if mime == "image/png" {
			extractPNGText(data, result)
		}

		// Extract GeoTIFF metadata (TIFF only).
		if mime == "image/tiff" {
			extractGeoTIFF(data, result.Metadata)
		}
	}

	// Perform OCR if an engine is configured.
	if e.OCR != nil {
		ocrText, confidence, ocrErr := e.OCR.RecognizeBytes(ctx, data)
		if ocrErr == nil && ocrText != "" {
			lang := e.OCRLanguage
			if lang == "" {
				lang = "eng"
			}
			result.Text = ocrText
			result.Metadata.Set("ocr:confidence", fmt.Sprintf("%.1f", confidence))
			result.Metadata.Set("ocr:language", lang)
		}
		// On OCR failure, fall through to metadata-only text summary.
	}

	// Build a text summary from key metadata for searchability (fallback when no OCR text).
	if result.Text == "" {
		result.Text = buildTextSummary(result.Metadata)
	}

	return result, nil
}

// SupportedMIMETypes returns the MIME types this extractor handles.
func (e *Extractor) SupportedMIMETypes() []string {
	types := make([]string, 0, len(supportedMIMEs))
	for mime := range supportedMIMEs {
		types = append(types, mime)
	}
	return types
}

// extractDimensions uses stdlib image.DecodeConfig to get width/height.
func extractDimensions(data []byte, result *extractor.Result) {
	cfg, _, err := image.DecodeConfig(bytes.NewReader(data))
	if err != nil {
		return
	}
	if cfg.Width > 0 {
		result.Metadata.Set("exif:width", fmt.Sprintf("%d", cfg.Width))
	}
	if cfg.Height > 0 {
		result.Metadata.Set("exif:height", fmt.Sprintf("%d", cfg.Height))
	}
}

// extractEXIF parses EXIF data from JPEG/TIFF files.
// Uses panic recovery because goexif may panic on malformed input.
func extractEXIF(data []byte, result *extractor.Result) {
	defer func() {
		if r := recover(); r != nil {
			// Malformed EXIF — silently ignore, same as parse error.
		}
	}()
	x, err := goexif.Decode(bytes.NewReader(data))
	if err != nil {
		return // No EXIF data is not an error.
	}
	extractEXIFFromDecoded(x, result)
}

// extractPNGText extracts tEXt chunks from PNG data.
// PNG text chunks have the format: keyword\0text
func extractPNGText(data []byte, result *extractor.Result) {
	// PNG signature is 8 bytes, then chunks follow.
	if len(data) < 8 {
		return
	}
	pos := 8 // skip signature

	for pos+12 <= len(data) {
		// Chunk: 4 bytes length, 4 bytes type, data, 4 bytes CRC.
		cl := uint32(data[pos])<<24 | uint32(data[pos+1])<<16 | uint32(data[pos+2])<<8 | uint32(data[pos+3])
		chunkType := string(data[pos+4 : pos+8])

		if cl > uint32(len(data)-pos-12) {
			break
		}
		chunkLen := int(cl)

		chunkData := data[pos+8 : pos+8+chunkLen]

		if chunkType == "tEXt" {
			// Format: keyword\0text
			if idx := bytes.IndexByte(chunkData, 0); idx >= 0 {
				key := string(chunkData[:idx])
				val := string(chunkData[idx+1:])
				pngKey := "png:" + strings.ToLower(key)
				result.Metadata.Set(pngKey, val)

				// Map well-known PNG keys to standard metadata.
				switch strings.ToLower(key) {
				case "title":
					result.Metadata.Set(metadata.KeyTitle, val)
				case "author":
					result.Metadata.Set(metadata.KeyCreator, val)
				case "description":
					result.Metadata.Set(metadata.KeyDescription, val)
				case "software":
					result.Metadata.Set(metadata.KeyApplication, val)
				}
			}
		}

		pos += 12 + chunkLen // length(4) + type(4) + data + crc(4)
	}
}

// buildTextSummary creates a searchable text representation from metadata.
func buildTextSummary(m metadata.Metadata) string {
	var parts []string
	for _, key := range []string{
		metadata.KeyTitle,
		metadata.KeyDescription,
		metadata.KeyCreator,
		"exif:make",
		"exif:model",
		"geo:crs",
		"geo:bounds",
	} {
		if v := m.Get(key); v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, "\n")
}

// cleanTag removes quotes from EXIF tag string values.
func cleanTag(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
	}
	return s
}
