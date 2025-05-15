package generator

import (
	"os"
	"path"
	"path/filepath"
)

type Task interface {
	Execute() error
	Name() string
}

func scheduleTasks(manifest ManifestFile, baseDir, outDir string) []Task {
	tasks := make([]Task, 0)

	sections := make([]string, 0)
	for section, _ := range manifest.Sections {
		sections = append(sections, section)
	}

	if manifest.HomeTemplate != "" {
		homePath := filepath.Join(outDir, "index.html")

		tasks = append(tasks, &HomeTask{
			Sections:       sections,
			OutputFile:     homePath,
			Template:       path.Join(baseDir, manifest.HomeTemplate),
			LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
		})
	}

	for sectionName, section := range manifest.Sections {
		tags := make(map[string][]Page)

		sectionBasePath := filepath.Join(outDir, sectionName)

		// Do not generate the section page if there is no template configured
		if manifest.DefaultSectionTemplate != "" {
			os.MkdirAll(sectionBasePath, 0755)

			outFile := filepath.Join(sectionBasePath, "index.html")

			tasks = append(tasks, &SectionTask{
				Section:        sectionName,
				Sections:       sections,
				OutputFile:     outFile,
				Template:       path.Join(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          section.Pages,
			})
		}

		for _, page := range section.Pages {
			outputFilename := convertExtension(page.MarkdownPath, ".html")
			outPath := filepath.Join(sectionBasePath, outputFilename)
			tasks = append(tasks, PageTask{
				InputFile:      path.Join(baseDir, page.MarkdownPath),
				OutputFile:     outPath,
				Url:            filepath.Join("/", sectionName, outputFilename),
				Template:       path.Join(baseDir, manifest.DefaultPageTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Metadata:       page.Metadata,
				Tags:           page.Tags,
				Section:        sectionName,
				Sections:       sections,
			})

			for _, tag := range page.Tags {
				if _, ok := tags[tag]; !ok {
					tags[tag] = make([]Page, 0)
				}
				tags[tag] = append(tags[tag], page)
			}
		}

		tagsBasePath := filepath.Join(sectionBasePath, "tags")

		for tag, pages := range tags {
			tagOutputFile := filepath.Join(tagsBasePath, slugify(tag)+".html")
			os.MkdirAll(tagsBasePath, 0755)

			// TODO - Create specific task for tags
			tasks = append(tasks, &SectionTask{
				OutputFile:     tagOutputFile,
				Template:       path.Join(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          pages,
				Section:        sectionName,
				Sections:       sections,
			})

		}
	}

	return tasks
}
