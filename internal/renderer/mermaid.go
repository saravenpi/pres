package renderer

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
)

type mermaidBlock struct {
	code string
}

var mermaidFenceRE = regexp.MustCompile("(?s)```mermaid\\s*\\n(.*?)```")

func writeMermaid(buf *bytes.Buffer, offline bool, script []byte, theme string) {
	if len(script) > 0 {
		buf.WriteString("<script>\n")
		buf.Write(script)
		buf.WriteString("\n</script>\n")
	} else {
		buf.WriteString(
			"<script src=\"https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js\"></script>\n",
		)
	}
	buf.WriteString(mermaidInit(theme))
}

func mermaidInit(theme string) string {
	return fmt.Sprintf(`<script>
if (typeof mermaid !== 'undefined') {
 mermaid.initialize({
  startOnLoad: true,
  theme: 'base',
  securityLevel: 'loose',
  themeVariables: %s
 });
}
</script>
`, mermaidThemeVariables(theme))
}

func mermaidThemeVariables(theme string) string {
	if theme == "dark" {
		return `{
   background: 'transparent',
   primaryColor: '#7c5cff',
   primaryTextColor: '#e8e8e8',
   primaryBorderColor: '#5a3fd4',
   lineColor: '#7c5cff',
   secondaryColor: '#1f1f1f',
   tertiaryColor: '#141414',
   textColor: '#e8e8e8',
   fontSize: '16px'
  }`
	}
	return `{
   background: 'transparent',
   primaryColor: '#f0f0f0',
   primaryTextColor: '#141414',
   primaryBorderColor: '#7c5cff',
   lineColor: '#7c5cff',
   secondaryColor: '#ececec',
   tertiaryColor: '#ffffff',
   textColor: '#141414',
   fontSize: '16px'
  }`
}

func extractMermaidBlocks(content string) (string, []mermaidBlock) {
	var blocks []mermaidBlock
	cleaned := mermaidFenceRE.ReplaceAllStringFunc(content, func(match string) string {
		sub := mermaidFenceRE.FindStringSubmatch(match)
		if len(sub) >= 2 {
			blocks = append(blocks, mermaidBlock{code: strings.TrimSpace(sub[1])})
		}
		return ""
	})
	return cleaned, blocks
}
