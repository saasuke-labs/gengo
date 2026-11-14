package generator

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetManifestFile_ContentGlobsAndThemeTemplates(t *testing.T) {
	baseDir := t.TempDir()

	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "src/layouts"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "src/static"), 0755))
	require.NoError(t, os.MkdirAll(filepath.Join(baseDir, "cities"), 0755))

	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "src/static/app.css"), []byte("body{}"), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "cities/barcelona.md"), []byte(`---
layout: layouts/city.html
title: Discover Barcelona
published-at: 2026-02-10
description: Barcelona
tags: ["Barcelona", "Europe"]
metadata:
  hero-image-url: "barcelona.jpg"
---
# Barcelona
`), 0644))
	require.NoError(t, os.WriteFile(filepath.Join(baseDir, "cities/lisbon.md"), []byte(`# Lisbon`), 0644))

	manifestPath := filepath.Join(baseDir, "gengo.yaml")
	require.NoError(t, os.WriteFile(manifestPath, []byte(`title: My Cities

theme:
  default:
    templates:
      layout: "src/layouts/layout.html"
      home: "src/layouts/home.html"
      citiesList: "src/layouts/citiesList.html"
    static-assets:
      - path: src/static
        destination: static

static-assets:
  - path: src/static
    destination: static

content:
  cities:
    files: cities/*.md
`), 0644))

	manifest := getManifestFile(manifestPath)

	assert.Equal(t, "src/layouts/layout.html", manifest.DefaultLayoutTemplate)
	assert.Equal(t, "src/layouts/home.html", manifest.HomeTemplate)
	require.Len(t, manifest.StaticAssets, 1)

	section, ok := manifest.Sections["cities"]
	require.True(t, ok)
	assert.Equal(t, "src/layouts/citiesList.html", section.Template)
	require.Len(t, section.Pages, 2)

	barcelona := section.Pages[0]
	assert.Equal(t, "Discover Barcelona", barcelona.Title)
	assert.Equal(t, "cities/barcelona.md", barcelona.MarkdownPath)
	assert.Equal(t, "layouts/city.html", barcelona.Template)
	assert.Equal(t, []string{"Barcelona", "Europe"}, barcelona.Tags)
	assert.Equal(t, "barcelona.jpg", barcelona.Metadata["hero-image-url"])

	lisbon := section.Pages[1]
	assert.Equal(t, "cities/lisbon.md", lisbon.MarkdownPath)
}
