// CONTRACT:C2-DETECT.1.0
package detect

// extMap maps file extensions to MIME types. Used as fallback when magic bytes
// don't match. Extensions include the leading dot and are lowercase.
var extMap = map[string]string{
	// Text
	".txt":  "text/plain",
	".csv":  "text/csv",
	".tsv":  "text/tab-separated-values",
	".html": "text/html",
	".htm":  "text/html",
	".xml":  "application/xml",
	".json": "application/json",

	// Office OOXML
	".docx": "application/vnd.openxmlformats-officedocument.wordprocessingml.document",
	".xlsx": "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	".pptx": "application/vnd.openxmlformats-officedocument.presentationml.presentation",

	// Office legacy
	".doc": "application/msword",
	".xls": "application/vnd.ms-excel",
	".ppt": "application/vnd.ms-powerpoint",

	// ODF
	".odt": "application/vnd.oasis.opendocument.text",
	".ods": "application/vnd.oasis.opendocument.spreadsheet",
	".odp": "application/vnd.oasis.opendocument.presentation",

	// PDF
	".pdf": "application/pdf",

	// Images
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
	".gif":  "image/gif",
	".bmp":  "image/bmp",
	".tif":  "image/tiff",
	".tiff": "image/tiff",
	".webp": "image/webp",

	// Audio
	".mp3":  "audio/mpeg",
	".wav":  "audio/wav",
	".ogg":  "audio/ogg",
	".flac": "audio/flac",

	// CAD
	".dxf": "image/vnd.dxf",
	".dwg": "image/vnd.dwg",
}
