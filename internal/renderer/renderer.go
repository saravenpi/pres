package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/saravenpi/catalyst/embed"
	"github.com/saravenpi/catalyst/internal/parser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

// WithUnsafe is intentionally left on: Catalyst is a local CLI tool, not a
// web service. The user controls every byte of their markdown and opens the
// rendered HTML in their own browser — same threat model as Hugo, Marp,
// and every other static site generator.

var md = goldmark.New(goldmark.WithExtensions(extension.GFM), goldmark.WithRendererOptions(goldmarkhtml.WithUnsafe()))

type mermaidBlock struct {
	code string
}

var mermaidFenceRE = regexp.MustCompile("(?s)```mermaid\\s*\\n(.*?)```")

var assetRefRE = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

func Render(p *parser.Presentation, offline bool) ([]byte, error) {
	var buf bytes.Buffer

	css, err := buildCSS()
	if err != nil {
		return nil, fmt.Errorf("building CSS: %w", err)
	}

	mermaidSrc := "https://cdn.jsdelivr.net/npm/mermaid@11/dist/mermaid.min.js"
	if offline {
		mermaidSrc = "./mermaid.min.js"
	}

	buf.WriteString("<!DOCTYPE html>\n<html lang=\"en\">\n<head>\n")
	buf.WriteString("<meta charset=\"UTF-8\">\n")
	buf.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&buf, "<title>%s</title>\n", stdhtml.EscapeString(p.Title))
	buf.WriteString("<style>\n")
	buf.WriteString(css)
	buf.WriteString("\n</style>\n</head>\n<body>\n")

	buf.WriteString("<div id=\"slides\">\n")
	for i, slide := range p.Slides {
		fmt.Fprintf(&buf,
			"<section class=\"slide%s\" id=\"slide-%d\">\n",
			activeClass(i), i,
		)
		buf.WriteString("<div class=\"slide-content\">\n")

		if slide.Mermaid != "" {
			buf.WriteString("<pre class=\"mermaid\">\n")
			buf.WriteString(stdhtml.EscapeString(slide.Mermaid))
			buf.WriteString("\n</pre>\n")
		}

		content := slide.Content
		content, mBlocks := extractMermaidBlocks(content)
		for _, mb := range mBlocks {
			buf.WriteString("<pre class=\"mermaid\">\n")
			buf.WriteString(stdhtml.EscapeString(mb.code))
			buf.WriteString("\n</pre>\n")
		}

		content = inlineAssets(content, p.Assets)

		var htmlBuf bytes.Buffer
		if err := md.Convert([]byte(content), &htmlBuf); err != nil {
			return nil, fmt.Errorf("rendering slide %d: %w", i, err)
		}
		buf.WriteString(htmlBuf.String())

		buf.WriteString("</div>\n</section>\n")
	}
	buf.WriteString("</div>\n")

	buf.WriteString("<div id=\"counter\" class=\"indicator\"></div>\n")

	fmt.Fprintf(&buf, mermaidTemplate, mermaidSrc)

	buf.WriteString(kbNavScript)
	if !offline {
		buf.WriteString(liveReloadScript)
	}

	buf.WriteString("</body>\n</html>\n")
	return buf.Bytes(), nil
}

const mermaidTemplate = `<script src="%s"></script>
<script>
mermaid.initialize({ startOnLoad: true, theme: 'dark', securityLevel: 'loose' });
</script>
`

const kbNavScript = `<script>
(function(){
var cur=0,slides=document.querySelectorAll('.slide'),
counter=document.getElementById('counter');
function show(n){
 if(slides.length===0)return;
 for(var i=0;i<slides.length;i++)slides[i].classList.remove('active');
 cur=((n%slides.length)+slides.length)%slides.length;
 slides[cur].classList.add('active');
 if(counter)counter.textContent=(cur+1)+' / '+slides.length;
}
document.addEventListener('keydown',function(e){
 if(e.code==='Space'||e.key==='ArrowRight'||e.key==='ArrowDown'){
  e.preventDefault();show(cur+1);
 }else if(e.key==='ArrowLeft'||e.key==='ArrowUp'){
  e.preventDefault();show(cur-1);
 }
});
show(0);
})();
</script>
`

const liveReloadScript = `<script>
new EventSource('/__catalyst_reload')
  .addEventListener('reload',function(){location.reload()});
</script>
`

func buildCSS() (string, error) {
	raw, err := embed.Assets.ReadFile("default.css")
	if err != nil {
		return "", fmt.Errorf("reading default.css: %w", err)
	}

	fontData, err := embed.Assets.ReadFile("fonts/undefined-medium.woff2")
	if err != nil {
		return "", fmt.Errorf("reading font: %w", err)
	}
	fontB64 := base64.StdEncoding.EncodeToString(fontData)

	fontFace := fmt.Sprintf(
		"@font-face { font-family: 'undefined-medium'; "+
			"src: url(data:font/woff2;base64,%s) format('woff2'); "+
			"font-weight: 400; font-style: normal; }\n",
		fontB64,
	)

	return fontFace + string(raw), nil
}

func activeClass(i int) string {
	if i == 0 {
		return " active"
	}
	return ""
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

func inlineAssets(content string, assets map[string]parser.Asset) string {
	return assetRefRE.ReplaceAllStringFunc(content, func(match string) string {
		sub := assetRefRE.FindStringSubmatch(match)
		if len(sub) < 3 {
			return match
		}
		alt := sub[1]
		filename := sub[2]

		if strings.Contains(filename, "://") {
			return match
		}

		asset, ok := assets[filename]
		if !ok {
			return match
		}

		b64 := base64.StdEncoding.EncodeToString(asset.Data)
		src := fmt.Sprintf("data:%s;base64,%s", asset.MimeType, b64)

		switch {
		case strings.HasPrefix(asset.MimeType, "video/"):
			return fmt.Sprintf(
				"<video src=\"%s\" controls></video>",
				stdhtml.EscapeString(src),
			)
		case strings.HasPrefix(asset.MimeType, "audio/"):
			return fmt.Sprintf(
				"<audio src=\"%s\" controls></audio>",
				stdhtml.EscapeString(src),
			)
		default:
			return fmt.Sprintf(
				"<img src=\"%s\" alt=\"%s\">",
				stdhtml.EscapeString(src), stdhtml.EscapeString(alt),
			)
		}
	})
}