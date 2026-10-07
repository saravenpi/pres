# Changelog

All notable changes to this project are documented here. The format is
[Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.5] — 2026-10-07

### Fixed
- `catalyst build .` without `-o` flag wrote output to a file named `dist` instead of deriving the filename from the presentation directory. Default output flag is now empty; the builder falls back to `<dirname>.html`.

## [0.1.4] — 2026-10-07

### Changed
- `build` now outputs a single self-contained HTML file instead of a folder. Mermaid.js is inlined — no separate `.js` file, no CDN dependency, works fully offline.
- Font switched to Helvetica Neue / Helvetica / Arial system stack. No more base64 font bloat.
- Mermaid diagrams now use a transparent background with Catalyst accent colors, blending into slides.

### Fixed
- Raw `<video>` and `<audio>` HTML tags are now base64-inlined.
- Tables now have proper border styling.

## [0.1.3] — 2026-10-07

### Fixed
- Video and audio files embedded via raw HTML tags were not base64-inlined.

## [0.1.2] — 2026-10-07

### Fixed
- GFM tables were not rendered.

## [0.1.1] — 2026-10-07

### Fixed
- Mermaid diagrams fail to load from `file://` due to CORS.

## [0.1.0] — 2026-10-07

### Added
- Initial release: markdown-to-slides CLI tool.

[0.1.5]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.5
[0.1.4]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.4
[0.1.3]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.3
[0.1.2]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.2
[0.1.1]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.1
[0.1.0]: https://github.com/saravenpi/catalyst/releases/tag/v0.1.0