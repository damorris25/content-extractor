// CONTRACT:C1-EXTRACTOR.1.0
package audio

import (
	"bytes"
	"context"
	"os"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/extractor/extractortest"
	"github.com/dmorris/content-extractor/metadata"
)

func loadTestFile(t *testing.T, name string) []byte {
	t.Helper()
	data, err := os.ReadFile("../../testdata/" + name)
	if err != nil {
		t.Fatalf("failed to load test file %s: %v", name, err)
	}
	return data
}

func TestAudio_Compliance_MP3(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	extractortest.RunComplianceTests(t, ext, "audio/mpeg", data)
}

func TestAudio_Compliance_OGG(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.ogg")
	extractortest.RunComplianceTests(t, ext, "audio/ogg", data)
}

func TestAudio_Compliance_FLAC(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.flac")
	extractortest.RunComplianceTests(t, ext, "audio/flac", data)
}

func TestAudio_Compliance_WAV(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.wav")
	extractortest.RunComplianceTests(t, ext, "audio/wav", data)
}

func TestAudio_MP3_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "Test Song"},
		{metadata.KeyCreator, "Test Artist"},
		{"audio:album", "Test Album"},
		{"audio:track", "3"},
		{"audio:genre", "Electronic"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_OGG_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.ogg")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/ogg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "OGG Test Song"},
		{metadata.KeyCreator, "OGG Test Artist"},
		{"audio:album", "OGG Test Album"},
		{"audio:track", "5"},
		{"audio:genre", "Ambient"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_FLAC_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.flac")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/flac")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "FLAC Test Song"},
		{metadata.KeyCreator, "FLAC Test Artist"},
		{"audio:album", "FLAC Test Album"},
		{"audio:track", "7"},
		{"audio:genre", "Classical"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_WAV_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.wav")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/wav")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{metadata.KeyTitle, "WAV Test Song"},
		{metadata.KeyCreator, "WAV Test Artist"},
		{"audio:album", "WAV Test Album"},
		{"audio:genre", "Jazz"},
		{"audio:channels", "2"},
		{"audio:sample-rate", "44100"},
		{"audio:bits-per-sample", "16"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestAudio_NilReader(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), nil, "audio/mpeg")
	if err != extractor.ErrNilReader {
		t.Errorf("error = %v, want ErrNilReader", err)
	}
}

func TestAudio_UnsupportedMIME(t *testing.T) {
	ext := New()
	_, err := ext.Extract(context.Background(), bytes.NewReader([]byte("data")), "text/plain")
	if err != extractor.ErrUnsupportedFormat {
		t.Errorf("error = %v, want ErrUnsupportedFormat", err)
	}
}

func TestAudio_TextSummary(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.mp3")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Text == "" {
		t.Error("Text summary is empty")
	}
}

// --- Contract error paths ---

func TestAudio_DefaultMaxSize(t *testing.T) {
	ext := &Extractor{MaxSize: 0}
	if got := ext.maxSize(); got != DefaultMaxAudioSize {
		t.Errorf("maxSize() = %d, want %d", got, DefaultMaxAudioSize)
	}
}

func TestAudio_CustomMaxSize(t *testing.T) {
	ext := &Extractor{MaxSize: 1024}
	if got := ext.maxSize(); got != 1024 {
		t.Errorf("maxSize() = %d, want 1024", got)
	}
}

func TestAudio_FileTooLarge(t *testing.T) {
	ext := &Extractor{MaxSize: 10}
	_, err := ext.Extract(context.Background(), bytes.NewReader(make([]byte, 20)), "audio/mpeg")
	if err != extractor.ErrFileTooLarge {
		t.Errorf("error = %v, want ErrFileTooLarge", err)
	}
}

