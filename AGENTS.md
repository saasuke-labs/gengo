# Agent notes for `gengo`

Gengo is a static site generator written in Go. It reads one or more YAML manifests
(`gengo.yaml`), renders Markdown pages through Go `html/template` files, copies static
assets and writes a static site. It ships as a single CLI binary (`gengo generate`,
`gengo serve`, `gengo version`) built from `cmd/main.go`. Module path:
`github.com/saasuke-labs/gengo`, Go 1.24.

## Ecosystem (Saasuke Labs)

- **nagare** (`github.com/saasuke-labs/nagare`, pinned in `go.mod`, currently `v0.0.14`):
  text-to-SVG diagram and chart renderer. Gengo imports it as a Go library
  (`nagare/pkg/nagare.RenderToSVG`) through the goldmark extension in `pkg/nagare/`.
  A fenced code block whose language is exactly `nagare` is replaced by the inline SVG.
  A nagare release does not reach gengo users until `go.mod` is bumped and gengo is released.
- **kotomi** (comments service): no code dependency. A gengo site consumes it like any
  other HTTP service, through the manifest `sources` feature (see below).
- **saasuke-labs.com**: built from the `saasuke-labs/site-source` repo. The release
  workflow here sends a `repository_dispatch` (`event_type: product-release`) to
  `site-source` so the site picks up the new version; it needs the `SITE_DISPATCH_TOKEN`
  secret and is skipped without it (the site also rebuilds daily).
- **`docs/index.html`** is this project's page on the site (`/gengo/`). It is hand-written
  HTML that loads CSS/JS from the main site (`/static/site.css`, `/static/site.js`), so it
  looks unstyled when opened locally. Keep it in sync when flags or manifest keys change.
- **demo-my-cities**: a sample site that installs gengo with `install/install.sh` and runs
  `gengo generate --manifest gengo.yaml --output ./site`.

## Commands

All verified with Go 1.24.7. The first run downloads modules, after that no network is needed.

```bash
go build -o gengo ./cmd/main.go      # `gengo` is gitignored; always pass -o (see gotchas)
go test ./...                        # what CI runs; all packages pass
go test ./pkg/generator -run TestWherePages -v   # single test / name prefix
go test -short ./...                 # skips TestIntegrationGenerate
gofmt -l .                           # currently clean; `gofmt -w .` to fix
go vet ./...                         # one known finding: unreachable code in pkg/testutils

# run against the bundled docs site (manifest paths are relative to the manifest's dir)
cd docs-source && ../gengo generate --manifest gengo.yaml --output /tmp/out --plain
./gengo serve --site /tmp/out --port 3000 --watch
```

Flags: `generate --manifest <path>` (repeatable, default `gengo.yaml`), `--output`
(default `output`), `--plain`, `--watch`; `serve --site` (default `site`), `--port`
(default 3000), `--watch`. Use `--plain` in CI and from agents: the default mode clears
the screen and redraws a progress view with ANSI codes.

There is no tracked `justfile` (it is gitignored) and no linter config.

## Architecture

Flow: `cmd/main.go` (cobra root, `serve`) → `pkg/cli` (`generate`, `version`, terminal
progress in `render.go`) → `generator.GenerateSiteAsync(manifestPaths, outputDir)` →
`getManifest` → `scheduleTasks` → every task runs in its own goroutine and reports
`FileProgress` (`pending/started/completed/failed`) on a channel that the CLI drains.

