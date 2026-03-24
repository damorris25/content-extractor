// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

func TestImage_HEIC_Dimensions(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.heic")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/heic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get("exif:width"); got != "100" {
		t.Errorf("width = %q, want %q", got, "100")
	}
	if got := result.Metadata.Get("exif:height"); got != "50" {
		t.Errorf("height = %q, want %q", got, "50")
	}
	if got := result.Metadata.Get(metadata.KeyFormat); got != "image/heic" {
		t.Errorf("format = %q, want %q", got, "image/heic")
	}
}

func TestImage_AVIF_Dimensions(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.avif")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/avif")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := result.Metadata.Get("exif:width"); got != "100" {
		t.Errorf("width = %q, want %q", got, "100")
	}
	if got := result.Metadata.Get("exif:height"); got != "50" {
		t.Errorf("height = %q, want %q", got, "50")
	}
	if got := result.Metadata.Get(metadata.KeyFormat); got != "image/avif" {
		t.Errorf("format = %q, want %q", got, "image/avif")
	}
}

func TestImage_HEIC_NoExif(t *testing.T) {
	// simple.heic has no EXIF tags (converted from simple.jpg which has none).
	ext := New()
	data := loadTestFile(t, "simple.heic")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/heic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// No EXIF means no make/model/date — only dimensions from ispe.
	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty", got)
	}
}

func TestImage_Compliance_HEIC(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "simple.heic")
	// Basic compliance: non-nil result, correct content type, non-nil metadata.
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/heic")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("result is nil")
	}
	if result.ContentType != "image/heic" {
		t.Errorf("ContentType = %q, want %q", result.ContentType, "image/heic")
	}
	if result.Metadata == nil {
		t.Error("Metadata is nil")
	}
}

// TestWalkBoxes_Synthetic tests the box walker with hand-crafted data.
func TestWalkBoxes_Synthetic(t *testing.T) {
	// Build two boxes: "abcd" (16 bytes payload) and "efgh" (8 bytes payload).
	var buf []byte

	// Box 1: size=24 (8 header + 16 payload), type "abcd"
	b1 := make([]byte, 4)
	binary.BigEndian.PutUint32(b1, 24)
	buf = append(buf, b1...)
	buf = append(buf, []byte("abcd")...)
	buf = append(buf, make([]byte, 16)...) // 16 bytes payload

	// Box 2: size=16 (8 header + 8 payload), type "efgh"
	b2 := make([]byte, 4)
	binary.BigEndian.PutUint32(b2, 16)
	buf = append(buf, b2...)
	buf = append(buf, []byte("efgh")...)
	buf = append(buf, make([]byte, 8)...) // 8 bytes payload

	boxes := walkBoxes(buf, 0, len(buf))
	if len(boxes) != 2 {
		t.Fatalf("walkBoxes returned %d boxes, want 2", len(boxes))
	}
	if boxes[0].boxType != "abcd" {
		t.Errorf("box[0].type = %q, want %q", boxes[0].boxType, "abcd")
	}
	if boxes[0].dataSize != 16 {
		t.Errorf("box[0].dataSize = %d, want 16", boxes[0].dataSize)
	}
	if boxes[1].boxType != "efgh" {
		t.Errorf("box[1].type = %q, want %q", boxes[1].boxType, "efgh")
	}
	if boxes[1].dataSize != 8 {
		t.Errorf("box[1].dataSize = %d, want 8", boxes[1].dataSize)
	}
}

// TestWalkBoxes_ExtendedSize tests 64-bit extended box size (size==1).
func TestWalkBoxes_ExtendedSize(t *testing.T) {
	// Extended size box: 4B size=1, 4B type, 8B extended size.
	var buf []byte
	b := make([]byte, 4)
	binary.BigEndian.PutUint32(b, 1) // size=1 means extended
	buf = append(buf, b...)
	buf = append(buf, []byte("test")...)
	ext := make([]byte, 8)
	binary.BigEndian.PutUint64(ext, 24) // total box size = 24
	buf = append(buf, ext...)
	buf = append(buf, make([]byte, 8)...) // 8 bytes payload (24 - 16 header)

	boxes := walkBoxes(buf, 0, len(buf))
	if len(boxes) != 1 {
		t.Fatalf("walkBoxes returned %d boxes, want 1", len(boxes))
	}
	if boxes[0].boxType != "test" {
		t.Errorf("box.type = %q, want %q", boxes[0].boxType, "test")
	}
	if boxes[0].dataSize != 8 {
		t.Errorf("box.dataSize = %d, want 8", boxes[0].dataSize)
	}
}