func TestAudio_EmptyReader(t *testing.T) {
	ext := New()
	result, err := ext.Extract(context.Background(), bytes.NewReader(nil), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
}

func TestAudio_ContextCancelled(t *testing.T) {
	ext := New()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := ext.Extract(ctx, bytes.NewReader([]byte("data")), "audio/mpeg")
	if err != context.Canceled {
		t.Errorf("error = %v, want context.Canceled", err)
	}
}

// --- readUint32BE/LE/readUint16LE success paths ---

func TestReadUint32BE_Valid(t *testing.T) {
	data := []byte{0x01, 0x02, 0x03, 0x04, 0x05}
	got := readUint32BE(data, 0)
	want := uint32(0x01020304)
	if got != want {
		t.Errorf("readUint32BE = %#x, want %#x", got, want)
	}
}

func TestReadUint32BE_OOB(t *testing.T) {
	data := []byte{0x01, 0x02}
	got := readUint32BE(data, 0)
	if got != 0 {
		t.Errorf("readUint32BE(oob) = %d, want 0", got)
	}
}

func TestReadUint32BE_NegativeOffset(t *testing.T) {
	got := readUint32BE([]byte{1, 2, 3, 4}, -1)
	if got != 0 {
		t.Errorf("readUint32BE(-1) = %d, want 0", got)
	}
}

func TestReadUint32LE_Valid(t *testing.T) {
	data := []byte{0x04, 0x03, 0x02, 0x01}
	got := readUint32LE(data, 0)
	want := uint32(0x01020304)
	if got != want {
		t.Errorf("readUint32LE = %#x, want %#x", got, want)
	}
}

func TestReadUint16LE_Valid(t *testing.T) {
	data := []byte{0xCD, 0xAB}
	got := readUint16LE(data, 0)
	want := uint16(0xABCD)
	if got != want {
		t.Errorf("readUint16LE = %#x, want %#x", got, want)
	}
}

// --- setAudioMeta comprehensive ---

func TestSetAudioMeta(t *testing.T) {
	tests := []struct {
		key     string
		value   string
		metaKey string
		want    string
	}{
		{"TITLE", "Song", metadata.KeyTitle, "Song"},
		{"TIT2", "Song2", metadata.KeyTitle, "Song2"},
		{"ARTIST", "Art", metadata.KeyCreator, "Art"},
		{"TPE1", "Art2", metadata.KeyCreator, "Art2"},
		{"ALBUM", "Alb", "audio:album", "Alb"},
		{"TALB", "Alb2", "audio:album", "Alb2"},
		{"TRACKNUMBER", "3", "audio:track", "3"},
		{"TRCK", "4", "audio:track", "4"},
		{"DATE", "2024", metadata.KeyDate, "2024-01-01T00:00:00Z"},
		{"YEAR", "2023", metadata.KeyDate, "2023-01-01T00:00:00Z"},
		{"TDRC", "2022", metadata.KeyDate, "2022-01-01T00:00:00Z"},
		{"TYER", "2021", metadata.KeyDate, "2021-01-01T00:00:00Z"},
		{"GENRE", "Rock", "audio:genre", "Rock"},
		{"TCON", "Pop", "audio:genre", "Pop"},
		{"COMMENT", "Nice", metadata.KeyDescription, "Nice"},
		{"COMM", "Good", metadata.KeyDescription, "Good"},
		{"ALBUMARTIST", "AA", "audio:album-artist", "AA"},
		{"TPE2", "AA2", "audio:album-artist", "AA2"},
		{"CUSTOM", "val", "audio:custom", "val"},
	}
	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			m := make(metadata.Metadata)
			setAudioMeta(m, tt.key, tt.value)
			if got := m.Get(tt.metaKey); got != tt.want {
				t.Errorf("setAudioMeta(%q, %q): Metadata[%q] = %q, want %q", tt.key, tt.value, tt.metaKey, got, tt.want)
			}
		})
	}
}

func TestSetAudioMeta_EmptyValue(t *testing.T) {
	m := make(metadata.Metadata)
	setAudioMeta(m, "TITLE", "  ")
	if got := m.Get(metadata.KeyTitle); got != "" {
		t.Errorf("setAudioMeta(empty) = %q, want empty", got)
	}
}