| Package | Owns |
| --- | --- |
| `pkg/generator` | Manifest types, loading and merging (`manifest.go`), task scheduling (`scheduler.go`), the tasks (`page-task.go`, `section-task.go`, `home-task.go`, `copy-task.go`), template application (`templates.go`), the `where` filter (`template-functions.go`) |
| `pkg/parser` | Markdown → HTML with goldmark: GFM, chroma highlighting (style `github`), auto heading IDs, hard wraps, XHTML, plus the nagare extension. Strips YAML front matter. Raw HTML in Markdown is dropped (goldmark `WithUnsafe` is not set) |
| `pkg/nagare` | Goldmark AST transformer + renderer for `nagare` blocks. On error, or when the SVG has no drawable elements (`isBareBackgroundSVG`), it emits `<div class="nagare-error">` followed by the original source block instead of failing the build |
| `pkg/server` | `gengo serve`: static file server; with `--watch` it injects a WebSocket reload script before `</body>` and reloads all clients on any change under the site dir |
| `pkg/watcher` | fsnotify wrapper used by serve and `generate --watch` |
| `pkg/telemetry` | PostHog client wrapper. Currently inert: `init()` never creates the client, so `Track`/`Close` are no-ops. `GENGO_TELEMETRY=off` is defined as a constant but not read anywhere |
| `pkg/gsx` | Experimental JSX-like syntax → Go template converter (`<Tag name="Go" />` → `{{ template "Tag" (dict "name" "Go") }}`). Tested, but not wired into the generator |
| `pkg/version` | `Version`/`Commit`/`Date`, set by goreleaser ldflags (`undefined` in local builds) |
| `pkg/testutils` | `CompareHtmlFiles`: DOM comparison that ignores whitespace-only text nodes |

### Manifest (`ManifestFile` in `pkg/generator/manifest.go`)

```yaml
title: My Site
metadata: { key: value }              # merged into every page's .Metadata
default-layout-template: layouts/layout.html   # outer shell, applied to every output
default-page-template: layouts/page.html       # optional wrapper around page HTML
default-section-template: layouts/section.html # section index; no template = no index
home-template: layouts/home.html               # no template = no /index.html
theme: { default: { templates: { layout: ..., home: ..., page: ..., section: ... },
                    static-assets: [...] } }   # fallbacks for the four keys above
static-assets:
  - { path: static, destination: static }      # file or directory copy
sections:
  blog:
    template: ...        # section index template override
    page-template: ...   # page template for this section
    metadata: {}
    pages:
      - title: Post
        markdown-path: posts/post.md   # .md or .html
        layout: ...                    # per-page template override
        description: ""                # also: published-at, last-edited-at
        tags: []                       # also: flags, metadata
        data: { comments: { source: comments } }
content:
  blog: { files: "posts/*.md" }        # string or list of globs; adds pages to section `blog`
sources:
  comments:
    url: https://example.com/pages/{{ index .Metadata "page-id" }}/comments
    headers: { Accept: application/json }
```

- All paths resolve against the directory of the **first** manifest.
- Output paths: `/<section>/<markdown basename>.html`, `/<section>/index.html`,
  `/<section>/tags/<slug>.html` (one per tag, rendered with the section template) and
  `/index.html`. Subdirectories in `markdown-path` are flattened.
- Page template precedence: page `layout` → section `page-template` →
  `default-page-template`. For `content` sections the section template falls back to
  theme templates `<section>List`, then `<section>`, then `section`.
- YAML front matter (`layout`, `title`, `published-at`, `description`, `tags`, `flags`,
  `metadata`, `data`) is only **read** for pages discovered through `content` globs. For
  pages listed under `sections.pages` it is stripped and the manifest values are used.
- Multiple `--manifest`: later files override scalar keys (title, default templates) and
  metadata; for `sections`, `content` and theme templates the first definition of a key
  wins; `sources` are taken from the first manifest only (`mergeManifest` skips them).

### Template data

- Layout and page templates receive `PageData`: `.Title`, `.Tags`, `.HTML`, `.Metadata`,
  `.Section`, `.Sections`, `.Data`. The page template gets an empty `.Title`, `.Section`
  and `.Sections`; only the layout gets them. No custom functions are registered.
- Section templates receive `.Section` and `.Pages` (`[]Page`: `.Title`, `.Description`,
  `.MarkdownPath`, `.PublishedAt`, `.Tags`, `.Flags`, `.Metadata`, `.OutFileName`), plus
  the `where` function: `{{ range .Pages | where "pinned,!archived" }}` (comma = AND,
  `!` = NOT). `.Metadata` is not passed to section templates.
- Home templates receive an empty struct.
- Sources: each `data` entry is fetched per page at build time (15s timeout, in
  parallel). JSON becomes structured data, `text/html` becomes `template.HTML`, anything
  else a string. A failed fetch prints a warning and leaves the key nil. The URL is a
  `text/template` with `.Metadata` and `now`. When a page has data, its Markdown file is
  also pre-rendered as a `text/template` with `.Data`, `.Metadata` and `now`; without
  data, `{{ }}` in Markdown is left as literal text.

