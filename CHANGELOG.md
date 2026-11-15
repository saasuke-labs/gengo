# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- External data sources: declare HTTP services under `sources` and reference them from a page's `data` block; the response is available as `.Data` in markdown, page and layout templates
- Markdown files can use Go template syntax when the page has data
- `content` sections that discover pages from file globs and read title, layout, tags, flags, metadata and data from YAML front matter
- Per-page and per-section template overrides
- 15 second timeout on source requests

### Changed
- The layout `.Title` is now the page title, falling back to the site title
- `gengo` exits with a non-zero status when a command fails
- Updated nagare to v0.0.14: redesigned diagram components, inline props rendered, chart axes on round values aligned with the data

### Removed
- The unreleased `external-data` manifest key, replaced by `sources`

## [0.0.8] - 2026-01-17

### Added
- Initial tagged release of Gengo
- Section-based manifest to organize pages
- Parallel rendering using Go routines
- Terminal UI with Bubble Tea & Charm
- Real-time progress bars and status messages
- Plain mode for CI-friendly output
- Non-commercial license with optional commercial licensing

[0.0.8]: https://github.com/saasuke-labs/gengo/releases/tag/v0.0.8
