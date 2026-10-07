package parser

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestParse_BasicSlides(t *testing.T) {
	dir := t.TempDir()

	pres := "# My Presentation\n\nSlide 1 content\n\n---\n\nSlide 2 content\n\n---\n\nSlide 3 content\n"
	os.WriteFile(filepath.Join(dir, "pres.md"), []byte(pres), 0644)

	p, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if p.Title != "My Presentation" {
		t.Errorf("Title = %q, want %q", p.Title, "My Presentation")
	}

	if len(p.Slides) != 3 {
		t.Fatalf("len(Slides) = %d, want 3", len(p.Slides))
	}

	for i, s := range p.Slides {
		if s.Index != i {
			t.Errorf("Slides[%d].Index = %d, want %d", i, s.Index, i)
		}
	}
}

func TestParse_TitleFromH1(t *testing.T) {
	dir := t.TempDir()

	pres := "# The Big Talk\n\nSome intro text\n"
	os.WriteFile(filepath.Join(dir, "pres.md"), []byte(pres), 0644)

	p, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if p.Title != "The Big Talk" {
		t.Errorf("Title = %q, want %q", p.Title, "The Big Talk")
	}
}

func TestParse_AssetsDetected(t *testing.T) {
	dir := t.TempDir()

	// Minimal PNG (1x1 transparent).
	png := []byte{
		0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A,
		0x00, 0x00, 0x00, 0x0D, 0x49, 0x48, 0x44, 0x52,
	}
	os.WriteFile(filepath.Join(dir, "pres.md"), []byte("# Test\n"), 0644)
	os.WriteFile(filepath.Join(dir, "image.png"), png, 0644)

	p, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	asset, ok := p.Assets["image.png"]
	if !ok {
		t.Fatal("Assets[\"image.png\"] not found")
	}

	if !strings.Contains(asset.MimeType, "image/png") {
		t.Errorf("MimeType = %q, want something containing image/png", asset.MimeType)
	}
}

func TestParse_MermaidFileRef(t *testing.T) {
	dir := t.TempDir()

	mmd := "graph TD\n  A[Start] --> B[End]\n"
	os.WriteFile(filepath.Join(dir, "flow.mmd"), []byte(mmd), 0644)

	pres := "# Mermaid Test\n\n!mermaid[flow.mmd]\n"
	os.WriteFile(filepath.Join(dir, "pres.md"), []byte(pres), 0644)

	p, err := Parse(dir)
	if err != nil {
		t.Fatalf("Parse() returned error: %v", err)
	}

	if len(p.Mermaids) != 1 {
		t.Fatalf("len(Mermaids) = %d, want 1", len(p.Mermaids))
	}

	mmdCode, ok := p.Mermaids["flow.mmd"]
	if !ok {
		t.Fatal("Mermaids[\"flow.mmd\"] not found")
	}
	if mmdCode != mmd {
		t.Errorf("Mermaids[\"flow.mmd\"] = %q, want %q", mmdCode, mmd)
	}

	if len(p.Slides) != 1 {
		t.Fatalf("len(Slides) = %d, want 1", len(p.Slides))
	}

	// The !mermaid[...] reference should be removed from content.
	if strings.Contains(p.Slides[0].Content, "!mermaid") {
		t.Errorf("Content still contains !mermaid reference: %q", p.Slides[0].Content)
	}

	if p.Slides[0].Mermaid != mmd {
		t.Errorf("Slides[0].Mermaid = %q, want %q", p.Slides[0].Mermaid, mmd)
	}
}

func TestParse_EmptyDir(t *testing.T) {
	dir := t.TempDir()

	_, err := Parse(dir)
	if err == nil {
		t.Fatal("Parse() expected error for empty dir, got nil")
	}
}

func TestParse_NoMdFile(t *testing.T) {
	dir := t.TempDir()

	os.WriteFile(filepath.Join(dir, "notes.txt"), []byte("hello"), 0644)

	_, err := Parse(dir)
	if err == nil {
		t.Fatal("Parse() expected error for dir with no .md file, got nil")
	}
}