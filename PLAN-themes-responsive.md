# Plan: responsive Swiss design, light/dark themes, fang CLI

## Goal

Three changes to the `pres` CLI in one pass:

1. Make the slide layout and typography adapt to screen size, so a deck looks good from a small laptop to an ultrawide monitor. Apply Swiss (International Typographic Style) layout practice.
2. Add a `--theme light|dark` flag to `serve` and `build`, default `light`. Light becomes the default rendered theme.
3. Move the CLI onto charmbracelet fang for styled help, version, and error output.

## Why

The renderer hardcodes a dark theme and fixed `rem`/`px` font sizes. On an ultrawide screen the type sits small in a sea of empty space, and on a small screen it overflows. The CLI prints plain cobra help and bare error strings.

## Research

Swiss design (International Typographic Style), applied to the screen:

- Grid system: one shared structure every slide answers to, so deliberate breaks read as intentional.
- Typography: one neutral sans-serif, hierarchy carried by size and weight (not variety), a modular scale (1.25 or 1.333), generous line-height.
- White space: a fixed spacing scale, generous padding, resist filling the margins.
- Reduction: remove ornament, one accent colour, black/white/neutral plus the accent.
- Clarity: unambiguous hierarchy, readable at every size.

Fluid/responsive typography:

- `clamp(min, preferred, max)` gives fluid type with no breakpoints.
- For a full-viewport slide deck, the preferred term should use `vmin` (the smaller of viewport width and height). On an ultrawide screen `vmin` tracks height, so type does not blow up horizontally; on a portrait phone it tracks width.

Sources:

- https://swissthemes.design/insights/swiss-design-for-web-designers
- https://clampgenerator.com/blog/fluid-typescale-modern-css-without-media-queries
- https://dev.to/137foundry/how-to-build-a-fluid-type-scale-with-css-clamp-a-complete-implementation-guide-29oh
- https://github.com/charmbracelet/fang
- https://github.com/charmbracelet/fang/blob/main/fang.go
- https://github.com/charmbracelet/fang/blob/main/theme.go

## Decisions

- fang: use `charm.land/fang/v2` v2.0.1, the current maintained line. It requires Go 1.25, so bump the module and the mise pin from 1.24 to 1.25. Go 1.25.14 is already installed in mise.
- fang surface: keep it minimal. Use `fang.WithVersion`, `fang.WithNotifySignal(os.Interrupt)`, and disable the extra `man` and `completion` subcommands so the CLI stays `serve`/`build`/`new` plus help/version/errors.
- Theme model: one CSS file. `:root` holds light values (the default); `[data-theme="dark"]` overrides. The renderer sets `data-theme` on `<html>`.
- `--theme` flag on `serve` and `build`, default `light`, values `light`/`dark`, validated at the CLI boundary.
- Responsive: fluid type and spacing via `clamp()` with `vmin` preferred terms, plus a capped content measure so lines stay readable on ultrawide.

## Shared contract

Implementers must agree on these exact signatures and tokens:

- `renderer.Render` becomes:

  `func Render(p *parser.Presentation, offline bool, mermaidScript []byte, theme string) ([]byte, error)`

  `theme` is `"light"` or `"dark"`; empty or unknown means `light`.

- `<html lang="en" data-theme="light|dark">`.
- CSS custom properties, defined in `:root` and overridden by `[data-theme="dark"]`:
  `--bg`, `--fg`, `--accent`, `--muted`, `--surface`, `--border`.
- Accent colour: `#7c5cff` (keep the existing identity).
- Suggested token values (agents may refine within these families):
  - light: `--bg #fafafa`, `--fg #141414`, `--muted #5c5c5c`, `--surface #efefef`, `--border rgba(0,0,0,0.12)`.
  - dark: `--bg #141414`, `--fg #e8e8e8`, `--muted #9a9a9a`, `--surface #1f1f1f`, `--border rgba(255,255,255,0.12)`.

## Work breakdown (three parallel agents)

### Agent A — fang CLI and theme plumbing

Files: `main.go`, `cmd/pres/cmd.go`, `go.mod`, `go.sum`, `mise.toml`, `internal/server/server.go`, `internal/builder/builder.go`.

1. Bump Go to 1.25 in `go.mod` and `mise.toml`, add `charm.land/fang/v2@v2.0.1`, run `go mod tidy`.
2. Replace the cobra `Execute` path with `fang.Execute(context.Background(), rootCmd, fang.WithVersion(version), fang.WithNotifySignal(os.Interrupt), fang.WithoutCompletions(), fang.WithoutManpage())`. Remove `SetVersionTemplate`.
3. Add a `--theme` string flag to `serve` and `build`, default `light`; validate the value is `light` or `dark`, returning a styled error otherwise.
4. Add `Theme string` to `server.Options` and `builder.Options`, and pass it through to `renderer.Render(..., opts.Theme)`.

### Agent B — renderer theme awareness

Files: `internal/renderer/renderer.go`, `internal/renderer/renderer_test.go`.

1. Change `Render` to the 4-arg contract above.
2. Emit `data-theme="..."` on `<html>`.
3. Make `mermaidInit` theme-aware: light mermaid colours (dark text on transparent/light, accent borders) versus the existing dark colours.
4. Update all `Render` calls in the test file; add a test asserting `data-theme="dark"` appears when the theme is dark and `data-theme="light"` otherwise.

### Agent C — Swiss responsive CSS

File: `embed/default.css` (full rewrite).

1. `:root` light tokens, `[data-theme="dark"]` overrides (the shared contract).
2. Fluid type scale via `clamp()` with `vmin` preferred terms (h1, h2, body, small, code).
3. Fluid spacing and a capped content measure for ultrawide.
4. Swiss layout: flush-left within a centered column, one accent, generous whitespace, reduction.

## Exit criteria

- `mise exec -- go build ./...` clean.
- `mise exec -- go test ./...` green.
- `filet check` passes (the repo uses filet — `filet.yml` is present).
- `pres build testdata/example -o /tmp/out.html --theme light` produces light-themed output; `--theme dark` produces dark.
- The deck renders legibly across viewport widths (fluid type, capped measure).

## Out of scope

- No presenter notes, PDF, animations, or plugin system (unchanged from the original `PLAN.md`).
- No config file — flag only.
- The fang terminal colour scheme auto-detects the terminal background. The `--theme` flag controls the rendered HTML, not the CLI's own colours.
