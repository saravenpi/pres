# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.4] — 2026-10-07

### Changed
- `build` now outputs a single self-contained HTML file instead of a folder. Mermaid.js is inlined as a `<script>` block — no separate `.js` file, no CDN dependency, works fully offline.
- Font switched from embedded Undefined Medium to Helvetica Neue / Helvetica / Arial system stack. No more base64 font bloat.
- Mermaid diagrams now use a transparent background (`theme: 'base'`) with Catalyst accent colors (`#7c5cff`). Diagrams blend seamlessly into slides.

### Fixed
- Raw `<video>` and `<audio>` HTML tags are now base64-inlined (previously only `![alt](file)` markdown syntax was handled).
- Tables now have proper border styling (accent-colored headers, row dividers).

## [0.1.3] — 2026-10-07

### Fixed
- Video and audio files embedded via raw HTML tags were not base64-inlined.

## [0.1.2] — 2026-10-07

### Fixed
- GFM tables were not rendered (goldmark needs the `extension.GFM` extension).

## [0.1.1] — 2026-10-07

### Fixed
- Mermaid diagrams fail to load from `file://` due to CORS blocking ES module imports. Switched from ESM to UMD build.

## [0.1.0] — 2026-10-07

### Added
- Initial release: markdown-to-slides CLI tool with serve, build, new commands.

[0.1.4]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.4
[0.1.3]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.3
[0.1.2]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.2
[0.1.1]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.1
[0.1.0]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.0