// TestWalkBoxes_Truncated tests graceful handling of truncated data.
func TestWalkBoxes_Truncated(t *testing.T) {
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", nil},
		{"too_short", []byte{0, 0, 0}},
		{"header_only", []byte{0, 0, 0, 8, 't', 'e', 's', 't'}},
		{"size_too_large", func() []byte {
			b := make([]byte, 12)
			binary.BigEndian.PutUint32(b, 100) // claims 100 bytes but only 12
			copy(b[4:], "test")
			return b
		}()},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			boxes := walkBoxes(tt.data, 0, len(tt.data))
			// Should not panic. May return 0 or 1 boxes depending on case.
			_ = boxes
		})
	}
}

// TestExtractISPE_Synthetic tests ispe extraction with hand-crafted ISOBMFF data.
func TestExtractISPE_Synthetic(t *testing.T) {
	// Build: meta > iprp > ipco > ispe(width=640, height=480)
	ispePayload := make([]byte, 12) // 4B version/flags + 4B width + 4B height
	binary.BigEndian.PutUint32(ispePayload[4:], 640)
	binary.BigEndian.PutUint32(ispePayload[8:], 480)
	ispeBox := makeBox("ispe", ispePayload)
	ipcoBox := makeBox("ipco", ispeBox)
	iprpBox := makeBox("iprp", ipcoBox)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	// extractISPE takes metaChildren which should contain iprp.
	allData := iprpBox
	children := walkBoxes(allData, 0, len(allData))
	extractISPE(allData, children, result)

	if got := result.Metadata.Get("exif:width"); got != "640" {
		t.Errorf("width = %q, want %q", got, "640")
	}
	if got := result.Metadata.Get("exif:height"); got != "480" {
		t.Errorf("height = %q, want %q", got, "480")
	}
}

// TestExtractISOBMFF_NoMeta tests that missing meta box is handled gracefully.
func TestExtractISOBMFF_NoMeta(t *testing.T) {
	// ftyp box only, no meta.
	ftyp := makeBox("ftyp", []byte("heic\x00\x00\x00\x00"))
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFF(ftyp, result)
	// Should not panic; no dimensions extracted.
	if got := result.Metadata.Get("exif:width"); got != "" {
		t.Errorf("width = %q, want empty", got)
	}
}

// makeBox constructs a minimal ISOBMFF box.
func makeBox(boxType string, payload []byte) []byte {
	size := 8 + len(payload)
	buf := make([]byte, size)
	binary.BigEndian.PutUint32(buf, uint32(size))
	copy(buf[4:8], boxType)
	copy(buf[8:], payload)
	return buf
}

// makeFullBox constructs an ISOBMFF FullBox (with 4 bytes version/flags).
func makeFullBox(boxType string, version byte, payload []byte) []byte {
	vf := []byte{version, 0, 0, 0}
	full := append(vf, payload...)
	return makeBox(boxType, full)
}

// --- readILOCSize tests ---

func TestReadILOCSize(t *testing.T) {
	data := []byte{0x00, 0x0A, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x02, 0x03}

	tests := []struct {
		name string
		pos  int
		size int
		want int
	}{
		{"size_0", 0, 0, 0},
		{"size_2_at_0", 0, 2, 0x000A},
		{"size_4_at_0", 0, 4, 0x000A0000},
		{"size_2_oob", 9, 2, 0}, // out of bounds
		{"size_4_oob", 8, 4, 0}, // out of bounds
		{"size_8_oob", 5, 8, 0}, // out of bounds for 8-byte read
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := readILOCSize(data, tt.pos, len(data), tt.size)
			if got != tt.want {
				t.Errorf("readILOCSize(pos=%d, size=%d) = %d, want %d", tt.pos, tt.size, got, tt.want)
			}
		})
	}
}

func TestReadILOCSize_8Byte(t *testing.T) {
	data := make([]byte, 16)
	binary.BigEndian.PutUint64(data[0:8], 12345678)
	got := readILOCSize(data, 0, len(data), 8)
	if got != 12345678 {
		t.Errorf("readILOCSize(size=8) = %d, want 12345678", got)
	}
}

// --- parseILOC tests ---

func TestParseILOC_Synthetic_V0(t *testing.T) {
	// Build a synthetic iloc box (version 0).
	// Format: [version/flags(4)] [offset_size|length_size(1)] [base_offset_size|0(1)]
	//         [item_count(2)] [items...]
	// Each item (v0): [item_id(2)] [data_ref_idx(2)] [base_offset(variable)] [extent_count(2)]
	//                  [extent_offset(variable)] [extent_length(variable)]

	var payload []byte
	// offset_size=4, length_size=4 → byte = 0x44
	// base_offset_size=0, index_size=0 → byte = 0x00
	payload = append(payload, 0x44, 0x00)
	// item_count = 1
	payload = append(payload, 0x00, 0x01)
	// item: item_id=42
	payload = append(payload, 0x00, 0x2A)
	// data_reference_index=0
	payload = append(payload, 0x00, 0x00)
	// base_offset (0 bytes, since base_offset_size=0) — nothing to append
	// extent_count=1
	payload = append(payload, 0x00, 0x01)
	// extent_offset (4 bytes) = 100
	off := make([]byte, 4)
	binary.BigEndian.PutUint32(off, 100)
	payload = append(payload, off...)
	// extent_length (4 bytes) = 50
	ln := make([]byte, 4)
	binary.BigEndian.PutUint32(ln, 50)
	payload = append(payload, ln...)

	ilocBox := makeFullBox("iloc", 0, payload)

	allData := ilocBox
	children := walkBoxes(allData, 0, len(allData))

	items := parseILOC(allData, children)
	if items == nil {
		t.Fatal("parseILOC returned nil")
	}
	item, ok := items[42]
	if !ok {
		t.Fatal("item 42 not found")
	}
	if item.offset != 100 {
		t.Errorf("item.offset = %d, want 100", item.offset)
	}
	if item.size != 50 {
		t.Errorf("item.size = %d, want 50", item.size)
	}
}

