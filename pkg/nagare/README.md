# Nagare Extension for Gengo

This extension adds support for `nagare` code blocks in Markdown files. When a fenced code block with the language `nagare` is encountered, the extension will:

1. Extract the nagare code from the code block
2. Send it via HTTP POST to a configured nagare service
3. Replace the code block with the HTML response from the service

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

## Configuration

### Service URL

The nagare service URL can be configured via the `NAGARE_SERVICE_URL` environment variable:

```bash
export NAGARE_SERVICE_URL=http://localhost:8080/render
```

If not set, it defaults to `http://localhost:8080/render`.

### Disabling Nagare Extension

If you want to disable the nagare extension entirely (useful when the nagare service is not available), set:

```bash
export NAGARE_DISABLE=true
```

When disabled, nagare code blocks will be rendered as regular fenced code blocks.

### Timeout

The extension has a built-in 5-second timeout for HTTP requests to prevent hanging when the service is unavailable.

## Nagare Service API

Your nagare service should accept:

- **Method**: POST
- **Content-Type**: text/plain
- **Body**: The raw nagare code from the code block
- **Response**: HTML content (typically a `<canvas>` element with rendering instructions)

### Example Service Response

```html
<canvas id="nagare-canvas" width="400" height="300">
  <p>Fallback content for browsers that don't support canvas</p>
</canvas>
<script>
  const canvas = document.getElementById("nagare-canvas")
  const ctx = canvas.getContext("2d")
  // ... rendering code based on the nagare input
</script>
```

## Error Handling

If the nagare service is unavailable or returns an error, the extension will:

1. Display an error message
2. Fall back to rendering the original nagare code as a regular code block

Example fallback output:

````html
## Error Handling If the nagare service is unavailable or returns an error, the
extension will: 1. Display a detailed error message including any error text
returned by the service 2. Fall back to rendering the original nagare code as a
regular code block ### Connection Errors If the service is unreachable: ```html
<div class="nagare-error">
  <p>
    <strong>Error processing nagare block:</strong> failed to call nagare
    service: connection refused
  </p>
</div>
<pre><code class="language-nagare">
circle(50, 50, 30)
fill("red")
text("Hello World", 10, 100)
</code></pre>
````

### Service Errors (400, 500, etc.)

If the service returns an error status with a message:

```html
<div class="nagare-error">
  <p>
    <strong>Error processing nagare block:</strong> nagare service returned
    status 400: Invalid nagare syntax: expected 'circle' but got 'triangle'
  </p>
</div>
<pre><code class="language-nagare">
triangle(50, 50, 30)
fill("red")
text("Hello World", 10, 100)
</code></pre>
```

```

## Implementation Details

The extension works by:

1. Using a Goldmark AST transformer to detect fenced code blocks with language "nagare"
2. Converting them to custom `NagareCodeBlock` AST nodes
3. Using a custom renderer that makes HTTP requests to the nagare service
4. Rendering the service response as raw HTML

This approach ensures compatibility with other Goldmark extensions and maintains the performance benefits of AST-based processing.
```
