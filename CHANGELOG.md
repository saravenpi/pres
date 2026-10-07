# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.2] — 2026-10-07

### Fixed
- GFM tables were not rendered (goldmark needs the `extension.GFM` extension explicitly). Tables now render as proper HTML `<table>` elements.

## [0.1.1] — 2026-10-07

### Fixed
- Mermaid diagrams fail to load when opening `index.html` from `file://` due to CORS blocking ES module imports. Switched from ESM (`mermaid.esm.min.mjs`) to UMD build (`mermaid.min.js`) with a plain `<script src>` tag.

## [0.1.0] — 2026-10-07

### Added
- Initial release: markdown-to-slides CLI tool
- `serve` command with live reload via SSE and file watching
- `build` command for self-contained static HTML output
- `new` command to scaffold a presentation folder
- Mermaid diagram support from `.mmd` files and inline fenced blocks
- Asset inlining (images, video, audio) as base64 data URIs
- Undefined Medium font embedded at compile time via `//go:embed`
- Keyboard navigation (Space/arrows) with slide counter
- Dark minimal theme

[0.1.2]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.2
[0.1.1]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.1
[0.1.0]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.0