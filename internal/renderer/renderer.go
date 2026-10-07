package renderer

import (
	"bytes"
	"encoding/base64"
	"fmt"
	stdhtml "html"
	"regexp"
	"strings"

	"github.com/saravenpi/pres/embed"
	"github.com/saravenpi/pres/internal/parser"
	"github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	goldmarkhtml "github.com/yuin/goldmark/renderer/html"
)

var md = goldmark.New(
	goldmark.WithExtensions(extension.GFM),
	goldmark.WithRendererOptions(
		goldmarkhtml.WithUnsafe(),
		goldmarkhtml.WithHardWraps(),
	),
)

var assetRefRE = regexp.MustCompile(`!\[([^\]]*)\]\(([^)]+)\)`)

func Render(p *parser.Presentation, offline bool, mermaidScript []byte, theme string) ([]byte, error) {
	var buf bytes.Buffer

	if theme != "dark" {
		theme = "light"
	}

	css, err := buildCSS()
	if err != nil {
		return nil, fmt.Errorf("building CSS: %w", err)
	}

	buf.WriteString("<!DOCTYPE html>\n")
	fmt.Fprintf(&buf, "<html lang=\"en\" data-theme=\"%s\">\n<head>\n", theme)
	buf.WriteString("<meta charset=\"UTF-8\">\n")
	buf.WriteString("<meta name=\"viewport\" content=\"width=device-width, initial-scale=1.0\">\n")
	fmt.Fprintf(&buf, "<title>%s</title>\n", stdhtml.EscapeString(p.Title))
	buf.WriteString("<style>\n")
	buf.WriteString(css)
	buf.WriteString("\n</style>\n</head>\n<body>\n")

	writeSlides(&buf, p)
	writeMermaid(&buf, offline, mermaidScript, theme)
	writeNav(&buf, offline)

	buf.WriteString("</body>\n</html>\n")
	return buf.Bytes(), nil
}

func writeSlides(buf *bytes.Buffer, p *parser.Presentation) {
	buf.WriteString("<div id=\"slides\">\n")
	for i, slide := range p.Slides {
		fmt.Fprintf(buf,
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
		md.Convert([]byte(content), &htmlBuf)
		buf.WriteString(htmlBuf.String())
		buf.WriteString("</div>\n</section>\n")
	}
	buf.WriteString("</div>\n")
	buf.WriteString("<div id=\"counter\" class=\"indicator\"></div>\n")
}

func writeNav(buf *bytes.Buffer, offline bool) {
	buf.WriteString(kbNavScript)
	if !offline {
		buf.WriteString(liveReloadScript)
	}
}

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
new EventSource('/__pres_reload')
  .addEventListener('reload',function(){location.reload()});
</script>
`

func buildCSS() (string, error) {
	raw, err := embed.Assets.ReadFile("default.css")
	if err != nil {
		return "", fmt.Errorf("reading default.css: %w", err)
	}
	return string(raw), nil
}

func activeClass(i int) string {
	if i == 0 {
		return " active"
	}
	return ""
}

func inlineAssets(content string, assets map[string]parser.Asset) string {
	content = assetRefRE.ReplaceAllStringFunc(content, func(match string) string {
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
		return emitAssetTag(asset, alt)
	})

	content = tagSrcRE.ReplaceAllStringFunc(content, func(match string) string {
		parts := tagSrcRE.FindStringSubmatch(match)
		if len(parts) < 5 {
			return match
		}
		tag := parts[1]
		before := parts[2]
		name := strings.TrimSpace(parts[3])
		after := parts[4]
		if strings.Contains(name, "://") {
			return match
		}
		asset, ok := assets[name]
		if !ok {
			return match
		}
		return fmt.Sprintf("<%s%s src=\"%s\"%s>",
			tag, before, stdhtml.EscapeString(dataURI(asset)), after)
	})

	return content
}

var tagSrcRE = regexp.MustCompile(`<(video|audio)\b([^>]*)src\s*=\s*"([^"]+)"([^>]*)>`)

func dataURI(asset parser.Asset) string {
	b64 := base64.StdEncoding.EncodeToString(asset.Data)
	return fmt.Sprintf("data:%s;base64,%s", asset.MimeType, b64)
}

func emitAssetTag(asset parser.Asset, alt string) string {
	src := dataURI(asset)
	switch {
	case strings.HasPrefix(asset.MimeType, "video/"):
		return fmt.Sprintf("<video src=\"%s\" controls></video>", stdhtml.EscapeString(src))
	case strings.HasPrefix(asset.MimeType, "audio/"):
		return fmt.Sprintf("<audio src=\"%s\" controls></audio>", stdhtml.EscapeString(src))
	default:
		return fmt.Sprintf("<img src=\"%s\" alt=\"%s\">", stdhtml.EscapeString(src), stdhtml.EscapeString(alt))
	}
}
