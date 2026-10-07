package parser

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func Parse(dir string) (*Presentation, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, fmt.Errorf("reading directory %s: %w", dir, err)
	}

	if len(entries) == 0 {
		return nil, fmt.Errorf("directory %s is empty", dir)
	}

	rootMD := findRootMD(entries)
	if rootMD == "" {
		return nil, fmt.Errorf("no .md file found in %s", dir)
	}

	rootPath := filepath.Join(dir, rootMD)
	content, err := os.ReadFile(rootPath)
	if err != nil {
		return nil, fmt.Errorf("reading root markdown %s: %w", rootPath, err)
	}

	title := extractTitle(string(content))
	if title == "" {
		title = strings.TrimSuffix(rootMD, filepath.Ext(rootMD))
	}

	slideChunks := splitSlides(string(content))
	slides := make([]Slide, 0, len(slideChunks))
	for i, chunk := range slideChunks {
		s := Slide{Index: i, Content: chunk}
		s.Content, s.Mermaid = resolveMermaidRefs(s.Content, dir)
		slides = append(slides, s)
	}

	mermaids, err := loadMermaids(entries, dir)
	if err != nil {
		return nil, err
	}

	assets, err := loadAssets(entries, dir)
	if err != nil {
		return nil, err
	}

	return &Presentation{
		Title:    title,
		Slides:   slides,
		Assets:   assets,
		Mermaids: mermaids,
	}, nil
}

func loadMermaids(entries []os.DirEntry, dir string) (map[string]string, error) {
	mermaids := make(map[string]string)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(e.Name())) != ".mmd" {
			continue
		}
		mmdPath := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(mmdPath)
		if err != nil {
			return nil, fmt.Errorf("reading mermaid file %s: %w", mmdPath, err)
		}
		mermaids[e.Name()] = string(data)
	}
	return mermaids, nil
}

func loadAssets(entries []os.DirEntry, dir string) (map[string]Asset, error) {
	assets := make(map[string]Asset)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(e.Name()))
		if ext == ".md" || ext == ".mmd" {
			continue
		}
		assetPath := filepath.Join(dir, e.Name())
		data, err := os.ReadFile(assetPath)
		if err != nil {
			return nil, fmt.Errorf("reading asset %s: %w", assetPath, err)
		}
		assets[e.Name()] = Asset{
			Name:     e.Name(),
			Data:     data,
			MimeType: detectMimeType(e.Name(), data),
		}
	}
	return assets, nil
}

func findRootMD(entries []os.DirEntry) string {
	var firstMD string
	has := map[string]string{}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if strings.ToLower(filepath.Ext(e.Name())) != ".md" {
			continue
		}
		name := e.Name()
		if firstMD == "" {
			firstMD = name
		}
		has[strings.ToLower(name)] = name
	}
	for _, key := range []string{"pres.md", "index.md", "slides.md"} {
		if v, ok := has[key]; ok {
			return v
		}
	}
	return firstMD
}

func extractTitle(content string) string {
	for _, line := range strings.Split(content, "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed, ok := strings.CutPrefix(trimmed, "# "); ok {
			return strings.TrimSpace(trimmed)
		}
	}
	return ""
}

func splitSlides(content string) []string {
	lines := strings.Split(content, "\n")
	var chunks []string
	var current strings.Builder
	inFence := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(trimmed, "```") {
			inFence = !inFence
		}
		if trimmed == "---" && !inFence {
			chunks = append(chunks, strings.TrimSpace(current.String()))
			current.Reset()
			continue
		}
		current.WriteString(line)
		current.WriteByte('\n')
	}

	if s := strings.TrimSpace(current.String()); s != "" {
		chunks = append(chunks, s)
	}

	var result []string
	for _, c := range chunks {
		if strings.TrimSpace(c) != "" {
			result = append(result, c)
		}
	}
	if len(result) == 0 {
		result = []string{""}
	}
	return result
}

var mermaidRefRE = regexp.MustCompile(`!mermaid\[([^\]]+)\]`)

func resolveMermaidRefs(content, dir string) (string, string) {
	matches := mermaidRefRE.FindAllStringSubmatch(content, -1)
	if len(matches) == 0 {
		return content, ""
	}

	cleaned := mermaidRefRE.ReplaceAllString(content, "")
	filename := matches[0][1]

	if filepath.IsAbs(filename) || strings.Contains(filename, "..") {
		return cleaned, ""
	}

	mmdPath := filepath.Join(dir, filename)
	data, err := os.ReadFile(mmdPath)
	if err != nil {
		return cleaned, ""
	}

	return strings.TrimSpace(cleaned), string(data)
}