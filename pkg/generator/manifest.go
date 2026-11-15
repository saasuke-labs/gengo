package generator

import (
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type DataRef struct {
	Source string `yaml:"source"`
}

type Page struct {
	Title        string            `yaml:"title"`
	Description  string            `yaml:"description"`
	MarkdownPath string            `yaml:"markdown-path"`
	Template     string            `yaml:"layout"`
	PublishedAt  string            `yaml:"published-at"`
	LastEditedAt string            `yaml:"last-edited-at"`
	Tags         []string          `yaml:"tags"`
	Flags        []string          `yaml:"flags"`
	Metadata     map[string]string `yaml:"metadata"`
	Section      string
	Data         map[string]DataRef `yaml:"data"`
}

func (p *Page) OutFileName() string {
	return convertExtension(p.MarkdownPath, ".html")
}

type Section struct {
	Template     string            `yaml:"template"`
	PageTemplate string            `yaml:"page-template"`
	Pages        []Page            `yaml:"pages"`
	Metadata     map[string]string `yaml:"metadata"`
}

type StaticAsset struct {
	Path        string `yaml:"path"`
	Destination string `yaml:"destination"`
}

type Source struct {
	Url     string            `yaml:"url"`
	Headers map[string]string `yaml:"headers"`
}

type ThemeVariant struct {
	Templates    map[string]string `yaml:"templates"`
	StaticAssets []StaticAsset     `yaml:"static-assets"`
}

type Theme struct {
	Default ThemeVariant `yaml:"default"`
}

type FileGlobs []string

func (f *FileGlobs) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		var single string
		if err := value.Decode(&single); err != nil {
			return err
		}
		*f = []string{single}
		return nil
	case yaml.SequenceNode:
		var many []string
		if err := value.Decode(&many); err != nil {
			return err
		}
		*f = many
		return nil
	default:
		return fmt.Errorf("files must be a string or list of strings")
	}
}

type ContentSection struct {
	Files FileGlobs `yaml:"files"`
}

type ManifestFile struct {
	Title                  string                    `yaml:"title"`
	DefaultLayoutTemplate  string                    `yaml:"default-layout-template"`
	DefaultPageTemplate    string                    `yaml:"default-page-template"`
	DefaultSectionTemplate string                    `yaml:"default-section-template"`
	Metadata               map[string]string         `yaml:"metadata"`
	HomeTemplate           string                    `yaml:"home-template"`
	Sections               map[string]Section        `yaml:"sections"`
	Theme                  Theme                     `yaml:"theme"`
	Content                map[string]ContentSection `yaml:"content"`
	StaticAssets           []StaticAsset             `yaml:"static-assets"`
	Sources                map[string]Source         `yaml:"sources"`
}

type markdownFrontMatter struct {
	Layout      string                 `yaml:"layout"`
	Title       string                 `yaml:"title"`
	PublishedAt string                 `yaml:"published-at"`
	Description string                 `yaml:"description"`
	Tags        []string               `yaml:"tags"`
	Flags       []string               `yaml:"flags"`
	Metadata    map[string]interface{} `yaml:"metadata"`
	Data        map[string]DataRef     `yaml:"data"`
}

