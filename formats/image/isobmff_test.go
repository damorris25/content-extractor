// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"encoding/binary"
	"testing"

	"github.com/damorris25/content-extractor/extractor"
	"github.com/damorris25/content-extractor/metadata"
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

// --- extractISOBMFFExif edge-case tests ---

// buildIINF constructs an iinf FullBox containing infe entries.
// version controls the iinf box version (0 = 2-byte count, 1+ = 4-byte count).
func buildIINF(version byte, infes [][]byte) []byte {
	var payload []byte
	if version == 0 {
		// 2-byte entry count
		cnt := make([]byte, 2)
		binary.BigEndian.PutUint16(cnt, uint16(len(infes)))
		payload = append(payload, cnt...)
	} else {
		// 4-byte entry count
		cnt := make([]byte, 4)
		binary.BigEndian.PutUint32(cnt, uint32(len(infes)))
		payload = append(payload, cnt...)
	}
	for _, infe := range infes {
		payload = append(payload, infe...)
	}
	return makeFullBox("iinf", version, payload)
}

// buildINFE constructs an infe FullBox.
// version controls the infe version.
// itemID is the item ID, itemType is the 4-byte item type (e.g. "Exif").
func buildINFE(version byte, itemID uint16, itemType string) []byte {
	var payload []byte
	if version >= 2 {
		// item_ID (2 bytes), item_protection_index (2 bytes), item_type (4 bytes)
		id := make([]byte, 2)
		binary.BigEndian.PutUint16(id, itemID)
		payload = append(payload, id...)
		payload = append(payload, 0x00, 0x00) // item_protection_index
		payload = append(payload, []byte(itemType)...)
	} else {
		// v0/v1: item_ID (2 bytes), item_protection_index (2 bytes) — no item_type
		id := make([]byte, 2)
		binary.BigEndian.PutUint16(id, itemID)
		payload = append(payload, id...)
		payload = append(payload, 0x00, 0x00)
	}
	return makeFullBox("infe", version, payload)
}

// buildILOC constructs an iloc FullBox (version 0) with given items.
// offsetSize and lengthSize control the variable-size fields (2 or 4).
func buildILOC(version byte, offsetSize, lengthSize, baseOffsetSize int, items []ilocItem) []byte {
	var payload []byte
	// sizes byte: (offsetSize << 4) | lengthSize
	payload = append(payload, byte(offsetSize<<4)|byte(lengthSize))
	// (baseOffsetSize << 4) | indexSize (0 for v0)
	indexSize := 0
	if version == 1 || version == 2 {
		indexSize = 0 // we use 0 for simplicity
	}
	payload = append(payload, byte(baseOffsetSize<<4)|byte(indexSize))

	// item count
	if version < 2 {
		cnt := make([]byte, 2)
		binary.BigEndian.PutUint16(cnt, uint16(len(items)))
		payload = append(payload, cnt...)
	} else {
		cnt := make([]byte, 4)
		binary.BigEndian.PutUint32(cnt, uint32(len(items)))
		payload = append(payload, cnt...)
	}

	for _, item := range items {
		// item_ID
		if version < 2 {
			id := make([]byte, 2)
			binary.BigEndian.PutUint16(id, item.itemID)
			payload = append(payload, id...)
		} else {
			id := make([]byte, 4)
			binary.BigEndian.PutUint32(id, uint32(item.itemID))
			payload = append(payload, id...)
		}

		// construction_method (v1/v2 only)
		if version == 1 || version == 2 {
			payload = append(payload, 0x00, 0x00)
		}

		// data_reference_index
		payload = append(payload, 0x00, 0x00)

		// base_offset (variable size)
		switch baseOffsetSize {
		case 2:
			payload = append(payload, 0x00, 0x00)
		case 4:
			payload = append(payload, 0x00, 0x00, 0x00, 0x00)
		}

		// extent_count = 1
		payload = append(payload, 0x00, 0x01)

		// extent_offset
		switch offsetSize {
		case 2:
			off := make([]byte, 2)
			binary.BigEndian.PutUint16(off, uint16(item.offset))
			payload = append(payload, off...)
		case 4:
			off := make([]byte, 4)
			binary.BigEndian.PutUint32(off, uint32(item.offset))
			payload = append(payload, off...)
		}

		// extent_length
		switch lengthSize {
		case 2:
			ln := make([]byte, 2)
			binary.BigEndian.PutUint16(ln, uint16(item.size))
			payload = append(payload, ln...)
		case 4:
			ln := make([]byte, 4)
			binary.BigEndian.PutUint32(ln, uint32(item.size))
			payload = append(payload, ln...)
		}
	}

	return makeFullBox("iloc", version, payload)
}

