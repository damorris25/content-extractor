// CONTRACT:C1-EXTRACTOR.1.0
package audio

import (
	"fmt"
	"strings"

	"github.com/damorris25/content-extractor/extractor"
)

// parseOggVorbis extracts Vorbis comments from OGG files.
// OGG structure: pages with segment tables. Vorbis comments are in the
// second packet (comment header) after the identification header.
func parseOggVorbis(data []byte, result *extractor.Result) {
	// Find the Vorbis comment packet by scanning for the marker.
	// Vorbis comment header starts with: 0x03 + "vorbis"
	marker := []byte{0x03, 'v', 'o', 'r', 'b', 'i', 's'}
	idx := findBytes(data, marker)
	if idx < 0 {
		return
	}

	parseVorbisComment(data[idx+len(marker):], result)
}

// parseFLAC extracts Vorbis comments from FLAC files.
// FLAC structure: "fLaC" + metadata blocks. Block type 4 = Vorbis comment.
func parseFLAC(data []byte, result *extractor.Result) {
	if len(data) < 4 || string(data[:4]) != "fLaC" {
		return
	}

	pos := 4
	for pos+4 <= len(data) {
		header := data[pos]
		isLast := header&0x80 != 0
		blockType := header & 0x7F
		blockSize := int(data[pos+1])<<16 | int(data[pos+2])<<8 | int(data[pos+3])
		pos += 4

		if pos+blockSize > len(data) {
			break
		}

		if blockType == 4 { // Vorbis comment
			parseVorbisComment(data[pos:pos+blockSize], result)
			return
		}

		pos += blockSize
		if isLast {
			break
		}
	}
}

// parseVorbisComment parses a Vorbis comment block.
// Format: vendor_length(4 LE) + vendor_string + comment_count(4 LE) + comments
// Each comment: length(4 LE) + "KEY=VALUE"
func parseVorbisComment(data []byte, result *extractor.Result) {
	if len(data) < 8 {
		return
	}

	// Skip vendor string — validate uint32 before casting to int.
	vendorLen := readUint32LE(data, 0)
	if vendorLen > uint32(len(data)-4) {
		return
	}
	pos := 4 + int(vendorLen)
	if pos+4 > len(data) {
		return
	}

	commentCount := readUint32LE(data, pos)
	pos += 4

	// Cap comment count to prevent excessive iteration.
	const maxComments = 65536
	if commentCount > maxComments {
		commentCount = maxComments
	}

	for i := uint32(0); i < commentCount && pos+4 <= len(data); i++ {
		cl := readUint32LE(data, pos)
		pos += 4

		if cl > uint32(len(data)-pos) {
			break
		}
		commentLen := int(cl)

		comment := string(data[pos : pos+commentLen])
		pos += commentLen

		if eqIdx := strings.IndexByte(comment, '='); eqIdx > 0 {
			key := comment[:eqIdx]
			value := comment[eqIdx+1:]
			setAudioMeta(result.Metadata, key, value)
		}
	}
}

// parseWAV extracts basic metadata from WAV files.
// WAV is RIFF format: "RIFF" + size + "WAVE" + chunks.
// Look for INFO list chunks for metadata.
func parseWAV(data []byte, result *extractor.Result) {
	if len(data) < 12 || string(data[:4]) != "RIFF" || string(data[8:12]) != "WAVE" {
		return
	}

	// Parse WAV format chunk for basic audio info.
	pos := 12
	for pos+8 <= len(data) {
		chunkID := string(data[pos : pos+4])
		cs := readUint32LE(data, pos+4)
		pos += 8

		if cs > uint32(len(data)-pos) {
			break
		}
		chunkSize := int(cs)

		switch chunkID {
		case "fmt ":
			if chunkSize >= 16 {
				channels := readUint16LE(data, pos+2)
				sampleRate := readUint32LE(data, pos+4)
				bitsPerSample := readUint16LE(data, pos+14)
				result.Metadata.Set("audio:channels", strings.TrimSpace(fmt.Sprintf("%d", channels)))
				result.Metadata.Set("audio:sample-rate", strings.TrimSpace(fmt.Sprintf("%d", sampleRate)))
				result.Metadata.Set("audio:bits-per-sample", strings.TrimSpace(fmt.Sprintf("%d", bitsPerSample)))
			}
		case "LIST":
			if chunkSize >= 4 && string(data[pos:pos+4]) == "INFO" {
				parseRIFFInfo(data[pos+4:pos+chunkSize], result)
			}
		}

		pos += chunkSize
		if chunkSize%2 == 1 {
			pos++ // Chunks are word-aligned.
		}
	}
}

// parseRIFFInfo parses INFO list chunks from WAV files.
func parseRIFFInfo(data []byte, result *extractor.Result) {
	pos := 0
	for pos+8 <= len(data) {
		tag := string(data[pos : pos+4])
		s := readUint32LE(data, pos+4)
		pos += 8

		if s > uint32(len(data)-pos) {
			break
		}
		size := int(s)

		value := cleanString(string(data[pos : pos+size]))
		switch tag {
		case "INAM":
			setAudioMeta(result.Metadata, "TITLE", value)
		case "IART":
			setAudioMeta(result.Metadata, "ARTIST", value)
		case "IPRD":
			setAudioMeta(result.Metadata, "ALBUM", value)
		case "IGNR":
			setAudioMeta(result.Metadata, "GENRE", value)
		case "ICRD":
			setAudioMeta(result.Metadata, "DATE", value)
		case "ICMT":
			setAudioMeta(result.Metadata, "COMMENT", value)
		}

		pos += size
		if size%2 == 1 {
			pos++
		}
	}
}

// findBytes finds the first occurrence of needle in data.
func findBytes(data, needle []byte) int {
	for i := 0; i+len(needle) <= len(data); i++ {
		match := true
		for j := range needle {
			if data[i+j] != needle[j] {
				match = false
				break
			}
		}
		if match {
			return i
		}
	}
	return -1
}
