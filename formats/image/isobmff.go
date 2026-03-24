// CONTRACT:C1-EXTRACTOR.1.0
package image

import (
	"bytes"
	"encoding/binary"
	"fmt"

	goexif "github.com/rwcarlsen/goexif/exif"

	"github.com/dmorris/content-extractor/extractor"
	"github.com/dmorris/content-extractor/metadata"
)

// Safety limits for ISOBMFF parsing.
const (
	maxBoxCount  = 1000  // Maximum boxes to iterate.
	maxItemCount = 10000 // Maximum iloc items.
)

// isobmffBox represents a parsed ISOBMFF box.
type isobmffBox struct {
	boxType    string
	dataOffset int // Offset of box payload in data.
	dataSize   int // Size of box payload.
}

// extractISOBMFF parses ISOBMFF container data (HEIC, HEIF, AVIF) and extracts
// dimensions from ispe and EXIF metadata from the embedded EXIF item.
func extractISOBMFF(data []byte, result *extractor.Result) {
	boxes := walkBoxes(data, 0, len(data))

	// Find the meta box.
	var metaBox *isobmffBox
	for i := range boxes {
		if boxes[i].boxType == "meta" {
			metaBox = &boxes[i]
			break
		}
	}
	if metaBox == nil {
		return
	}

	// meta is a FullBox — skip 4 bytes of version/flags before children.
	metaStart := metaBox.dataOffset + 4
	metaEnd := metaBox.dataOffset + metaBox.dataSize
	if metaStart >= metaEnd || metaStart >= len(data) {
		return
	}

	metaChildren := walkBoxes(data, metaStart, metaEnd)

	// Extract dimensions from ispe (meta > iprp > ipco > ispe).
	extractISPE(data, metaChildren, result)

	// Extract EXIF via iloc + iinf.
	extractISOBMFFExif(data, metaChildren, result)
}

// walkBoxes iterates ISOBMFF boxes between start and end offsets.
func walkBoxes(data []byte, start, end int) []isobmffBox {
	if end > len(data) {
		end = len(data)
	}
	var boxes []isobmffBox
	pos := start

	for pos+8 <= end && len(boxes) < maxBoxCount {
		// Box header: 4 bytes size + 4 bytes type.
		size := int(binary.BigEndian.Uint32(data[pos : pos+4]))
		boxType := string(data[pos+4 : pos+8])

		headerSize := 8
		boxEnd := pos + size

		if size == 1 {
			// Extended size: 8-byte uint64.
			if pos+16 > end {
				break
			}
			extSize := binary.BigEndian.Uint64(data[pos+8 : pos+16])
			if extSize > uint64(end-pos) {
				break
			}
			boxEnd = pos + int(extSize)
			headerSize = 16
		} else if size == 0 {
			// Box extends to end of data.
			boxEnd = end
		} else if size < 8 {
			break // Invalid box size.
		}

		if boxEnd > end {
			boxEnd = end
		}

		boxes = append(boxes, isobmffBox{
			boxType:    boxType,
			dataOffset: pos + headerSize,
			dataSize:   boxEnd - pos - headerSize,
		})

		pos = boxEnd
	}

	return boxes
}

// extractISPE finds the ispe box inside meta > iprp > ipco and extracts dimensions.
func extractISPE(data []byte, metaChildren []isobmffBox, result *extractor.Result) {
	// Find iprp box.
	var iprp *isobmffBox
	for i := range metaChildren {
		if metaChildren[i].boxType == "iprp" {
			iprp = &metaChildren[i]
			break
		}
	}
	if iprp == nil {
		return
	}

	// Find ipco inside iprp.
	iprpChildren := walkBoxes(data, iprp.dataOffset, iprp.dataOffset+iprp.dataSize)
	var ipco *isobmffBox
	for i := range iprpChildren {
		if iprpChildren[i].boxType == "ipco" {
			ipco = &iprpChildren[i]
			break
		}
	}
	if ipco == nil {
		return
	}

	// Find ispe inside ipco.
	ipcoChildren := walkBoxes(data, ipco.dataOffset, ipco.dataOffset+ipco.dataSize)
	for _, box := range ipcoChildren {
		if box.boxType == "ispe" && box.dataSize >= 12 {
			// ispe payload: 4 bytes version/flags + 4 bytes width + 4 bytes height.
			off := box.dataOffset
			if off+12 > len(data) {
				continue
			}
			w := binary.BigEndian.Uint32(data[off+4 : off+8])
			h := binary.BigEndian.Uint32(data[off+8 : off+12])
			if w > 0 {
				result.Metadata.Set("exif:width", fmt.Sprintf("%d", w))
			}
			if h > 0 {
				result.Metadata.Set("exif:height", fmt.Sprintf("%d", h))
			}
			return
		}
	}
}

