package builder

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/saravenpi/catalyst/internal/parser"
	"github.com/saravenpi/catalyst/internal/renderer"
)

type Options struct {
	Dir    string
	OutDir string
}

func Build(opts Options) error {
	p, err := parser.Parse(opts.Dir)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", opts.Dir, err)
	}

	b, err := renderer.Render(p, true)
	if err != nil {
		return fmt.Errorf("rendering: %w", err)
	}

	if err := os.MkdirAll(opts.OutDir, 0755); err != nil {
		return fmt.Errorf("creating %s: %w", opts.OutDir, err)
	}

	idxPath := filepath.Join(opts.OutDir, "index.html")
	if err := os.WriteFile(idxPath, b, 0644); err != nil {
		return fmt.Errorf("writing index.html: %w", err)
	}

	count := 1

	mermaidURL := "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.esm.min.mjs"
	mermaidPath := filepath.Join(opts.OutDir, "mermaid.esm.min.mjs")
	if err := atomicDownload(mermaidURL, mermaidPath); err != nil {
		fmt.Printf("! Could not download mermaid.js (%v) — mermaid diagrams need a network connection\n", err)
	} else {
		count++
	}

	fmt.Printf("▸ Building to %s/\n", opts.OutDir)
	fmt.Printf("✓ Built %d files to %s/\n", count, opts.OutDir)

	return nil
}

func atomicDownload(url, path string) error {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	tmp := path + ".catalyst-dl"
	f, err := os.Create(tmp)
	if err != nil {
		return err
	}
	if _, err := io.Copy(f, resp.Body); err != nil {
		f.Close()
		os.Remove(tmp)
		return err
	}
	f.Close()

	return os.Rename(tmp, path)
}