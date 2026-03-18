# Nagare Extension for Gengo

This extension adds support for `nagare` code blocks in Markdown files. When a fenced code block with the language `nagare` is encountered, the extension will:

1. Extract the Nagare source from the code block
2. Call Nagare's public `pkg/nagare.RenderToSVG` entry point so both diagrams and charts are rendered through the upstream renderer
3. Show an inline error message plus the original source block when Nagare returns an error or only produces an empty/background-only SVG
4. Embed successful SVG output directly into the HTML

## Usage

In your markdown files, use fenced code blocks with `nagare` as the language:

````markdown
# Diagram Example

```nagare
@layout(w:500,h:300)
client:Browser(url: "https://example.com", text: "Web App", x:50,y:100,w:180,h:120)
server:Server(title: "API Server", icon: "server", port: 8080, x:300,y:100,w:150,h:50)
client.e --> server.w
```

# Chart Example

```nagare
chart
title: Test Chart
xaxis: number

series: test
color: #ff0000
data:
  0: 10
  1: 20
  2: 15
```
````

## Error Handling

If Nagare cannot render a block cleanly, Gengo renders:

1. A visible `nagare-error` message containing the upstream error text when available
2. The original `nagare` source block so the author can inspect and fix it

That also covers cases where Nagare returns an almost-empty SVG containing only the background rectangle.

## Implementation Details

The extension works by:

1. Using a Goldmark AST transformer to detect fenced code blocks with language `nagare`
2. Converting them to custom `NagareCodeBlock` AST nodes
3. Calling Nagare's public renderer entry point so chart-vs-diagram detection stays aligned with upstream behavior
4. Rejecting empty/background-only SVG responses before embedding them inline

This keeps Gengo's integration small while improving chart support and surfacing actionable rendering errors.