// --- ID3v2.2 frames ---

func TestParseID3v22(t *testing.T) {
	// Build a synthetic ID3v2.2 tag.
	// Header: "ID3" + version(2,0) + flags(0) + size(syncsafe)
	var tag []byte
	tag = append(tag, 'I', 'D', '3')
	tag = append(tag, 2, 0) // version 2.0
	tag = append(tag, 0)    // flags

	// Build frames: TT2 (title), TP1 (artist)
	var frames []byte

	// TT2 frame: 3-byte ID + 3-byte size + data
	// encoding(1) + "Test Song v2.2"
	frameData := append([]byte{3}, []byte("Test Song v2.2")...) // encoding=3 (UTF-8)
	frames = append(frames, 'T', 'T', '2')
	frames = append(frames, byte(len(frameData)>>16), byte(len(frameData)>>8), byte(len(frameData)))
	frames = append(frames, frameData...)

	// TP1 frame
	frameData2 := append([]byte{3}, []byte("Test Artist v2.2")...)
	frames = append(frames, 'T', 'P', '1')
	frames = append(frames, byte(len(frameData2)>>16), byte(len(frameData2)>>8), byte(len(frameData2)))
	frames = append(frames, frameData2...)

	// Syncsafe size of frames
	size := len(frames)
	tag = append(tag,
		byte((size>>21)&0x7F),
		byte((size>>14)&0x7F),
		byte((size>>7)&0x7F),
		byte(size&0x7F),
	)
	tag = append(tag, frames...)

	ext := New()
	result, err := ext.Extract(context.Background(), bytes.NewReader(tag), "audio/mpeg")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get(metadata.KeyTitle); got != "Test Song v2.2" {
		t.Errorf("title = %q, want 'Test Song v2.2'", got)
	}
	if got := result.Metadata.Get(metadata.KeyCreator); got != "Test Artist v2.2" {
		t.Errorf("artist = %q, want 'Test Artist v2.2'", got)
	}
}

// --- decodeTextFrame encoding variants ---

func TestDecodeTextFrame_UTF8(t *testing.T) {
	data := append([]byte{3}, []byte("Hello UTF-8")...)
	got := decodeTextFrame(data)
	if got != "Hello UTF-8" {
		t.Errorf("decodeTextFrame(UTF-8) = %q, want 'Hello UTF-8'", got)
	}
}

func TestDecodeTextFrame_Latin1(t *testing.T) {
	data := append([]byte{0}, []byte("Hello Latin-1")...)
	got := decodeTextFrame(data)
	if got != "Hello Latin-1" {
		t.Errorf("decodeTextFrame(Latin-1) = %q, want 'Hello Latin-1'", got)
	}
}

func TestDecodeTextFrame_UTF16_LE(t *testing.T) {
	// Encoding=1 (UTF-16 with BOM), LE BOM = FF FE, then "Hi" = 0x48 0x00 0x69 0x00
	data := []byte{1, 0xFF, 0xFE, 0x48, 0x00, 0x69, 0x00}
	got := decodeTextFrame(data)
	if got != "Hi" {
		t.Errorf("decodeTextFrame(UTF-16 LE) = %q, want 'Hi'", got)
	}
}

func TestDecodeTextFrame_UTF16_BE(t *testing.T) {
	// Encoding=1 (UTF-16 with BOM), BE BOM = FE FF, then "Hi" = 0x00 0x48 0x00 0x69
	data := []byte{1, 0xFE, 0xFF, 0x00, 0x48, 0x00, 0x69}
	got := decodeTextFrame(data)
	if got != "Hi" {
		t.Errorf("decodeTextFrame(UTF-16 BE) = %q, want 'Hi'", got)
	}
}

func TestDecodeTextFrame_UTF16BE_NoBOM(t *testing.T) {
	// Encoding=2 (UTF-16BE without BOM), "Ab" = 0x00 0x41 0x00 0x62
	data := []byte{2, 0x00, 0x41, 0x00, 0x62}
	got := decodeTextFrame(data)
	if got != "Ab" {
		t.Errorf("decodeTextFrame(UTF-16BE no BOM) = %q, want 'Ab'", got)
	}
}