// buildMinimalExifBlob builds a minimal EXIF blob with the 4-byte TIFF offset prefix.
// tiffOffset is the offset from byte 4 to the TIFF header start.
// tiffHeader is the raw TIFF header bytes (e.g. MM\x00\x2a...).
func buildMinimalExifBlob(tiffOffset uint32, tiffHeader []byte) []byte {
	prefix := make([]byte, 4)
	binary.BigEndian.PutUint32(prefix, tiffOffset)
	padding := make([]byte, tiffOffset) // padding between prefix and TIFF header
	var blob []byte
	blob = append(blob, prefix...)
	blob = append(blob, padding...)
	blob = append(blob, tiffHeader...)
	return blob
}

func TestExtractISOBMFFExif_ValidExifItem(t *testing.T) {
	// Build a synthetic ISOBMFF structure: meta > { iinf > infe(Exif, id=1), iloc(item 1 -> EXIF blob) }
	// The EXIF blob has a 4-byte TIFF offset prefix + valid minimal TIFF header.

	// Minimal TIFF header: MM (big-endian), magic 42, IFD offset 8, 0 entries
	tiffHeader := []byte{
		'M', 'M', 0x00, 0x2A, // byte order + magic
		0x00, 0x00, 0x00, 0x08, // IFD0 offset = 8
		0x00, 0x00, // 0 IFD entries
	}
	// EXIF blob: 4-byte prefix (tiffOffset=6) + 6 bytes padding + TIFF header
	exifBlob := buildMinimalExifBlob(6, tiffHeader)

	infe := buildINFE(2, 1, "Exif")
	iinf := buildIINF(0, [][]byte{infe})

	// We'll place the EXIF blob at a known offset in the final data buffer.
	// First build iinf and iloc as meta children, then append the EXIF blob after.
	// The iloc must point to the absolute offset of the EXIF blob in the final buffer.

	// Build iloc pointing to a placeholder offset (we'll fix it after assembling).
	ilocItems := []ilocItem{{itemID: 1, offset: 0, size: len(exifBlob)}}
	iloc := buildILOC(0, 4, 4, 0, ilocItems)

	// Assemble: meta FullBox wrapping iinf + iloc, then exif blob appended at end.
	metaPayload := append([]byte{0, 0, 0, 0}, iinf...) // 4 bytes version/flags
	metaPayload = append(metaPayload, iloc...)

	metaBox := makeBox("meta", metaPayload)
	// EXIF blob offset is right after the meta box.
	exifOffset := len(metaBox)

	// Now fix the iloc to point to the correct offset.
	// Reassemble with correct offset.
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)

	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)
	// Recalculate since meta box size may have changed (it shouldn't, same iloc size).
	exifOffset = len(metaBox)
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)
	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)

	fullData := append(metaBox, exifBlob...)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFF(fullData, result)

	// The minimal TIFF has 0 IFD entries, so goexif won't extract metadata,
	// but the code path through extractISOBMFFExif should be fully exercised
	// without panic. No EXIF metadata expected from an empty IFD.
	// The key test is that extractISOBMFFExif runs through the full code path.
}

func TestFindExifItemID_InfeVersionLessThan2(t *testing.T) {
	// infe with version < 2 should be skipped (no item_type field).
	infe := buildINFE(0, 1, "Exif") // version 0 — no item_type parsed
	iinf := buildIINF(0, [][]byte{infe})
	allData := iinf
	children := walkBoxes(allData, 0, len(allData))

	id := findExifItemID(allData, children)
	if id != 0 {
		t.Errorf("findExifItemID with infe v0 = %d, want 0", id)
	}
}

func TestFindExifItemID_MultipleInfes(t *testing.T) {
	// Two infe entries: one non-Exif (v2, "hvc1") and one Exif (v2, "Exif").
	infe1 := buildINFE(2, 1, "hvc1")
	infe2 := buildINFE(2, 2, "Exif")
	iinf := buildIINF(0, [][]byte{infe1, infe2})
	allData := iinf
	children := walkBoxes(allData, 0, len(allData))

	id := findExifItemID(allData, children)
	if id != 2 {
		t.Errorf("findExifItemID = %d, want 2", id)
	}
}

func TestFindExifItemID_IinfVersion1(t *testing.T) {
	// iinf version 1 uses 4-byte entry count.
	infe := buildINFE(2, 5, "Exif")
	iinf := buildIINF(1, [][]byte{infe})
	allData := iinf
	children := walkBoxes(allData, 0, len(allData))

	id := findExifItemID(allData, children)
	if id != 5 {
		t.Errorf("findExifItemID with iinf v1 = %d, want 5", id)
	}
}

func TestParseILOC_Version1_ConstructionMethod(t *testing.T) {
	// iloc version 1 has a 2-byte construction_method field per item.
	ilocItems := []ilocItem{{itemID: 7, offset: 200, size: 80}}
	iloc := buildILOC(1, 4, 4, 0, ilocItems)
	allData := iloc
	children := walkBoxes(allData, 0, len(allData))

	items := parseILOC(allData, children)
	if items == nil {
		t.Fatal("parseILOC v1 returned nil")
	}
	item, ok := items[7]
	if !ok {
		t.Fatal("item 7 not found in iloc v1")
	}
	if item.offset != 200 {
		t.Errorf("item.offset = %d, want 200", item.offset)
	}
	if item.size != 80 {
		t.Errorf("item.size = %d, want 80", item.size)
	}
}

