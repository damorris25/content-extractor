// CONTRACT:C1-EXTRACTOR.1.0
package audio

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

const (
	// DefaultMaxAudioSize is the maximum size for reading audio metadata.
	// We only need the header portion, but buffer the whole file for safety.
	DefaultMaxAudioSize = 100 * 1024 * 1024 // 100MB
)

var supportedMIMEs = map[string]bool{
	"audio/mpeg": true,
	"audio/ogg":  true,
	"audio/flac": true,
	"audio/wav":  true,
}

// Extractor extracts metadata from audio files.
// Text content is a summary of metadata for searchability.
// Safe for concurrent use.
type Extractor struct {
	MaxSize int64
}

// New creates a new audio metadata extractor.
func New() *Extractor {
	return &Extractor{}
}

func (e *Extractor) maxSize() int64 {
	if e.MaxSize > 0 {
		return e.MaxSize
	}
	return DefaultMaxAudioSize
}

// Extract reads an audio file from r and returns extracted metadata.
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

	switch mime {
	case "audio/mpeg":
		parseID3v2(data, result)
	case "audio/ogg":
		parseOggVorbis(data, result)
	case "audio/flac":
		parseFLAC(data, result)
	case "audio/wav":
		parseWAV(data, result)
	}

	result.Text = buildTextSummary(result.Metadata)

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

// buildTextSummary creates a searchable text from audio metadata.
func buildTextSummary(m metadata.Metadata) string {
	var parts []string
	for _, key := range []string{
		metadata.KeyTitle,
		metadata.KeyCreator,
		"audio:album",
		"audio:genre",
	} {
		if v := m.Get(key); v != "" {
			parts = append(parts, v)
		}
	}
	return strings.Join(parts, "\n")
}

// setAudioMeta maps common tag names to standard metadata keys.
func setAudioMeta(m metadata.Metadata, key, value string) {
	value = strings.TrimSpace(value)
	if value == "" {
		return
	}
	switch strings.ToUpper(key) {
	case "TITLE", "TIT2":
		m.Set(metadata.KeyTitle, value)
	case "ARTIST", "TPE1":
		m.Set(metadata.KeyCreator, value)
	case "ALBUM", "TALB":
		m.Set("audio:album", value)
	case "TRACKNUMBER", "TRCK":
		m.Set("audio:track", value)
	case "DATE", "YEAR", "TDRC", "TYER":
		m.Set(metadata.KeyDate, metadata.NormalizeDate(value))
	case "GENRE", "TCON":
		m.Set("audio:genre", value)
	case "COMMENT", "COMM":
		m.Set(metadata.KeyDescription, value)
	case "ALBUMARTIST", "TPE2":
		m.Set("audio:album-artist", value)
	default:
		// Store other tags in the audio namespace.
		m.Set("audio:"+strings.ToLower(key), value)
	}
}

// readUint32BE reads a big-endian uint32 from data at offset.
// Returns 0 if offset is out of bounds.
func readUint32BE(data []byte, offset int) uint32 {
	if offset < 0 || offset+4 > len(data) {
		return 0
	}
	return uint32(data[offset])<<24 | uint32(data[offset+1])<<16 |
		uint32(data[offset+2])<<8 | uint32(data[offset+3])
}

// readUint32LE reads a little-endian uint32 from data at offset.
// Returns 0 if offset is out of bounds.
func readUint32LE(data []byte, offset int) uint32 {
	if offset < 0 || offset+4 > len(data) {
		return 0
	}
	return uint32(data[offset]) | uint32(data[offset+1])<<8 |
		uint32(data[offset+2])<<16 | uint32(data[offset+3])<<24
}

// readUint16LE reads a little-endian uint16 from data at offset.
// Returns 0 if offset is out of bounds.
func readUint16LE(data []byte, offset int) uint16 {
	if offset < 0 || offset+2 > len(data) {
		return 0
	}
	return uint16(data[offset]) | uint16(data[offset+1])<<8
}