func TestDecodeTextFrame_UnknownEncoding(t *testing.T) {
	data := []byte{99, 'H', 'e', 'l', 'l', 'o'}
	got := decodeTextFrame(data)
	if got != "Hello" {
		t.Errorf("decodeTextFrame(unknown) = %q, want 'Hello'", got)
	}
}

func TestDecodeTextFrame_TooShort(t *testing.T) {
	got := decodeTextFrame([]byte{3})
	if got != "" {
		t.Errorf("decodeTextFrame(short) = %q, want empty", got)
	}
}

// --- decodeUTF16 edge cases ---

func TestDecodeUTF16_NullTerminator(t *testing.T) {
	// LE BOM, "A" then null.
	data := []byte{0xFF, 0xFE, 0x41, 0x00, 0x00, 0x00}
	got := decodeUTF16(data)
	if got != "A" {
		t.Errorf("decodeUTF16(null term) = %q, want 'A'", got)
	}
}

func TestDecodeUTF16_NoBOM(t *testing.T) {
	// No BOM — defaults to big-endian.
	data := []byte{0x00, 0x41, 0x00, 0x42}
	got := decodeUTF16(data)
	if got != "AB" {
		t.Errorf("decodeUTF16(no BOM) = %q, want 'AB'", got)
	}
}

func TestDecodeUTF16_TooShort(t *testing.T) {
	got := decodeUTF16([]byte{0xFF})
	if got != "" {
		t.Errorf("decodeUTF16(short) = %q, want empty", got)
	}
}

func TestDecodeUTF16BE_NullTerminator(t *testing.T) {
	data := []byte{0x00, 0x41, 0x00, 0x00}
	got := decodeUTF16BE(data)
	if got != "A" {
		t.Errorf("decodeUTF16BE(null term) = %q, want 'A'", got)
	}
}

// --- mapV22Frame ---

