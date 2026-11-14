# Gengo

<p align="center">
  <img src="docs-source/static/logo.png" alt="Logo" width="200"/>
</p>
A fast, pluggable static site generator written in Go. Designed to render content from a manifest file and display progress interactively (via terminal UI) or plainly for CI environments.

---

## Installation

To install Gengo you can download the tar files from github releases:

https://github.com/saasuke-labs/gengo/releases

download the target version, uncompress the file and add it to your path.
In case that you found the package somewhere else, do not forget to double check the checksums
that are part of each release.

Or you can use the script to install the latest version:

```sh
curl -fsSL https://raw.githubusercontent.com/saasuke-labs/gengo/main/install/install.sh | bash
```

To install a specific version in one line:

```sh
curl -fsSL https://raw.githubusercontent.com/saasuke-labs/gengo/main/install/install.sh | VERSION=0.0.10 bash
```

## 🚀 Features

- 🧩 Section-based manifest to organize pages (e.g., `blog`, `projects`)
- 🧵 Parallel rendering using Go routines
- 🖥️ Terminal UI with [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Charm](https://charm.sh/)
- 📊 Real-time progress bars and status messages
- 💻 `--plain` mode for CI-friendly output (e.g., GitHub Actions)
- 🧭 Nagare Markdown rendering for both diagrams and charts, with inline error output when a Nagare block fails to render cleanly
- 🔐 Non-commercial license with optional commercial licensing

---

## 📦 Installation

Download a release binary from [Releases](https://github.com/saasuke-labs/gengo/releases) or build from source:

```bash
git clone https://github.com/saasuke-labs/gengo
cd gengo
go build -o gengo ./cmd/main.go
```
---

## Usage

```
./gengo generate --manifest gengo.yaml
```

### Optional Flags

| Flag       | Description                       |
| ---------- | --------------------------------- |
| `--plain`  | Disable interactive TUI rendering |
| `--output` | Specify output directory          |


---

## External data sources

Pages can pull data from any HTTP service at build time (a comments service
such as [Kotomi](https://github.com/saasuke-labs/kotomi), a weather API, your
own backend). Declare the services once under `sources`, then reference them
from a page's `data` block, either in the manifest or in the markdown front
matter:

```yaml
# gengo.yaml
sources:
  comments:
    # Go template, rendered per page with that page's metadata
    url: https://comments.example.com/pages/{{ index .Metadata "page-id" }}/comments
    headers:
      Accept: application/json
```

```markdown
---
title: Hello
metadata:
  page-id: hello-world
data:
  comments:
    source: comments
---
This post has {{ len .Data.comments }} comments.
```

Each entry is available as `.Data.<key>` in the markdown file, the page
template and the layout template. JSON responses are parsed into structured
data, `text/html` is inserted as HTML, anything else is a plain string.
Requests time out after 15 seconds; a failed source logs a warning and leaves
its key empty instead of failing the build.

## Contributing

Contributions are not accepted yet. Once the project stabilizes I will consider it. 
Thanks for your understanding.
