package parser

// Slide represents a single presentation slide.
type Slide struct {
	Index   int
	Content string
	Mermaid string
}

// Asset represents an embedded file (image, video, audio).
type Asset struct {
	Name     string
	Data     []byte
	MimeType string
}

// Presentation holds the full parsed presentation.
type Presentation struct {
	Title    string
	Slides   []Slide
	Assets   map[string]Asset
	Mermaids map[string]string
}