func TestMapV22Frame(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"TT2", "TIT2"},
		{"TP1", "TPE1"},
		{"TAL", "TALB"},
		{"TRK", "TRCK"},
		{"TYE", "TYER"},
		{"TCO", "TCON"},
		{"COM", "COMM"},
		{"XXX", "XXX"}, // unknown → pass-through
	}
	for _, tt := range tests {
		got := mapV22Frame(tt.input)
		if got != tt.want {
			t.Errorf("mapV22Frame(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- Vorbis/FLAC/WAV edge cases ---

func TestParseVorbisComment_Truncated(t *testing.T) {
	// Vendor length claims more than available.
	data := []byte{0xFF, 0xFF, 0x00, 0x00} // vendorLen=65535, but data is only 4 bytes.
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseVorbisComment(data, result)
	// Should not panic.
}

func TestParseVorbisComment_TooShort(t *testing.T) {
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseVorbisComment([]byte{1, 2, 3}, result)
	// Should not panic (< 8 bytes).
}

func TestParseOggVorbis_NoMarker(t *testing.T) {
	// OGG without Vorbis marker.
	data := []byte("OggS" + "no vorbis marker in this data at all")
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseOggVorbis(data, result)
	// Should not panic, no metadata extracted.
	if got := result.Metadata.Get(metadata.KeyTitle); got != "" {
		t.Errorf("title = %q, want empty", got)
	}
}

func TestParseFLAC_TruncatedBlock(t *testing.T) {
	// FLAC with block that extends past data.
	data := []byte("fLaC")
	data = append(data, 0x04)             // block type 4 (vorbis comment), not last
	data = append(data, 0x00, 0xFF, 0xFF) // block size = 65535
	data = append(data, 0x00, 0x00)       // only 2 bytes of payload

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseFLAC(data, result)
	// Should not panic.
}

func TestParseFLAC_LastBlock(t *testing.T) {
	// FLAC with isLast flag set on a non-vorbis block — should stop iteration.
	data := []byte("fLaC")
	data = append(data, 0x80|0x00)        // block type 0 (streaminfo), isLast=true
	data = append(data, 0x00, 0x00, 0x04) // block size = 4
	data = append(data, 0x00, 0x00, 0x00, 0x00)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseFLAC(data, result)
	// Should stop after first block; no vorbis comment found.
}

func TestParseFLAC_NotFLAC(t *testing.T) {
	data := []byte("notFLAC")
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseFLAC(data, result)
	// Should return immediately.
}

func TestParseWAV_ChunkSizeOverflow(t *testing.T) {
	// WAV with a chunk claiming more than available.
	data := []byte("RIFF")
	size := make([]byte, 4)
	// Little-endian file size (doesn't matter much for parsing).
	data = append(data, size...)
	data = append(data, []byte("WAVE")...)
	// fmt chunk with size way too large.
	data = append(data, []byte("fmt ")...)
	data = append(data, 0xFF, 0xFF, 0x00, 0x00) // chunk size = 65535
	data = append(data, make([]byte, 4)...)      // only 4 bytes of payload

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseWAV(data, result)
	// Should not panic.
}

func TestParseWAV_LISTWithoutINFO(t *testing.T) {
	// WAV with LIST chunk but without INFO sub-type.
	var data []byte
	data = append(data, []byte("RIFF")...)
	data = append(data, 0, 0, 0, 0) // file size
	data = append(data, []byte("WAVE")...)
	// LIST chunk with "JUNK" instead of "INFO"
	data = append(data, []byte("LIST")...)
	listPayload := []byte("JUNKsome data here")
	listSize := make([]byte, 4)
	listSize[0] = byte(len(listPayload))
	data = append(data, listSize...)
	data = append(data, listPayload...)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseWAV(data, result)
	// Should not panic, no metadata from JUNK.
}

func TestParseRIFFInfo_FieldOverflow(t *testing.T) {
	// RIFF INFO with a field claiming more than available.
	data := []byte("INAM")
	data = append(data, 0xFF, 0xFF, 0x00, 0x00) // size=65535
	data = append(data, []byte("short")...)      // only 5 bytes

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	parseRIFFInfo(data, result)
	// Should break safely.
}

func TestFindBytes_NotFound(t *testing.T) {
	data := []byte("hello world")
	got := findBytes(data, []byte("xyz"))
	if got != -1 {
		t.Errorf("findBytes(not found) = %d, want -1", got)
	}
}

func TestFindBytes_Found(t *testing.T) {
	data := []byte("hello world")
	got := findBytes(data, []byte("world"))
	if got != 6 {
		t.Errorf("findBytes(found) = %d, want 6", got)
	}
}

// --- readSyncsafe ---

func TestReadSyncsafe(t *testing.T) {
	// 0x00000080 in syncsafe = 0x01 0x00 = 128
	got := readSyncsafe([]byte{0x00, 0x00, 0x01, 0x00})
	if got != 128 {
		t.Errorf("readSyncsafe = %d, want 128", got)
	}
}

// --- cleanString ---

func TestCleanString(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"hello\x00world", "hello"},
		{"  spaces  ", "spaces"},
		{"normal", "normal"},
		{"\x00", ""},
	}
	for _, tt := range tests {
		got := cleanString(tt.input)
		if got != tt.want {
			t.Errorf("cleanString(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

// --- buildTextSummary ---

func TestBuildTextSummary_Empty(t *testing.T) {
	m := make(metadata.Metadata)
	got := buildTextSummary(m)
	if got != "" {
		t.Errorf("buildTextSummary(empty) = %q, want empty", got)
	}
}

func TestBuildTextSummary_WithData(t *testing.T) {
	m := make(metadata.Metadata)
	m.Set(metadata.KeyTitle, "Song")
	m.Set(metadata.KeyCreator, "Artist")
	got := buildTextSummary(m)
	if got == "" {
		t.Error("buildTextSummary should not be empty")
	}
}
