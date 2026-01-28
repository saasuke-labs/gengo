# Nagare Extension for Gengo

This extension adds support for `nagare` code blocks in Markdown files. When a fenced code block with the language `nagare` is encountered, the extension will:

1. Extract the nagare code from the code block
2. Use the nagare library directly to render it as an SVG diagram
3. Embed the resulting SVG into the HTML output

## Usage

In your markdown files, use fenced code blocks with `nagare` as the language:

````markdown
# My Document

Here's some nagare code:

```nagare
circle(50, 50, 30)
fill("red")
text("Hello World", 10, 100)
```

Regular markdown continues here...
````

## How it works

The extension uses the [nagare](https://github.com/saasuke-labs/nagare) Go library to convert nagare code into SVG diagrams. This happens during the markdown processing phase, with no external service dependencies.

## Error Handling

If there's an error rendering the nagare code, the extension will:

1. Display an error message with details
2. Fall back to rendering the original nagare code as a regular code block

Example fallback output:

```html
<div class="nagare-error">
  <p><strong>Error processing nagare block:</strong> syntax error on line 2</p>
</div>
<pre><code class="language-nagare">
circle(50, 50, 30)
fill("red")
text("Hello World", 10, 100)
</code></pre>
```

## Implementation Details

The extension works by:

1. Using a Goldmark AST transformer to detect fenced code blocks with language "nagare"
2. Converting them to custom `NagareCodeBlock` AST nodes
3. Using a custom renderer that calls the nagare library directly
4. Rendering the resulting SVG inline in the HTML

This approach ensures compatibility with other Goldmark extensions and maintains the performance benefits of AST-based processing.
