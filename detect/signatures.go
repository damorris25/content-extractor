// CONTRACT:C2-DETECT.1.0
package detect

// MagicSignature defines a file format's magic bytes.
type MagicSignature struct {
	MIME   string // MIME type this signature identifies
	Offset int64  // Byte offset where signature starts
	Magic  []byte // The magic byte sequence
	Mask   []byte // Optional mask (nil = exact match)
}

// signatures is the magic byte table, ordered by specificity (most specific first).
var signatures = []MagicSignature{
	// PDF
	{MIME: "application/pdf", Offset: 0, Magic: []byte("%PDF")},

	// ZIP (used for DOCX, XLSX, PPTX, ODT, ODS, ODP, JAR, etc.)
	{MIME: "application/zip", Offset: 0, Magic: []byte("PK\x03\x04")},

	// PNG
	{MIME: "image/png", Offset: 0, Magic: []byte("\x89PNG\r\n\x1a\n")},

	// JPEG
	{MIME: "image/jpeg", Offset: 0, Magic: []byte("\xff\xd8\xff")},

	// TIFF (little-endian)
	{MIME: "image/tiff", Offset: 0, Magic: []byte("II\x2a\x00")},

	// TIFF (big-endian)
	{MIME: "image/tiff", Offset: 0, Magic: []byte("MM\x00\x2a")},

	// BMP
	{MIME: "image/bmp", Offset: 0, Magic: []byte("BM")},

	// WebP (RIFF....WEBP)
	{MIME: "image/webp", Offset: 0, Magic: []byte("RIFF")},
	// TRACKED-TASK: WebP detection should also check bytes 8-11 for "WEBP" - see TODO.md "Code Debt"

	// GIF
	{MIME: "image/gif", Offset: 0, Magic: []byte("GIF87a")},
	{MIME: "image/gif", Offset: 0, Magic: []byte("GIF89a")},

	// MP3 (ID3v2 tag)
	{MIME: "audio/mpeg", Offset: 0, Magic: []byte("ID3")},

	// MP3 (sync word, no ID3 tag) — 0xFFE0 mask matches frame sync bits.
	{MIME: "audio/mpeg", Offset: 0, Magic: []byte{0xFF, 0xE0}, Mask: []byte{0xFF, 0xE0}},

	// OGG
	{MIME: "audio/ogg", Offset: 0, Magic: []byte("OggS")},

	// FLAC
	{MIME: "audio/flac", Offset: 0, Magic: []byte("fLaC")},

	// WAV (RIFF....WAVE)
	{MIME: "audio/wav", Offset: 0, Magic: []byte("RIFF")},
	// TRACKED-TASK: WAV/WebP disambiguation needs bytes 8-11 check - see TODO.md "Code Debt"

	// OLE2 Compound Document (legacy DOC, XLS, PPT)
	{MIME: "application/x-ole2", Offset: 0, Magic: []byte{0xD0, 0xCF, 0x11, 0xE0, 0xA1, 0xB1, 0x1A, 0xE1}},

	// DWG
	{MIME: "image/vnd.dwg", Offset: 0, Magic: []byte("AC10")},
}
