# Flags Feature - Usage Example

This directory contains an example of how to use the flags feature in gengo.

## Overview

The flags feature allows you to tag pages with custom flags (e.g., "pinned", "archived", "featured") and filter them in your templates using the `where` function.

## Configuration

In your `gengo.yaml`, add flags to your pages:

```yaml
sections:
  blog:
    pages:
      - title: "Pinned Post"
        markdown-path: pinned.md
        flags:
          - pinned
      - title: "Archived Post"
        markdown-path: archived.md
        flags:
          - archived
      - title: "Featured Post"
        markdown-path: featured.md
        flags:
          - pinned
          - featured
```

## Using the `where` Function in Templates

The `where` function is available in section templates and filters pages based on flag conditions.

### Syntax

```
{{ range .Pages | where "flag1,flag2,!flag3" }}
  <!-- Template code -->
{{ end }}
```

- Multiple conditions are separated by commas
- Use `!` prefix to negate a condition (exclude pages with that flag)
- All conditions must be satisfied (AND logic)

### Examples

**Show only pinned pages:**
```html
{{ range .Pages | where "pinned" }}
<div>{{ .Title }}</div>
{{ end }}
```

**Show pinned pages that are NOT archived:**
```html
{{ range .Pages | where "pinned,!archived" }}
<div>{{ .Title }}</div>
{{ end }}
```

**Show pages without any special flags:**
```html
{{ range .Pages | where "!pinned,!archived" }}
<div>{{ .Title }}</div>
{{ end }}
```

**Show pages with multiple flags:**
```html
{{ range .Pages | where "pinned,featured" }}
<div>{{ .Title }}</div>
{{ end }}
```

## Running the Example

To generate this example:

```bash
cd test-resources/flags-test/input
gengo generate --manifest gengo.yaml --output ../output --plain
```

Check the generated `output/blog/index.html` to see the filtered results.

## Output

The example section template demonstrates four different filters:
1. **All Pages** - Shows all 4 pages
2. **Pinned Pages (not archived)** - Shows "Pinned Post" and "Pinned and Featured Post"
3. **Not Pinned, Not Archived** - Shows "Normal Post"
4. **Featured Pages** - Shows "Pinned and Featured Post"