// ilocItem represents an item location entry from the iloc box.
type ilocItem struct {
	itemID uint16
	offset int
	size   int
}

// extractISOBMFFExif finds the EXIF item via iinf/infe, locates it via iloc,
// and extracts EXIF metadata using goexif.
func extractISOBMFFExif(data []byte, metaChildren []isobmffBox, result *extractor.Result) {
	// Find EXIF item ID from iinf box.
	exifItemID := findExifItemID(data, metaChildren)
	if exifItemID == 0 {
		return
	}

	// Find iloc box and parse item locations.
	items := parseILOC(data, metaChildren)
	if items == nil {
		return
	}

	// Look up the EXIF item.
	item, ok := items[exifItemID]
	if !ok {
		return
	}

	if item.offset+item.size > len(data) || item.size < 10 {
		return
	}

	exifBlob := data[item.offset : item.offset+item.size]

	// HEIC/AVIF EXIF items have a 4-byte prefix indicating the offset to the
	// TIFF header within the blob. Skip it to get to the actual TIFF data.
	if len(exifBlob) > 4 {
		tiffOffset := int(binary.BigEndian.Uint32(exifBlob[:4]))
		headerStart := 4 + tiffOffset
		if headerStart < len(exifBlob) {
			exifBlob = exifBlob[headerStart:]
		}
	}

	// Parse the EXIF/TIFF data using goexif.
	x, err := goexif.Decode(bytes.NewReader(exifBlob))
	if err != nil {
		return
	}

	// Reuse the same EXIF extraction logic as JPEG/TIFF.
	extractEXIFFromDecoded(x, result)
}

// findExifItemID scans the iinf box for an infe entry with item_type "Exif".
func findExifItemID(data []byte, metaChildren []isobmffBox) uint16 {
	var iinf *isobmffBox
	for i := range metaChildren {
		if metaChildren[i].boxType == "iinf" {
			iinf = &metaChildren[i]
			break
		}
	}
	if iinf == nil {
		return 0
	}

	// iinf is a FullBox: 4 bytes version/flags, then 2 bytes entry_count (v0) or 4 bytes (v1+).
	off := iinf.dataOffset
	end := iinf.dataOffset + iinf.dataSize
	if off+6 > end || off+6 > len(data) {
		return 0
	}

	version := data[off]
	var entryCountSize int
	if version == 0 {
		entryCountSize = 6 // 4 version/flags + 2 count
	} else {
		entryCountSize = 8 // 4 version/flags + 4 count
	}

	// Walk infe boxes inside iinf.
	infeStart := off + entryCountSize
	infeBoxes := walkBoxes(data, infeStart, end)

	for _, box := range infeBoxes {
		if box.boxType != "infe" || box.dataSize < 8 {
			continue
		}

		infeOff := box.dataOffset
		if infeOff+8 > len(data) {
			continue
		}

		infeVersion := data[infeOff]
		// infe v2+: 4 bytes version/flags, 2 bytes item_ID, 2 bytes item_protection_index, 4 bytes item_type.
		if infeVersion >= 2 {
			if infeOff+12 > len(data) || infeOff+12 > box.dataOffset+box.dataSize {
				continue
			}
			itemID := binary.BigEndian.Uint16(data[infeOff+4 : infeOff+6])
			itemType := string(data[infeOff+8 : infeOff+12])
			if itemType == "Exif" {
				return itemID
			}
		}
	}

	return 0
}

