// CONTRACT:C1-EXTRACTOR.1.0
package audio

import (
	"strings"

	"github.com/damorris25/content-extractor/extractor"
)

// parseID3v2 extracts metadata from ID3v2 tags in MP3 files.
// ID3v2 header: "ID3" + version(2 bytes) + flags(1 byte) + size(4 bytes syncsafe).
func parseID3v2(data []byte, result *extractor.Result) {
	if len(data) < 10 || string(data[:3]) != "ID3" {
		return
	}

	majorVersion := data[3]
	// flags := data[5]
	tagSize := readSyncsafe(data[6:10])

	if int(tagSize)+10 > len(data) {
		tagSize = uint32(len(data) - 10)
	}

	pos := 10
	end := pos + int(tagSize)

	switch majorVersion {
	case 3, 4:
		parseID3v2Frames(data, pos, end, majorVersion, result)
	case 2:
		parseID3v22Frames(data, pos, end, result)
	}
}

// parseID3v2Frames parses ID3v2.3/2.4 frames.
// Frame: ID(4 bytes) + size(4 bytes) + flags(2 bytes) + data.
func parseID3v2Frames(data []byte, pos, end int, version byte, result *extractor.Result) {
	for pos+10 <= end {
		frameID := string(data[pos : pos+4])
		if frameID[0] == 0 {
			break // Padding.
		}

		var frameSize uint32
		if version == 4 {
			frameSize = readSyncsafe(data[pos+4 : pos+8])
		} else {
			frameSize = readUint32BE(data, pos+4)
		}

		pos += 10 // Skip header.

		if frameSize == 0 || frameSize > uint32(end-pos) {
			break
		}

		fs := int(frameSize)
		frameData := data[pos : pos+fs]
		value := decodeTextFrame(frameData)
		setAudioMeta(result.Metadata, frameID, value)

		pos += fs
	}
}

// parseID3v22Frames parses ID3v2.2 frames (3-byte IDs, 3-byte sizes).
func parseID3v22Frames(data []byte, pos, end int, result *extractor.Result) {
	for pos+6 <= end {
		frameID := string(data[pos : pos+3])
		if frameID[0] == 0 {
			break
		}

		frameSize := int(data[pos+3])<<16 | int(data[pos+4])<<8 | int(data[pos+5])
		pos += 6

		if frameSize == 0 || pos+frameSize > end {
			break
		}

		frameData := data[pos : pos+frameSize]

		// Map v2.2 frame IDs to v2.3+ equivalents.
		mapped := mapV22Frame(frameID)
		value := decodeTextFrame(frameData)
		setAudioMeta(result.Metadata, mapped, value)

		pos += frameSize
	}
}

// mapV22Frame maps ID3v2.2 3-char frame IDs to v2.3+ 4-char equivalents.
func mapV22Frame(id string) string {
	switch id {
	case "TT2":
		return "TIT2"
	case "TP1":
		return "TPE1"
	case "TAL":
		return "TALB"
	case "TRK":
		return "TRCK"
	case "TYE":
		return "TYER"
	case "TCO":
		return "TCON"
	case "COM":
		return "COMM"
	default:
		return id
	}
}

// decodeTextFrame decodes an ID3v2 text frame.
// First byte is encoding: 0=ISO-8859-1, 1=UTF-16, 2=UTF-16BE, 3=UTF-8.
func decodeTextFrame(data []byte) string {
	if len(data) < 2 {
		return ""
	}

	encoding := data[0]
	payload := data[1:]

	switch encoding {
	case 0: // ISO-8859-1
		return cleanString(string(payload))
	case 3: // UTF-8
		return cleanString(string(payload))
	case 1: // UTF-16 with BOM
		return decodeUTF16(payload)
	case 2: // UTF-16BE without BOM
		return decodeUTF16BE(payload)
	default:
		return cleanString(string(payload))
	}
}

// decodeUTF16 decodes UTF-16 with BOM.
func decodeUTF16(data []byte) string {
	if len(data) < 2 {
		return ""
	}
	// Check BOM.
	bigEndian := true
	if data[0] == 0xFF && data[1] == 0xFE {
		bigEndian = false
		data = data[2:]
	} else if data[0] == 0xFE && data[1] == 0xFF {
		data = data[2:]
	}

	var b strings.Builder
	for i := 0; i+1 < len(data); i += 2 {
		var ch uint16
		if bigEndian {
			ch = uint16(data[i])<<8 | uint16(data[i+1])
		} else {
			ch = uint16(data[i]) | uint16(data[i+1])<<8
		}
		if ch == 0 {
			break
		}
		b.WriteRune(rune(ch))
	}
	return b.String()
}

// decodeUTF16BE decodes UTF-16BE without BOM.
func decodeUTF16BE(data []byte) string {
	var b strings.Builder
	for i := 0; i+1 < len(data); i += 2 {
		ch := uint16(data[i])<<8 | uint16(data[i+1])
		if ch == 0 {
			break
		}
		b.WriteRune(rune(ch))
	}
	return b.String()
}

// readSyncsafe reads a 4-byte syncsafe integer (7 bits per byte).
func readSyncsafe(data []byte) uint32 {
	return uint32(data[0])<<21 | uint32(data[1])<<14 |
		uint32(data[2])<<7 | uint32(data[3])
}

// cleanString removes null terminators and trims whitespace.
func cleanString(s string) string {
	if idx := strings.IndexByte(s, 0); idx >= 0 {
		s = s[:idx]
	}
	return strings.TrimSpace(s)
}
