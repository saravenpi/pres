package renderer

import (
	"strings"
	"testing"

	"github.com/saravenpi/pres/internal/parser"
)

func TestRender_BasicSlide(t *testing.T) {
	p := &parser.Presentation{
		Title: "Test",
		Slides: []parser.Slide{
			{Index: 0, Content: "Hello world"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if !strings.Contains(html, "Hello world") {
		t.Error("output does not contain slide content")
	}
	if !strings.Contains(html, "<title>Test</title>") {
		t.Error("output does not contain title")
	}
	if !strings.Contains(html, "Helvetica") {
		t.Error("output does not use Helvetica font")
	}
}

func TestRender_MultipleSlides(t *testing.T) {
	p := &parser.Presentation{
		Title: "Multi",
		Slides: []parser.Slide{
			{Index: 0, Content: "Slide 1"},
			{Index: 1, Content: "Slide 2"},
			{Index: 2, Content: "Slide 3"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if strings.Count(html, "<section class=\"slide") != 3 {
		t.Errorf("expected 3 slide sections, got %d",
			strings.Count(html, "<section class=\"slide"))
	}
	if !strings.Contains(html, "counter") {
		t.Error("output should contain counter element")
	}
}

func TestRender_MermaidDiagram(t *testing.T) {
	p := &parser.Presentation{
		Title: "Mermaid",
		Slides: []parser.Slide{
			{Index: 0, Content: "Look at this:", Mermaid: "graph TD\n  A --> B"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if !strings.Contains(html, "class=\"mermaid\"") {
		t.Error("output should contain mermaid pre block")
	}
}

func TestRender_Offline(t *testing.T) {
	p := &parser.Presentation{
		Title: "Offline",
		Slides: []parser.Slide{
			{Index: 0, Content: "Offline slide"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	script := []byte("/* mermaid.js */")
	result, err := Render(p, true, script, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if strings.Contains(html, "cdn.jsdelivr.net") {
		t.Error("offline mode with inline script should not contain CDN URL")
	}
	if !strings.Contains(html, "/* mermaid.js */") {
		t.Error("offline mode should inline the mermaid script")
	}
	if strings.Contains(html, "EventSource") {
		t.Error("offline mode should not contain live reload")
	}
}

func TestRender_InlinedMermaidFence(t *testing.T) {
	p := &parser.Presentation{
		Title: "Fence",
		Slides: []parser.Slide{
			{Index: 0, Content: "```mermaid\ngraph LR\n  X --> Y\n```"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if !strings.Contains(html, "class=\"mermaid\"") {
		t.Error("output should contain mermaid block from fenced code")
	}
}

func TestRender_AssetInlining(t *testing.T) {
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	}
	p := &parser.Presentation{
		Title: "Assets",
		Slides: []parser.Slide{
			{Index: 0, Content: "![alt](cat.png)"},
		},
		Assets: map[string]parser.Asset{
			"cat.png": {Name: "cat.png", Data: png, MimeType: "image/png"},
		},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	if !strings.Contains(html, "data:image/png;base64,") {
		t.Error("output should contain base64 data URI for image")
	}
}

func TestRender_KeyboardNavScript(t *testing.T) {
	p := &parser.Presentation{
		Title: "Nav",
		Slides: []parser.Slide{
			{Index: 0, Content: "Test"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	result, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render() error: %v", err)
	}

	html := string(result)
	for _, key := range []string{"ArrowRight", "ArrowLeft", "e.code==='Space'"} {
		if !strings.Contains(html, key) {
			t.Errorf("keyboard nav script missing %s", key)
		}
	}
}

func TestRender_Theme(t *testing.T) {
	p := &parser.Presentation{
		Title: "Theme",
		Slides: []parser.Slide{
			{Index: 0, Content: "Theme slide"},
		},
		Assets:   map[string]parser.Asset{},
		Mermaids: map[string]string{},
	}

	dark, err := Render(p, false, nil, "dark")
	if err != nil {
		t.Fatalf("Render(dark) error: %v", err)
	}
	if !strings.Contains(string(dark), `data-theme="dark"`) {
		t.Error(`dark render missing data-theme="dark"`)
	}

	light, err := Render(p, false, nil, "light")
	if err != nil {
		t.Fatalf("Render(light) error: %v", err)
	}
	if !strings.Contains(string(light), `data-theme="light"`) {
		t.Error(`light render missing data-theme="light"`)
	}

	empty, err := Render(p, false, nil, "")
	if err != nil {
		t.Fatalf("Render(empty) error: %v", err)
	}
	if !strings.Contains(string(empty), `data-theme="light"`) {
		t.Error(`empty theme should normalize to data-theme="light"`)
	}
}
