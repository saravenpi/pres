## Goal

A single Go binary (`catalyst`) that turns a folder of markdown files, assets, and Mermaid
diagrams into a self-contained HTML slideshow, serves it locally, and compiles it to a static
`dist/` folder.

## Why (evidence)

The user wants to write presentations in plain markdown with inline assets and `.mmd` diagrams,
not in a proprietary format or a web-based editor. The demo loop is "edit, refresh" — not
"export to PowerPoint, fix slide 14, re-export."

## Approach

One Go binary with no runtime dependencies. It embeds the font and default CSS at compile time,
parses the markdown folder into a `Presentation` struct, renders a single self-contained HTML
page (all assets base64-inlined, no network fetches), and serves it with a tiny webserver. The
same render path produces the static `dist/` output.

Four deep modules behind a thin CLI:

- **Parser** — `Parse(dir string) (*Presentation, error)`. Walks the folder, reads the root
  `.md`, splits slides, resolves local asset references and `.mmd` files.
- **Renderer** — `Render(p *Presentation) ([]byte, error)`. Produces a single HTML document
  with everything inlined (CSS, JS from CDN, font as base64 woff2, all assets).
- **Server** — `Serve(p *Presentation, port int) error`. Renders once, serves with a
  file-watch loop that re-parses on change and pushes a refresh signal.
- **Builder** — `Build(p *Presentation, outDir string) error`. Renders to `outDir/index.html`
  and copies referenced assets alongside.

The Renderer is the deep module — a lot of behaviour (CSS layout, slide transitions, keyboard
navigation, Mermaid integration, font embedding) behind a small interface (one method).

## Files to create (ordered by dependency)

1. `go.mod` + `go.sum` — module `github.com/saravenpi/catalyst`
2. `embed/fonts/undefined-medium.woff2` — downloaded from
   https://github.com/andirueckel/undefined-medium/raw/v1.3/fonts/webfonts/undefined-medium.woff2
3. `embed/default.css` — slide layout CSS (full-viewport slides, centered content,
   Undefined Medium font, dark theme)
4. `embed/embed.go` — `//go:embed fonts/* default.css` with an `embed.FS`
5. `internal/parser/parser.go` — the Parser module
6. `internal/renderer/renderer.go` — the Renderer module (deep module — embeds CSS,
   constructs HTML template with Mermaid.js CDN script, keyboard nav JS, all assets
   base64-inlined)
7. `internal/server/server.go` — the Server module (net/http + fsnotify or polling watch,
   SSE for live reload)
8. `internal/builder/builder.go` — the Builder module
9. `cmd/catalyst/main.go` — CLI entry point: `serve`, `build`, `new` subcommands
10. `main.go` — package main, calls `cmd/catalyst`

## Presentation model

A folder containing:

```
pres.md          ← root markdown (slides separated by ---)
cat.png          ← image asset
dance.mp4        ← video asset
profit.mmd       ← Mermaid diagram (referenced inline or embedded)
```

- Slides split on `---` (horizontal rule) in the root markdown.
- `![alt](cat.png)` renders the image inline (base64 in the HTML).
- `.mmd` files are read and rendered as Mermaid diagrams by the Mermaid.js runtime.
- `<video>`, `<audio>` pass through natively.
- A `!mermaid[profit.mmd]` syntax (or ` ```mermaid` fenced block with a file ref) embeds
  a `.mmd` file's contents as a Mermaid code block.

Simpler approach: markdown fenced blocks with `mermaid` are already handled by Mermaid.js
in the browser, so `.mmd` files can just be read into fenced blocks during parsing. The
user writes their markdown normally and can reference `.mmd` files with a special syntax
or just paste the mermaid code directly.

## CLI surface

```
catalyst serve <dir>        Parse, render, serve on localhost:3000, watch for changes
catalyst serve <dir> -p 8080  Serve on a specific port
catalyst build <dir>        Compile to dist/<dir>/index.html with all assets
catalyst build <dir> -o out  Compile to out/index.html
catalyst new <name>         Scaffold a new presentation folder
```

Following CLI standard convention: `--version` prints `catalyst 0.1.0`, `--help`
is automatic via cobra, subcommands are single lowercase words.

## Keyboard navigation

Injected JS in the rendered page:
- **Space / ArrowRight / ArrowDown** → next slide
- **ArrowLeft / ArrowUp** → previous slide
- Wrap-around at both ends
- Current slide indicator (dots or fraction)

## Font

`undefined-medium.woff2` embedded as base64 in the rendered HTML. Applied via CSS
`@font-face` to all text elements. No external font dependency at serve time or
in the compiled output.

## Exit criteria

- `go build ./...` compiles clean
- `catalyst serve testdata/example` opens a browser showing slides
- Space/arrow keys navigate slides
- Images, videos, audio render correctly
- Mermaid diagrams render from `.mmd` files and inline fenced blocks
- `catalyst build testdata/example -o /tmp/out` produces a self-contained folder
  that works when opened directly (file://) or served statically
- The font is the Undefined Medium pixel font on all text

## Risks / unknowns

- Mermaid.js is ~3MB minified. Loading from CDN is simpler; bundling inline makes
  the HTML huge. The CDN approach with a `<script>` tag is the pragmatic choice.
- `fsnotify` requires CGO on some platforms. Alternative: poll every 500ms with a
  simple file-watcher, which is simpler and cross-platform.
- File:// protocol may block CDN fetches — `build` mode should work fully offline,
  so either bundle Mermaid.js or note that an internet connection is needed for
  first render.
- For offline `build` output: download Mermaid.js once and embed it. This makes
  `dist/` truly self-contained.

## Skip (YAGNI)

- No theme system — hardcoded dark minimal theme
- No presenter notes
- No PDF export
- No animations or transitions between slides (instant swap)
- No speaker view / dual-screen
- No markdown extensions beyond commonmark + GFM + mermaid
- No config file — CLI flags only
- No plugin system