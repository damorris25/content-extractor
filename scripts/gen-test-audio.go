//go:build ignore

// gen-test-audio generates minimal audio test fixtures with metadata.
package main

import (
	"encoding/binary"
	"os"
	"path/filepath"
)

func main() {
	dir := "testdata"
	os.MkdirAll(dir, 0o755)

	createMP3(filepath.Join(dir, "simple.mp3"))
	createOGG(filepath.Join(dir, "simple.ogg"))
	createFLAC(filepath.Join(dir, "simple.flac"))
	createWAV(filepath.Join(dir, "simple.wav"))
}

func createMP3(path string) {
	// Create a minimal MP3 with ID3v2.3 tag.
	var buf []byte

	// ID3v2.3 header.
	buf = append(buf, 'I', 'D', '3')     // ID3 marker
	buf = append(buf, 3, 0)               // Version 2.3.0
	buf = append(buf, 0)                  // Flags
	buf = append(buf, 0, 0, 0, 0)         // Size placeholder (syncsafe)

	tagStart := len(buf)

	// TIT2 frame (Title).
	buf = appendID3Frame(buf, "TIT2", "Test Song")

	// TPE1 frame (Artist).
	buf = appendID3Frame(buf, "TPE1", "Test Artist")

	// TALB frame (Album).
	buf = appendID3Frame(buf, "TALB", "Test Album")

	// TRCK frame (Track).
	buf = appendID3Frame(buf, "TRCK", "3")

	// TYER frame (Year).
	buf = appendID3Frame(buf, "TYER", "2026")

	// TCON frame (Genre).
	buf = appendID3Frame(buf, "TCON", "Electronic")

	tagSize := len(buf) - tagStart

	// Write syncsafe size.
	buf[6] = byte((tagSize >> 21) & 0x7F)
	buf[7] = byte((tagSize >> 14) & 0x7F)
	buf[8] = byte((tagSize >> 7) & 0x7F)
	buf[9] = byte(tagSize & 0x7F)

	// Add a minimal MP3 frame (silence) so it's recognizable.
	// MPEG1 Layer3 frame header: 0xFF 0xFB 0x90 0x00
	buf = append(buf, 0xFF, 0xFB, 0x90, 0x00)
	// Pad with zeros for frame body.
	buf = append(buf, make([]byte, 413)...) // 417 byte frame for 128kbps

	os.WriteFile(path, buf, 0o644)
}

func appendID3Frame(buf []byte, id, value string) []byte {
	// Frame: ID(4) + Size(4 BE) + Flags(2) + Encoding(1) + Data
	data := []byte{0x03} // UTF-8 encoding
	data = append(data, []byte(value)...)

	buf = append(buf, []byte(id)...)
	size := make([]byte, 4)
	binary.BigEndian.PutUint32(size, uint32(len(data)))
	buf = append(buf, size...)
	buf = append(buf, 0, 0) // Flags
	buf = append(buf, data...)
	return buf
}

func createOGG(path string) {
	// Create a minimal OGG file with Vorbis comment header.
	var buf []byte

	// OGG page header (identification).
	buf = append(buf, oggPageHeader(0, 2, 0, 1)...) // BOS flag
	// Vorbis identification header.
	vorbisID := []byte{0x01, 'v', 'o', 'r', 'b', 'i', 's'}
	vorbisID = append(vorbisID, make([]byte, 23)...) // Minimum ID header fields
	buf[len(buf)-1] = byte(len(vorbisID))            // Segment table
	buf = append(buf, vorbisID...)

	// OGG page header (comment).
	page2Start := len(buf)
	buf = append(buf, oggPageHeader(0, 0, 1, 1)...)
	// Vorbis comment header.
	var comment []byte
	comment = append(comment, 0x03, 'v', 'o', 'r', 'b', 'i', 's')
	// Vendor string.
	vendor := "test"
	comment = appendLE32(comment, uint32(len(vendor)))
	comment = append(comment, vendor...)
	// Comment count.
	comments := []string{
		"TITLE=OGG Test Song",
		"ARTIST=OGG Test Artist",
		"ALBUM=OGG Test Album",
		"TRACKNUMBER=5",
		"DATE=2026",
		"GENRE=Ambient",
	}
	comment = appendLE32(comment, uint32(len(comments)))
	for _, c := range comments {
		comment = appendLE32(comment, uint32(len(c)))
		comment = append(comment, c...)
	}
	buf[page2Start+26] = 1                   // Number of segments
	buf[page2Start+27] = byte(len(comment))  // Segment size
	buf = append(buf, comment...)

	os.WriteFile(path, buf, 0o644)
}