func mergeManifest(manifest1, manifest2 ManifestFile) ManifestFile {
	// Merge the two manifests
	merged := manifest1

	if manifest2.Title != "" {
		merged.Title = manifest2.Title
	}

	merged.Metadata = merge(manifest1.Metadata, manifest2.Metadata)

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

	if manifest2.Theme.Default.Templates != nil {
		if merged.Theme.Default.Templates == nil {
			merged.Theme.Default.Templates = make(map[string]string)
		}
		for key, value := range manifest2.Theme.Default.Templates {
			if _, exists := merged.Theme.Default.Templates[key]; !exists {
				merged.Theme.Default.Templates[key] = value
			}
		}
	}

	if manifest2.Theme.Default.StaticAssets != nil {
		if merged.Theme.Default.StaticAssets == nil {
			merged.Theme.Default.StaticAssets = make([]StaticAsset, 0)
		}
		for _, asset := range manifest2.Theme.Default.StaticAssets {
			exists := false
			for _, mergedAsset := range merged.Theme.Default.StaticAssets {
				if mergedAsset.Path == asset.Path && mergedAsset.Destination == asset.Destination {
					exists = true
					break
				}
			}
			if !exists {
				merged.Theme.Default.StaticAssets = append(merged.Theme.Default.StaticAssets, asset)
			}
		}
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

	if manifest2.Content != nil {
		if merged.Content == nil {
			merged.Content = make(map[string]ContentSection)
		}
		for sectionName, section := range manifest2.Content {
			if _, exists := merged.Content[sectionName]; !exists {
				merged.Content[sectionName] = section
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
	mergedManifest = normalizeManifest(mergedManifest)
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

	manifest = normalizeManifest(manifest)
	manifest = populateSectionsFromContent(manifest, filepath.Dir(manifestPath))

	return manifest
}

func normalizeManifest(manifest ManifestFile) ManifestFile {
	templates := manifest.Theme.Default.Templates

	if manifest.DefaultLayoutTemplate == "" {
		manifest.DefaultLayoutTemplate = templates["layout"]
	}
	if manifest.HomeTemplate == "" {
		manifest.HomeTemplate = templates["home"]
	}
	if manifest.DefaultPageTemplate == "" {
		manifest.DefaultPageTemplate = templates["page"]
	}
	if manifest.DefaultSectionTemplate == "" {
		manifest.DefaultSectionTemplate = templates["section"]
	}

	for _, asset := range manifest.Theme.Default.StaticAssets {
		exists := false
		for _, existing := range manifest.StaticAssets {
			if existing.Path == asset.Path && existing.Destination == asset.Destination {
				exists = true
				break
			}
		}
		if !exists {
			manifest.StaticAssets = append(manifest.StaticAssets, asset)
		}
	}

	return manifest
}

func populateSectionsFromContent(manifest ManifestFile, baseDir string) ManifestFile {
	if len(manifest.Content) == 0 {
		return manifest
	}

	if manifest.Sections == nil {
		manifest.Sections = make(map[string]Section)
	}

	for sectionName, contentSection := range manifest.Content {
		section := manifest.Sections[sectionName]

		pages, err := pagesFromGlobs(baseDir, contentSection.Files)
		if err != nil {
			log.Fatalf("failed to load content for section %s: %v", sectionName, err)
		}

		section.Pages = append(section.Pages, pages...)

		if section.Template == "" {
			templates := manifest.Theme.Default.Templates
			candidateKey := sectionName + "List"
			if templates[candidateKey] != "" {
				section.Template = templates[candidateKey]
			} else if templates[sectionName] != "" {
				section.Template = templates[sectionName]
			} else if templates["section"] != "" {
				section.Template = templates["section"]
			}
		}

		manifest.Sections[sectionName] = section
	}

	return manifest
}

func pagesFromGlobs(baseDir string, fileGlobs FileGlobs) ([]Page, error) {
	pages := make([]Page, 0)

	for _, pattern := range fileGlobs {
		globPattern := filepath.Join(baseDir, filepath.FromSlash(pattern))
		matches, err := filepath.Glob(globPattern)
		if err != nil {
			return nil, err
		}

		sort.Strings(matches)

		for _, fullPath := range matches {
			relativePath, err := filepath.Rel(baseDir, fullPath)
			if err != nil {
				return nil, err
			}

			frontMatter, err := readMarkdownFrontMatter(fullPath)
			if err != nil {
				return nil, err
			}

			pages = append(pages, Page{
				Title:        frontMatter.Title,
				Description:  frontMatter.Description,
				MarkdownPath: filepath.ToSlash(relativePath),
				Template:     frontMatter.Layout,
				PublishedAt:  frontMatter.PublishedAt,
				Tags:         frontMatter.Tags,
				Flags:        frontMatter.Flags,
				Metadata:     interfaceMapToStringMap(frontMatter.Metadata),
				Data:         frontMatter.Data,
			})
		}
	}

	return pages, nil
}

func readMarkdownFrontMatter(markdownPath string) (markdownFrontMatter, error) {
	content, err := os.ReadFile(markdownPath)
	if err != nil {
		return markdownFrontMatter{}, err
	}

	frontMatterData, hasFrontMatter := extractYAMLFrontMatter(content)
	if !hasFrontMatter {
		return markdownFrontMatter{}, nil
	}

	var frontMatter markdownFrontMatter
	if err := yaml.Unmarshal(frontMatterData, &frontMatter); err != nil {
		return markdownFrontMatter{}, fmt.Errorf("invalid front matter in %s: %w", markdownPath, err)
	}

	return frontMatter, nil
}

func extractYAMLFrontMatter(content []byte) ([]byte, bool) {
	lines := bytes.Split(content, []byte("\n"))
	if len(lines) == 0 {
		return nil, false
	}

	if strings.TrimSpace(strings.TrimSuffix(string(lines[0]), "\r")) != "---" {
		return nil, false
	}

	for idx := 1; idx < len(lines); idx++ {
		if strings.TrimSpace(strings.TrimSuffix(string(lines[idx]), "\r")) == "---" {
			return bytes.Join(lines[1:idx], []byte("\n")), true
		}
	}

	return nil, false
}

func interfaceMapToStringMap(in map[string]interface{}) map[string]string {
	if len(in) == 0 {
		return nil
	}

	out := make(map[string]string)
	for key, value := range in {
		out[key] = fmt.Sprint(value)
	}

	return out
}
