// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"context"
	"encoding/binary"
	"math"
	"testing"

	"github.com/dmorris/content-extractor/metadata"
)

func TestImage_GeoTIFF_Metadata(t *testing.T) {
	ext := New()
	data := loadTestFile(t, "geotiff-sample.tif")
	result, err := ext.Extract(context.Background(), bytes.NewReader(data), "image/tiff")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	tests := []struct {
		key  string
		want string
	}{
		{"geo:crs", "WGS 84"},
		{"geo:pixel-scale", "0.001000,0.001000"},
		{"geo:bounds", "-122.000000,36.996000,-121.996000,37.000000"},
		{"exif:width", "4"},
		{"exif:height", "4"},
	}
	for _, tt := range tests {
		if got := result.Metadata.Get(tt.key); got != tt.want {
			t.Errorf("Metadata[%q] = %q, want %q", tt.key, got, tt.want)
		}
	}

	// Text summary should include geo info.
	if result.Text == "" {
		t.Error("Text summary is empty, expected geo metadata in summary")
	}
}

func TestImage_RegularTIFF_NoGeoMetadata(t *testing.T) {
	// Non-GeoTIFF data should produce no geo: keys.
	md := make(map[string][]string)
	// Call extractGeoTIFF with non-GeoTIFF data.
	extractGeoTIFF([]byte{}, md)
	if len(md) != 0 {
		t.Errorf("expected no geo metadata from empty data, got %v", md)
	}
}

func TestGeoTIFF_TruncatedData(t *testing.T) {
	data := loadTestFile(t, "geotiff-sample.tif")
	tests := []struct {
		name string
		data []byte
	}{
		{"empty", []byte{}},
		{"too_short", []byte{0x49, 0x49}},
		{"truncated_header", data[:7]},
		{"truncated_ifd", data[:12]},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			md := make(map[string][]string)
			extractGeoTIFF(tt.data, md)
			// Should not panic or produce geo metadata.
			if md["geo:crs"] != nil || md["geo:bounds"] != nil || md["geo:pixel-scale"] != nil {
				t.Errorf("expected no geo metadata from truncated data, got %v", md)
			}
		})
	}
}

func TestGeoTIFF_BigEndian(t *testing.T) {
	// Verify that a big-endian TIFF header is handled (even if no geo tags).
	md := make(map[string][]string)
	// Minimal big-endian TIFF: MM, 42, IFD offset=8, 0 entries
	data := []byte{
		'M', 'M', 0x00, 0x2A, // Byte order + magic
		0x00, 0x00, 0x00, 0x08, // IFD0 offset = 8
		0x00, 0x00, // 0 entries
	}
	extractGeoTIFF(data, md)
	// Should not panic or produce metadata.
	if len(md) != 0 {
		t.Errorf("expected no metadata from empty IFD, got %v", md)
	}
}

// --- extractCRS unit tests ---

func TestExtractCRS_GeogCitation(t *testing.T) {
	// GeoKeyDirectory with GeogCitationGeoKey (2049) pointing to GeoAsciiParams.
	geoKeys := []uint16{
		1, 1, 0, 1, // version, revision, minor, numKeys=1
		2049, tagGeoAsciiParams, 7, 0, // GeogCitationGeoKey, loc=GeoAscii, count=7, offset=0
	}
	geoAscii := "WGS 84|"
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "WGS 84" {
		t.Errorf("extractCRS = %q, want %q", crs, "WGS 84")
	}
}

func TestExtractCRS_PipeDelimitedGeoAscii(t *testing.T) {
	// GeoAsciiParams with multiple pipe-delimited strings.
	// The citation key points to a substring within the larger string.
	geoAscii := "NAD27|UTM Zone 10N|"
	geoKeys := []uint16{
		1, 1, 0, 1, // version, revision, minor, numKeys=1
		2049, tagGeoAsciiParams, 5, 6, // GeogCitationGeoKey, offset=6, count=5
	}
	crs := extractCRS(geoKeys, geoAscii)
	// Should extract "UTM Z" (5 chars starting at offset 6).
	// Actually: geoAscii[6:11] = "UTM Z"
	if crs != "UTM Z" {
		t.Errorf("extractCRS pipe-delimited = %q, want %q", crs, "UTM Z")
	}
}

func TestExtractCRS_FullCitationExtraction(t *testing.T) {
	// Citation pointing to full string.
	geoAscii := "NAD83 / UTM zone 10N|"
	geoKeys := []uint16{
		1, 1, 0, 1,
		2049, tagGeoAsciiParams, 20, 0, // count=20, offset=0
	}
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "NAD83 / UTM zone 10N" {
		t.Errorf("extractCRS = %q, want %q", crs, "NAD83 / UTM zone 10N")
	}
}