func TestParseILOC_NoIlocBox(t *testing.T) {
	ftyp := makeBox("ftyp", []byte("heic\x00\x00\x00\x00"))
	children := walkBoxes(ftyp, 0, len(ftyp))
	items := parseILOC(ftyp, children)
	if items != nil {
		t.Errorf("parseILOC with no iloc = %v, want nil", items)
	}
}

func TestParseILOC_TooManyItems(t *testing.T) {
	// iloc with item_count > maxItemCount.
	var payload []byte
	payload = append(payload, 0x44, 0x00) // offset_size=4, length_size=4
	// item_count = 20000 (> maxItemCount=10000)
	payload = append(payload, 0x4E, 0x20) // 20000 in big-endian uint16
	ilocBox := makeFullBox("iloc", 0, payload)
	children := walkBoxes(ilocBox, 0, len(ilocBox))
	items := parseILOC(ilocBox, children)
	if items != nil {
		t.Errorf("parseILOC with too many items should return nil")
	}
}

// --- extractISOBMFFExif tests ---

func TestExtractISOBMFFExif_NoIinf(t *testing.T) {
	// meta with iprp/ipco/ispe but no iinf — EXIF extraction should gracefully skip.
	ispePayload := make([]byte, 12)
	binary.BigEndian.PutUint32(ispePayload[4:], 100)
	binary.BigEndian.PutUint32(ispePayload[8:], 50)
	ispeBox := makeBox("ispe", ispePayload)
	ipcoBox := makeBox("ipco", ispeBox)
	iprpBox := makeBox("iprp", ipcoBox)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	children := walkBoxes(iprpBox, 0, len(iprpBox))
	extractISOBMFFExif(iprpBox, children, result)
	// No panic, no EXIF metadata.
	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty", got)
	}
}

func TestExtractISPE_NoIprp(t *testing.T) {
	// metaChildren with no iprp box.
	ftyp := makeBox("ftyp", []byte("heic\x00\x00\x00\x00"))
	children := walkBoxes(ftyp, 0, len(ftyp))
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISPE(ftyp, children, result)
	if got := result.Metadata.Get("exif:width"); got != "" {
		t.Errorf("width = %q, want empty", got)
	}
}

func TestExtractISPE_NoIpco(t *testing.T) {
	// iprp with no ipco inside.
	iprpBox := makeBox("iprp", makeBox("xxxx", []byte("data")))
	children := walkBoxes(iprpBox, 0, len(iprpBox))
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISPE(iprpBox, children, result)
	if got := result.Metadata.Get("exif:width"); got != "" {
		t.Errorf("width = %q, want empty", got)
	}
}

// --- walkBoxes edge cases ---

func TestWalkBoxes_SizeZero(t *testing.T) {
	// Box with size=0 means it extends to end of data.
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], 0) // size=0
	copy(buf[4:8], "test")
	// 8 bytes payload
	boxes := walkBoxes(buf, 0, len(buf))
	if len(boxes) != 1 {
		t.Fatalf("walkBoxes(size=0) returned %d boxes, want 1", len(boxes))
	}
	if boxes[0].dataSize != 8 {
		t.Errorf("box.dataSize = %d, want 8", boxes[0].dataSize)
	}
}

func TestWalkBoxes_InvalidSmallSize(t *testing.T) {
	// Box with size=4 (< 8 minimum) should break.
	buf := make([]byte, 16)
	binary.BigEndian.PutUint32(buf[0:4], 4)
	copy(buf[4:8], "test")
	boxes := walkBoxes(buf, 0, len(buf))
	if len(boxes) != 0 {
		t.Errorf("walkBoxes(size=4) returned %d boxes, want 0", len(boxes))
	}
}

func TestExtractISOBMFF_MetaTooSmall(t *testing.T) {
	// meta box with payload too small for FullBox skip.
	metaBox := makeBox("meta", []byte{0, 0}) // Only 2 bytes payload (need 4 for FullBox)
	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFF(metaBox, result)
	// Should not panic.
}