func TestParseILOC_Version2(t *testing.T) {
	// iloc version 2 has 4-byte item IDs and 4-byte item count.
	ilocItems := []ilocItem{{itemID: 3, offset: 500, size: 100}}
	iloc := buildILOC(2, 4, 4, 0, ilocItems)
	allData := iloc
	children := walkBoxes(allData, 0, len(allData))

	items := parseILOC(allData, children)
	if items == nil {
		t.Fatal("parseILOC v2 returned nil")
	}
	item, ok := items[3]
	if !ok {
		t.Fatal("item 3 not found in iloc v2")
	}
	if item.offset != 500 {
		t.Errorf("item.offset = %d, want 500", item.offset)
	}
	if item.size != 100 {
		t.Errorf("item.size = %d, want 100", item.size)
	}
}

func TestExtractISOBMFFExif_ExifBlobTooSmall(t *testing.T) {
	// EXIF item exists but blob is < 10 bytes — should be skipped.
	infe := buildINFE(2, 1, "Exif")
	iinf := buildIINF(0, [][]byte{infe})

	// Build a tiny EXIF blob (5 bytes < 10 minimum).
	exifBlob := []byte{0, 0, 0, 0, 0}

	ilocItems := []ilocItem{{itemID: 1, offset: 0, size: len(exifBlob)}}
	iloc := buildILOC(0, 4, 4, 0, ilocItems)

	metaPayload := append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox := makeBox("meta", metaPayload)

	// Fix iloc offset to point after meta box.
	exifOffset := len(metaBox)
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)
	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)
	exifOffset = len(metaBox)
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)
	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)

	fullData := append(metaBox, exifBlob...)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFF(fullData, result)

	// Should not panic; no EXIF metadata extracted.
	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty (blob too small)", got)
	}
}

func TestExtractISOBMFFExif_TiffOffsetPastEnd(t *testing.T) {
	// EXIF blob with tiffOffset pointing past the end of the blob.
	infe := buildINFE(2, 1, "Exif")
	iinf := buildIINF(0, [][]byte{infe})

	// EXIF blob: 4-byte prefix with tiffOffset=100 (way past 16-byte blob)
	exifBlob := make([]byte, 16)
	binary.BigEndian.PutUint32(exifBlob[:4], 100) // tiffOffset=100

	ilocItems := []ilocItem{{itemID: 1, offset: 0, size: len(exifBlob)}}
	iloc := buildILOC(0, 4, 4, 0, ilocItems)

	metaPayload := append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox := makeBox("meta", metaPayload)

	exifOffset := len(metaBox)
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)
	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)
	exifOffset = len(metaBox)
	ilocItems[0].offset = exifOffset
	iloc = buildILOC(0, 4, 4, 0, ilocItems)
	metaPayload = append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)
	metaBox = makeBox("meta", metaPayload)

	fullData := append(metaBox, exifBlob...)

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFF(fullData, result)

	// Should not panic; goexif.Decode will fail on garbage data.
	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty (tiffOffset past end)", got)
	}
}

func TestExtractISOBMFFExif_ItemNotInILOC(t *testing.T) {
	// iinf has Exif item ID=1, but iloc has no entry for item 1.
	infe := buildINFE(2, 1, "Exif")
	iinf := buildIINF(0, [][]byte{infe})

	// iloc with a different item (id=99)
	ilocItems := []ilocItem{{itemID: 99, offset: 0, size: 50}}
	iloc := buildILOC(0, 4, 4, 0, ilocItems)

	metaPayload := append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)

	allData := metaPayload
	// Walk as if these are metaChildren (skipping 4-byte version/flags of meta).
	children := walkBoxes(allData, 4, len(allData))

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFFExif(allData, children, result)

	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty (item not in iloc)", got)
	}
}

func TestExtractISOBMFFExif_ItemOffsetOutOfBounds(t *testing.T) {
	// iloc entry points past end of data.
	infe := buildINFE(2, 1, "Exif")
	iinf := buildIINF(0, [][]byte{infe})

	ilocItems := []ilocItem{{itemID: 1, offset: 99999, size: 50}}
	iloc := buildILOC(0, 4, 4, 0, ilocItems)

	metaPayload := append([]byte{0, 0, 0, 0}, iinf...)
	metaPayload = append(metaPayload, iloc...)

	allData := metaPayload
	children := walkBoxes(allData, 4, len(allData))

	result := &extractor.Result{Metadata: make(metadata.Metadata)}
	extractISOBMFFExif(allData, children, result)

	if got := result.Metadata.Get("exif:make"); got != "" {
		t.Errorf("exif:make = %q, want empty (offset out of bounds)", got)
	}
}