func TestExtractCRS_EmptyGeoKeyDirectory(t *testing.T) {
	// Empty geoKeys (fewer than 4 entries).
	crs := extractCRS(nil, "WGS 84")
	if crs != "" {
		t.Errorf("extractCRS(nil keys) = %q, want empty", crs)
	}

	crs = extractCRS([]uint16{1, 1, 0}, "WGS 84")
	if crs != "" {
		t.Errorf("extractCRS(3 keys) = %q, want empty", crs)
	}
}

func TestExtractCRS_NoCitationFallsBackToGeoAscii(t *testing.T) {
	// GeoKeyDirectory with keys but no GeogCitationGeoKey — falls back to geoAscii.
	geoKeys := []uint16{
		1, 1, 0, 1,
		1024, 0, 1, 1, // GTModelTypeGeoKey=1 (Geographic), stored inline
	}
	geoAscii := "WGS 84"
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "WGS 84" {
		t.Errorf("extractCRS fallback = %q, want %q", crs, "WGS 84")
	}
}

func TestExtractCRS_NoCitationNoGeoAscii(t *testing.T) {
	// No citation key and no geoAscii string.
	geoKeys := []uint16{
		1, 1, 0, 1,
		1024, 0, 1, 2, // GTModelTypeGeoKey=2 (Projected), stored inline
	}
	crs := extractCRS(geoKeys, "")
	if crs != "" {
		t.Errorf("extractCRS no ascii = %q, want empty", crs)
	}
}

func TestExtractCRS_CitationOffsetOutOfBounds(t *testing.T) {
	// Citation key offset+count exceeds geoAscii length — should fall back to full geoAscii.
	geoKeys := []uint16{
		1, 1, 0, 1,
		2049, tagGeoAsciiParams, 100, 0, // count=100 but geoAscii is short
	}
	geoAscii := "WGS 84"
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "WGS 84" {
		t.Errorf("extractCRS OOB fallback = %q, want %q", crs, "WGS 84")
	}
}

func TestExtractCRS_MultipleGeoKeys(t *testing.T) {
	// Multiple keys including GTModelTypeGeoKey and GeogCitationGeoKey.
	geoKeys := []uint16{
		1, 1, 0, 3, // 3 keys
		1024, 0, 1, 1,                    // GTModelTypeGeoKey = 1 (Geographic)
		2048, 0, 1, 4326,                 // GeographicTypeGeoKey = 4326
		2049, tagGeoAsciiParams, 6, 0,    // GeogCitationGeoKey
	}
	geoAscii := "WGS 84"
	crs := extractCRS(geoKeys, geoAscii)
	if crs != "WGS 84" {
		t.Errorf("extractCRS multi-key = %q, want %q", crs, "WGS 84")
	}
}

func TestExtractCRS_TruncatedKeyEntry(t *testing.T) {
	// numKeys says 2 but only 1 full key entry fits.
	geoKeys := []uint16{
		1, 1, 0, 2, // claims 2 keys
		1024, 0, 1, 1, // only 1 key fits
		// second key would need indices 8,9,10,11 but geoKeys only has 8 elements
	}
	// Should not panic; processes only the keys that fit.
	crs := extractCRS(geoKeys, "")
	if crs != "" {
		t.Errorf("extractCRS truncated = %q, want empty", crs)
	}
}

// --- Synthetic GeoTIFF builder for extractGeoTIFF tests ---

// buildSyntheticGeoTIFF builds a minimal little-endian TIFF with the given IFD entries.
// Each entry is an ifdEntry struct. Double values are stored in an external data area.
type ifdEntry struct {
	tag      uint16
	typ      uint16
	count    uint32
	// For inline values (SHORT/LONG with total <= 4 bytes), use inlineVal.
	inlineVal uint32
	// For external data (DOUBLE arrays, long SHORT arrays, ASCII), use extData.
	extData []byte
}