func oggPageHeader(granulePos uint64, headerType byte, pageSeqNo uint32, segments int) []byte {
	var buf []byte
	buf = append(buf, 'O', 'g', 'g', 'S') // Capture pattern
	buf = append(buf, 0)                   // Version
	buf = append(buf, headerType)          // Header type
	// Granule position (8 bytes LE).
	gp := make([]byte, 8)
	binary.LittleEndian.PutUint64(gp, granulePos)
	buf = append(buf, gp...)
	// Serial number (4 bytes LE).
	buf = appendLE32(buf, 1)
	// Page sequence number.
	buf = appendLE32(buf, pageSeqNo)
	// CRC checksum (4 bytes, leave as 0 for test).
	buf = appendLE32(buf, 0)
	// Number of segments.
	buf = append(buf, byte(segments))
	// Segment table (filled by caller).
	buf = append(buf, make([]byte, segments)...)
	return buf
}

func createFLAC(path string) {
	var buf []byte

	// FLAC marker.
	buf = append(buf, 'f', 'L', 'a', 'C')

	// Metadata block: STREAMINFO (type 0, not last).
	// Header: 1 byte (flags+type) + 3 bytes (size) = 4 bytes.
	buf = append(buf, 0x00)           // Type 0 (STREAMINFO), not last
	buf = append(buf, 0, 0, 34)      // Block body size = 34 bytes
	buf = append(buf, make([]byte, 34)...) // STREAMINFO body (zeros for test)

	// Metadata block: Vorbis comment (type 4, last).
	var comment []byte
	vendor := "test"
	comment = appendLE32(comment, uint32(len(vendor)))
	comment = append(comment, vendor...)
	comments := []string{
		"TITLE=FLAC Test Song",
		"ARTIST=FLAC Test Artist",
		"ALBUM=FLAC Test Album",
		"TRACKNUMBER=7",
		"DATE=2025",
		"GENRE=Classical",
	}
	comment = appendLE32(comment, uint32(len(comments)))
	for _, c := range comments {
		comment = appendLE32(comment, uint32(len(c)))
		comment = append(comment, c...)
	}

	// Block header: type 4 (Vorbis comment) + last flag.
	buf = append(buf, 0x84) // 0x80 (last) | 0x04 (type)
	buf = append(buf, byte(len(comment)>>16), byte(len(comment)>>8), byte(len(comment)))
	buf = append(buf, comment...)

	os.WriteFile(path, buf, 0o644)
}

func createWAV(path string) {
	var buf []byte

	// Format chunk data.
	fmtData := make([]byte, 16)
	binary.LittleEndian.PutUint16(fmtData[0:], 1)      // PCM
	binary.LittleEndian.PutUint16(fmtData[2:], 2)      // Stereo
	binary.LittleEndian.PutUint32(fmtData[4:], 44100)  // Sample rate
	binary.LittleEndian.PutUint32(fmtData[8:], 176400) // Byte rate
	binary.LittleEndian.PutUint16(fmtData[12:], 4)     // Block align
	binary.LittleEndian.PutUint16(fmtData[14:], 16)    // Bits per sample

	// INFO list with metadata.
	var info []byte
	info = append(info, 'I', 'N', 'F', 'O')
	info = appendRIFFTag(info, "INAM", "WAV Test Song")
	info = appendRIFFTag(info, "IART", "WAV Test Artist")
	info = appendRIFFTag(info, "IPRD", "WAV Test Album")
	info = appendRIFFTag(info, "IGNR", "Jazz")

	// Minimal data chunk (silence).
	dataChunk := make([]byte, 100)

	// Calculate total RIFF size.
	totalSize := 4 + // "WAVE"
		8 + len(fmtData) + // fmt chunk
		8 + len(info) + // LIST chunk
		8 + len(dataChunk) // data chunk

	// RIFF header.
	buf = append(buf, 'R', 'I', 'F', 'F')
	buf = appendLE32(buf, uint32(totalSize))
	buf = append(buf, 'W', 'A', 'V', 'E')

	// fmt chunk.
	buf = append(buf, 'f', 'm', 't', ' ')
	buf = appendLE32(buf, uint32(len(fmtData)))
	buf = append(buf, fmtData...)

	// LIST chunk.
	buf = append(buf, 'L', 'I', 'S', 'T')
	buf = appendLE32(buf, uint32(len(info)))
	buf = append(buf, info...)

	// data chunk.
	buf = append(buf, 'd', 'a', 't', 'a')
	buf = appendLE32(buf, uint32(len(dataChunk)))
	buf = append(buf, dataChunk...)

	os.WriteFile(path, buf, 0o644)
}

func appendRIFFTag(buf []byte, id, value string) []byte {
	buf = append(buf, []byte(id)...)
	val := append([]byte(value), 0) // null-terminated
	buf = appendLE32(buf, uint32(len(val)))
	buf = append(buf, val...)
	if len(val)%2 == 1 {
		buf = append(buf, 0) // word-align
	}
	return buf
}

func appendLE32(buf []byte, v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return append(buf, b...)
}
