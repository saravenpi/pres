package builder

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/saravenpi/pres/internal/parser"
	"github.com/saravenpi/pres/internal/renderer"
)

type Options struct {
	Dir  string
	Out  string
}

const mermaidURL = "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"

func Build(opts Options) error {
	p, err := parser.Parse(opts.Dir)
	if err != nil {
		return fmt.Errorf("parsing %s: %w", opts.Dir, err)
	}

	mermaidScript, err := downloadMermaid()
	if err != nil {
		return fmt.Errorf("downloading mermaid.js: %w", err)
	}

	b, err := renderer.Render(p, true, mermaidScript)
	if err != nil {
		return fmt.Errorf("rendering: %w", err)
	}

	out := opts.Out
	if out == "" {
		out = "pres.html"
	}

	if err := os.WriteFile(out, b, 0644); err != nil {
		return fmt.Errorf("writing %s: %w", out, err)
	}

	fmt.Printf("✓ Built %s\n", out)
	return nil
}

func downloadMermaid() ([]byte, error) {
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Get(mermaidURL)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP %d", resp.StatusCode)
	}

	return io.ReadAll(resp.Body)
}