package parser

import (
	"net/http"
	"path/filepath"
	"strings"
)

func detectMimeType(name string, data []byte) string {
	mime := http.DetectContentType(data)

	if mime == "text/plain; charset=utf-8" || len(data) == 0 {
		switch strings.ToLower(filepath.Ext(name)) {
		case ".png":
			return "image/png"
		case ".jpg", ".jpeg":
			return "image/jpeg"
		case ".gif":
			return "image/gif"
		case ".svg":
			return "image/svg+xml"
		case ".webp":
			return "image/webp"
		case ".mp4":
			return "video/mp4"
		case ".webm":
			return "video/webm"
		case ".mov":
			return "video/quicktime"
		case ".mp3":
			return "audio/mpeg"
		case ".wav":
			return "audio/wav"
		case ".ogg":
			return "audio/ogg"
		case ".pdf":
			return "application/pdf"
		}
	}

	return mime
}