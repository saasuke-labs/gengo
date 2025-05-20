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

func mergeManifest(manifest1, manifest2 ManifestFile) ManifestFile {
	// Merge the two manifests
	merged := manifest1

	if manifest2.Title != "" {
		merged.Title = manifest2.Title
	}

	if manifest2.DefaultLayoutTemplate != "" {
		merged.DefaultLayoutTemplate = manifest2.DefaultLayoutTemplate
	}

	if manifest2.DefaultPageTemplate != "" {
		merged.DefaultPageTemplate = manifest2.DefaultPageTemplate
	}
	if manifest2.DefaultSectionTemplate != "" {
		merged.DefaultSectionTemplate = manifest2.DefaultSectionTemplate
	}
	if manifest2.HomeTemplate != "" {
		merged.HomeTemplate = manifest2.HomeTemplate
	}

	if manifest2.Sections != nil {
		if merged.Sections == nil {
			merged.Sections = make(map[string]Section)
		}
		for sectionName, section := range manifest2.Sections {
			if _, exists := merged.Sections[sectionName]; !exists {
				merged.Sections[sectionName] = section
			}
		}
	}

	if manifest2.StaticAssets != nil {
		if merged.StaticAssets == nil {
			merged.StaticAssets = make([]StaticAsset, 0)
		}
		for _, asset := range manifest2.StaticAssets {
			// Check if the asset already exists in the merged list
			exists := false
			for _, mergedAsset := range merged.StaticAssets {
				if mergedAsset.Path == asset.Path && mergedAsset.Destination == asset.Destination {
					exists = true
					break
				}
			}
			if !exists {
				merged.StaticAssets = append(merged.StaticAssets, asset)
			}
		}
	}

	return merged
}

func getManifest(manifestPaths []string) ManifestFile {
	// Read the manifest files and merge them
	var mergedManifest ManifestFile = getManifestFile(manifestPaths[0])
	for _, manifestPath := range manifestPaths[1:] {
		manifest := getManifestFile(manifestPath)
		mergedManifest = mergeManifest(mergedManifest, manifest)
	}
	return mergedManifest
}

func getManifestFile(manifestPath string) ManifestFile {
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