## Conventions and gotchas

- **Do not run `go build ./cmd/main.go` without `-o`.** It overwrites the tracked 27 MB
  `main` binary at the repo root. That binary is a stale committed artifact; never
  rebuild or re-commit it.
- Stale files at the repo root, not used by build or tests: `gengo.yaml`, `page.html`,
  `section.html` (they use an old schema with `layout`/`posts`/`file`/`.Content`; running
  `gengo generate` at the root produces 0 files), `test-nagare.md`, `test-simple.md`,
  `test-parser.go` (a `package main` scratch file). `cmd/main.go` also carries a large
  commented-out SSE block. `site/index.html` is an old Tailwind landing page.
- `docs/` is hand-written (the `/gengo/` page). `docs-source/` is a gengo site used as a
  manual smoke test; `docs-source/dist/` is committed generated output, so regenerate it
  only on purpose and write test output elsewhere.
- Golden test: `pkg/cli/generate_test.go` generates `test-resources/simple-blog/input`
  into `test-resources/simple-blog/output` (tracked, deleted and rewritten on every test
  run) and DOM-compares `blog/blog1.html` with `expected-output/blog/blog1.html`. To
  update the golden file, run the tests, review the diff, then copy `output/blog/blog1.html`
  over the expected file. If a rendering change is intended, expect `git status` to show
  `output/` changes after `go test` and commit them with the golden file.
- `test-resources/flags-test/` is a documented example for `where`; no Go test reads it.
  `pkg/test-resources/` and `test-resources/simple-blog/input/output/` are leftovers.
- Error handling is abrupt: missing or invalid manifests and template execution errors
  call `log.Fatalf`; template parse errors panic through `template.Must` inside a task
  goroutine. Only `CopyTask` returns errors, so `failed` status is rare and a failed task
  does not change the exit code.
- `generate --watch` watches a hard-coded `./blog` directory (TODO in
  `pkg/cli/generate.go`). The working loop is: re-run `generate`, keep `serve --watch`
  open (it hard-codes `ws://localhost:<port>` and only watches directories that exist
  at startup).
- `generate` prints the whole parsed manifest (`Generating site... {…}`), and the nagare
  library prints debug lines such as `State layout props: …` to stdout. Do not parse stdout.
- The README is partly out of date: it mentions Bubble Tea (the progress UI is plain
  lipgloss + ANSI) and lists only two flags. Trust `--help` and the code.

## Release and CI

- `validate.yaml`: `go test ./...` on pushes and PRs to `main`.
- `pr-goreleaser-check.yaml`: `goreleaser release --snapshot --skip=publish --clean` on
  PRs, inside the `goreleaser/goreleaser-cross:v1.24` container.
- `release.yaml`: pushing a tag `vX.Y.Z` runs `goreleaser release --clean` (linux and
  darwin amd64/arm64, windows amd64; `prerelease: auto`), then notifies `site-source`.
  To release: merge to `main`, then `git tag vX.Y.Z && git push origin vX.Y.Z`.
- `goreleaser.yaml` builds with `CGO_ENABLED=1` and cross compilers. That dates from when
  nagare pulled in a cgo WebP library; a `CGO_ENABLED=0` build of the current tree works.
- `RELEASE.md` is a one-off note about re-tagging v0.0.9, not a general process doc.
  `CHANGELOG.md` only has `Unreleased` and `0.0.8`; the public release notes live on
  GitHub Releases.
- `install/install.sh` downloads `gengo_<version>_<os>_<arch>.tar.gz` from GitHub
  Releases into `~/.local/bin` (needs `curl` and `jq`); keep archive naming compatible.
- Licence: Polyform Noncommercial 1.0.0 (`LICENSE`); commercial terms in `COMMERCIAL.md`.
  External contributions are not accepted yet (README).

## Deeper docs

- `docs/index.html`: the public user documentation (manifest, templates, data, CI).
- `README.md`: install and the `sources` walkthrough. `pkg/nagare/README.md`: nagare block
  examples (diagram and chart). `test-resources/flags-test/README.md`: `flags` and `where`.