// parseILOC parses the iloc box and returns a map of item_id -> ilocItem.
func parseILOC(data []byte, metaChildren []isobmffBox) map[uint16]ilocItem {
	var iloc *isobmffBox
	for i := range metaChildren {
		if metaChildren[i].boxType == "iloc" {
			iloc = &metaChildren[i]
			break
		}
	}
	if iloc == nil {
		return nil
	}

	off := iloc.dataOffset
	end := iloc.dataOffset + iloc.dataSize
	if off+8 > end || off+8 > len(data) {
		return nil
	}

	version := data[off]
	// Byte at off+4: (offset_size << 4) | length_size
	// Byte at off+5: (base_offset_size << 4) | (index_size for v1/v2, reserved for v0)
	sizeByte1 := data[off+4]
	sizeByte2 := data[off+5]
	offsetSize := int(sizeByte1 >> 4)
	lengthSize := int(sizeByte1 & 0x0F)
	baseOffsetSize := int(sizeByte2 >> 4)

	pos := off + 6

	// Item count.
	var itemCount int
	if version < 2 {
		if pos+2 > end {
			return nil
		}
		itemCount = int(binary.BigEndian.Uint16(data[pos : pos+2]))
		pos += 2
	} else {
		if pos+4 > end {
			return nil
		}
		itemCount = int(binary.BigEndian.Uint32(data[pos : pos+4]))
		pos += 4
	}

	if itemCount > maxItemCount {
		return nil
	}

	items := make(map[uint16]ilocItem, itemCount)

	for i := 0; i < itemCount; i++ {
		// item_ID: 2 bytes (v0/v1) or 4 bytes (v2).
		var itemID uint16
		if version < 2 {
			if pos+2 > end {
				break
			}
			itemID = binary.BigEndian.Uint16(data[pos : pos+2])
			pos += 2
		} else {
			if pos+4 > end {
				break
			}
			itemID = uint16(binary.BigEndian.Uint32(data[pos : pos+4]))
			pos += 4
		}

		// construction_method: 2 bytes (v1/v2 only).
		if version == 1 || version == 2 {
			pos += 2
		}

		// data_reference_index: 2 bytes.
		pos += 2

		// base_offset: variable size.
		baseOffset := readILOCSize(data, pos, end, baseOffsetSize)
		pos += baseOffsetSize

		// extent_count: 2 bytes.
		if pos+2 > end {
			break
		}
		extentCount := int(binary.BigEndian.Uint16(data[pos : pos+2]))
		pos += 2

		// For simplicity, use the first extent.
		for j := 0; j < extentCount; j++ {
			// extent_index: variable (v1/v2 only, if index_size > 0).
			if version == 1 || version == 2 {
				indexSize := int(sizeByte2 & 0x0F)
				pos += indexSize
			}

			extentOffset := readILOCSize(data, pos, end, offsetSize)
			pos += offsetSize

			extentLength := readILOCSize(data, pos, end, lengthSize)
			pos += lengthSize

			if j == 0 {
				items[itemID] = ilocItem{
					itemID: itemID,
					offset: baseOffset + extentOffset,
					size:   extentLength,
				}
			}
		}
	}

	return items
}

// readILOCSize reads a variable-size integer (0, 2, or 4 bytes) from iloc data.
func readILOCSize(data []byte, pos, end, size int) int {
	switch size {
	case 2:
		if pos+2 > end || pos+2 > len(data) {
			return 0
		}
		return int(binary.BigEndian.Uint16(data[pos : pos+2]))
	case 4:
		if pos+4 > end || pos+4 > len(data) {
			return 0
		}
		return int(binary.BigEndian.Uint32(data[pos : pos+4]))
	case 8:
		if pos+8 > end || pos+8 > len(data) {
			return 0
		}
		return int(binary.BigEndian.Uint64(data[pos : pos+8]))
	default:
		return 0
	}
}

// extractEXIFFromDecoded extracts metadata from an already-decoded goexif object.
// This is shared logic between JPEG/TIFF and ISOBMFF EXIF extraction.
func extractEXIFFromDecoded(x *goexif.Exif, result *extractor.Result) {
	if tag, err := x.Get(goexif.Make); err == nil {
		result.Metadata.Set("exif:make", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.Model); err == nil {
		result.Metadata.Set("exif:model", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.DateTimeOriginal); err == nil {
		result.Metadata.Set(metadata.KeyDate, metadata.NormalizeDate(cleanTag(tag.String())))
	} else if tag, err := x.Get(goexif.DateTime); err == nil {
		result.Metadata.Set(metadata.KeyDate, metadata.NormalizeDate(cleanTag(tag.String())))
	}
	if tag, err := x.Get(goexif.Orientation); err == nil {
		result.Metadata.Set("exif:orientation", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.Software); err == nil {
		result.Metadata.Set(metadata.KeyApplication, cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.Artist); err == nil {
		result.Metadata.Set(metadata.KeyCreator, cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.Copyright); err == nil {
		result.Metadata.Set("exif:copyright", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.ImageDescription); err == nil {
		result.Metadata.Set(metadata.KeyDescription, cleanTag(tag.String()))
	}
	if lat, lon, err := x.LatLong(); err == nil {
		result.Metadata.Set("exif:gps-latitude", fmt.Sprintf("%.6f", lat))
		result.Metadata.Set("exif:gps-longitude", fmt.Sprintf("%.6f", lon))
	}
	if tag, err := x.Get(goexif.ExposureTime); err == nil {
		result.Metadata.Set("exif:exposure-time", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.FNumber); err == nil {
		result.Metadata.Set("exif:f-number", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.ISOSpeedRatings); err == nil {
		result.Metadata.Set("exif:iso", cleanTag(tag.String()))
	}
	if tag, err := x.Get(goexif.FocalLength); err == nil {
		result.Metadata.Set("exif:focal-length", cleanTag(tag.String()))
	}
}