func buildSyntheticGeoTIFF(entries []ifdEntry) []byte {
	bo := binary.LittleEndian

	// TIFF header: II, 42, IFD offset=8
	header := make([]byte, 8)
	header[0], header[1] = 'I', 'I'
	bo.PutUint16(header[2:], 42)
	bo.PutUint32(header[4:], 8) // IFD at offset 8

	// IFD: 2-byte count + 12 bytes per entry + 4-byte next IFD offset (0)
	numEntries := len(entries)
	ifdSize := 2 + numEntries*12 + 4
	ifd := make([]byte, ifdSize)
	bo.PutUint16(ifd[0:], uint16(numEntries))

	// External data starts after header + IFD
	extDataOffset := 8 + ifdSize
	var extDataBuf []byte

	for i, e := range entries {
		off := 2 + i*12
		bo.PutUint16(ifd[off:], e.tag)
		bo.PutUint16(ifd[off+2:], e.typ)
		bo.PutUint32(ifd[off+4:], e.count)

		totalBytes := typSize(e.typ) * int(e.count)
		if e.extData != nil && totalBytes > 4 {
			// Store offset to external data
			bo.PutUint32(ifd[off+8:], uint32(extDataOffset+len(extDataBuf)))
			extDataBuf = append(extDataBuf, e.extData...)
		} else {
			// Inline value
			bo.PutUint32(ifd[off+8:], e.inlineVal)
		}
	}

	// Next IFD offset = 0 (no more IFDs)
	bo.PutUint32(ifd[2+numEntries*12:], 0)

	var data []byte
	data = append(data, header...)
	data = append(data, ifd...)
	data = append(data, extDataBuf...)
	return data
}

func typSize(typ uint16) int {
	switch typ {
	case 2: // ASCII
		return 1
	case 3: // SHORT
		return 2
	case 4: // LONG
		return 4
	case 12: // DOUBLE
		return 8
	}
	return 1
}

func leDoubles(vals ...float64) []byte {
	buf := make([]byte, len(vals)*8)
	for i, v := range vals {
		binary.LittleEndian.PutUint64(buf[i*8:], math.Float64bits(v))
	}
	return buf
}

func leShorts(vals ...uint16) []byte {
	buf := make([]byte, len(vals)*2)
	for i, v := range vals {
		binary.LittleEndian.PutUint16(buf[i*2:], v)
	}
	return buf
}

func TestExtractGeoTIFF_GeographicModelType(t *testing.T) {
	// Synthetic GeoTIFF with:
	// - ImageWidth=10, ImageHeight=10
	// - GeoKeyDirectory with GTModelTypeGeoKey=1 (Geographic) + GeogCitationGeoKey
	// - GeoAsciiParams = "WGS 84|"
	// - PixelScale and Tiepoint for bounds computation

	geoKeyData := leShorts(
		1, 1, 0, 2, // version, revision, minor, numKeys=2
		1024, 0, 1, 1, // GTModelTypeGeoKey = 1 (Geographic)
		2049, tagGeoAsciiParams, 6, 0, // GeogCitationGeoKey
	)
	geoAsciiData := []byte("WGS 84|\x00")
	pixelScaleData := leDoubles(0.01, 0.01, 0.0)
	tiepointData := leDoubles(0, 0, 0, -122.0, 37.0, 0)

	entries := []ifdEntry{
		{tag: 256, typ: 3, count: 1, inlineVal: 10},  // ImageWidth
		{tag: 257, typ: 3, count: 1, inlineVal: 10},  // ImageHeight
		{tag: tagModelPixelScale, typ: 12, count: 3, extData: pixelScaleData},
		{tag: tagModelTiepoint, typ: 12, count: 6, extData: tiepointData},
		{tag: tagGeoKeyDirectory, typ: 3, count: uint32(len(geoKeyData) / 2), extData: geoKeyData},
		{tag: tagGeoAsciiParams, typ: 2, count: uint32(len(geoAsciiData)), extData: geoAsciiData},
	}

	data := buildSyntheticGeoTIFF(entries)
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)

	if got := md.Get("geo:crs"); got != "WGS 84" {
		t.Errorf("geo:crs = %q, want %q", got, "WGS 84")
	}
	if got := md.Get("geo:pixel-scale"); got != "0.010000,0.010000" {
		t.Errorf("geo:pixel-scale = %q, want %q", got, "0.010000,0.010000")
	}
	if got := md.Get("geo:bounds"); got != "-122.000000,36.900000,-121.900000,37.000000" {
		t.Errorf("geo:bounds = %q, want %q", got, "-122.000000,36.900000,-121.900000,37.000000")
	}
	if got := md.Get("exif:width"); got != "10" {
		t.Errorf("exif:width = %q, want %q", got, "10")
	}
	if got := md.Get("exif:height"); got != "10" {
		t.Errorf("exif:height = %q, want %q", got, "10")
	}
}

