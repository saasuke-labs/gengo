package generator

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

type Page struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	MarkdownPath string            `yaml:"markdown-path"`
	PublishedAt  string            `yaml:"published-at"`
	LastEditedAt string            `yaml:"last-edited-at"`
	Tags         []string          `yaml:"tags"`
	Metadata     map[string]string `yaml:"metadata"`
	Section      string
}

func (p *Page) OutFileName() string {
	return convertExtension(p.MarkdownPath, ".html")
}

type Section struct {
	Template     string `yaml:"template"`
	PageTemplate string `yaml:"page-template"`
	Pages        []Page `yaml:"pages"`
}

type StaticAsset struct {
	Path        string `yaml:"path"`
	Destination string `yaml:"destination"`
}

type ManifestFile struct {
	Title                  string             `yaml:"title"`
	DefaultLayoutTemplate  string             `yaml:"default-layout-template"`
	DefaultPageTemplate    string             `yaml:"default-page-template"`
	DefaultSectionTemplate string             `yaml:"default-section-template"`
	HomeTemplate           string             `yaml:"home-template"`
	Sections               map[string]Section `yaml:"sections"`
	StaticAssets           []StaticAsset      `yaml:"static-assets"`
}

func getManifest(manifestPath string) ManifestFile {
	log.Printf("Reading manifest file: %s", manifestPath)
	data, err := os.ReadFile(manifestPath)
	if err != nil {
		log.Fatalf("failed to read manifest file: %v", err)
	}
	var manifest ManifestFile
	if err := yaml.Unmarshal(data, &manifest); err != nil {
		log.Fatalf("failed to parse YAML: %v", err)
	}

	return manifest
}
