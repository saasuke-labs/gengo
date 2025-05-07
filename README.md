# Gengo

<p align="center">
  <img src="docs/static/logo.png" alt="Logo" width="200"/>
</p>
A fast, pluggable static site generator written in Go. Designed to render content from a manifest file and display progress interactively (via terminal UI) or plainly for CI environments.

---

## 🚀 Features

- 🧩 Section-based manifest to organize pages (e.g., `blog`, `projects`)
- 🧵 Parallel rendering using Go routines
- 🖥️ Terminal UI with [Bubble Tea](https://github.com/charmbracelet/bubbletea) & [Charm](https://charm.sh/)
- 📊 Real-time progress bars and status messages
- 💻 `--plain` mode for CI-friendly output (e.g., GitHub Actions)
- 🔐 Non-commercial license with optional commercial licensing

---

## 📦 Installation

Download a release binary from [Releases](https://github.com/tonitienda/gengo/releases) or build from source:

```bash
git clone https://github.com/tonitienda/gengo
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

## Contributing

Let me know if you'd like to tailor this for your actual GitHub username/repo or add advanced examples (e.g., plugins, themes, rendering hooks).