func TestExtractGeoTIFF_ProjectedModelType(t *testing.T) {
	// GTModelTypeGeoKey=2 (Projected) with citation.
	geoKeyData := leShorts(
		1, 1, 0, 2,
		1024, 0, 1, 2, // GTModelTypeGeoKey = 2 (Projected)
		2049, tagGeoAsciiParams, 18, 0,
	)
	geoAsciiData := []byte("NAD83 / UTM 10N|\x00\x00")

	entries := []ifdEntry{
		{tag: 256, typ: 4, count: 1, inlineVal: 256}, // ImageWidth as LONG
		{tag: 257, typ: 4, count: 1, inlineVal: 256}, // ImageHeight as LONG
		{tag: tagGeoKeyDirectory, typ: 3, count: uint32(len(geoKeyData) / 2), extData: geoKeyData},
		{tag: tagGeoAsciiParams, typ: 2, count: uint32(len(geoAsciiData)), extData: geoAsciiData},
	}

	data := buildSyntheticGeoTIFF(entries)
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)

	if got := md.Get("geo:crs"); got != "NAD83 / UTM 10N" {
		t.Errorf("geo:crs = %q, want %q", got, "NAD83 / UTM 10N")
	}
	if got := md.Get("exif:width"); got != "256" {
		t.Errorf("exif:width = %q, want %q", got, "256")
	}
}

func TestExtractGeoTIFF_NoGeoKeyDirectory(t *testing.T) {
	// TIFF with pixel scale and tiepoints but no GeoKeyDirectory — no CRS.
	pixelScaleData := leDoubles(0.5, 0.5, 0.0)
	tiepointData := leDoubles(0, 0, 0, 10.0, 50.0, 0)

	entries := []ifdEntry{
		{tag: 256, typ: 3, count: 1, inlineVal: 100},
		{tag: 257, typ: 3, count: 1, inlineVal: 100},
		{tag: tagModelPixelScale, typ: 12, count: 3, extData: pixelScaleData},
		{tag: tagModelTiepoint, typ: 12, count: 6, extData: tiepointData},
	}

	data := buildSyntheticGeoTIFF(entries)
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)

	if got := md.Get("geo:crs"); got != "" {
		t.Errorf("geo:crs = %q, want empty (no GeoKeyDirectory)", got)
	}
	// Bounds should still be computed from pixel scale + tiepoints.
	if got := md.Get("geo:bounds"); got != "10.000000,0.000000,60.000000,50.000000" {
		t.Errorf("geo:bounds = %q, want %q", got, "10.000000,0.000000,60.000000,50.000000")
	}
}

func TestExtractGeoTIFF_MissingGeoDoubleParams(t *testing.T) {
	// GeoKeyDirectory present but no GeoAsciiParams or GeoDoubleParams.
	geoKeyData := leShorts(
		1, 1, 0, 1,
		1024, 0, 1, 1, // GTModelTypeGeoKey = 1 (Geographic), inline
	)

	entries := []ifdEntry{
		{tag: 256, typ: 3, count: 1, inlineVal: 4},
		{tag: 257, typ: 3, count: 1, inlineVal: 4},
		{tag: tagGeoKeyDirectory, typ: 3, count: uint32(len(geoKeyData) / 2), extData: geoKeyData},
	}

	data := buildSyntheticGeoTIFF(entries)
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)

	// No geoAscii, no citation key pointing to ascii → empty CRS.
	if got := md.Get("geo:crs"); got != "" {
		t.Errorf("geo:crs = %q, want empty (no GeoAsciiParams)", got)
	}
}

func TestExtractGeoTIFF_DimensionsNotOverwritten(t *testing.T) {
	// If exif:width is already set, extractGeoTIFF should not overwrite it.
	entries := []ifdEntry{
		{tag: 256, typ: 3, count: 1, inlineVal: 100},
		{tag: 257, typ: 3, count: 1, inlineVal: 200},
	}

	data := buildSyntheticGeoTIFF(entries)
	md := make(metadata.Metadata)
	md.Set("exif:width", "999")
	md.Set("exif:height", "888")
	extractGeoTIFF(data, md)

	if got := md.Get("exif:width"); got != "999" {
		t.Errorf("exif:width = %q, want %q (should not overwrite)", got, "999")
	}
	if got := md.Get("exif:height"); got != "888" {
		t.Errorf("exif:height = %q, want %q (should not overwrite)", got, "888")
	}
}

func TestExtractGeoTIFF_InvalidByteOrder(t *testing.T) {
	// Data with invalid byte order marker.
	data := []byte{'X', 'X', 0x00, 0x2A, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00}
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)
	if len(md) != 0 {
		t.Errorf("expected no metadata from invalid byte order, got %v", md)
	}
}

func TestExtractGeoTIFF_InvalidMagic(t *testing.T) {
	// Valid byte order but wrong magic number (99 instead of 42).
	data := []byte{'I', 'I', 0x63, 0x00, 0x00, 0x00, 0x00, 0x08, 0x00, 0x00}
	md := make(metadata.Metadata)
	extractGeoTIFF(data, md)
	if len(md) != 0 {
		t.Errorf("expected no metadata from invalid magic, got %v", md)
	}
}
