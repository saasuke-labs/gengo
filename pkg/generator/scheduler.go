package generator

import (
	"fmt"
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

	// Copy static files
	for _, asset := range manifest.StaticAssets {
		assetPath := path.Join(baseDir, asset.Path)
		outPath := path.Join(outDir, asset.Destination)
		tasks = append(tasks, &CopyTask{
			FromPath: assetPath,
			ToPath:   outPath,
		})
	}

	sections := make([]string, 0)
	for section, _ := range manifest.Sections {
		sections = append(sections, section)
	}

	if manifest.HomeTemplate != "" {
		homePath := filepath.Join(outDir, "index.html")

		tasks = append(tasks, &HomeTask{
			Title:          manifest.Title,
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
				Title:          manifest.Title,
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

			externalDataTasks := []ExternalDataTask{}

			for key, value := range page.ExternalData {

				fmt.Println("\t\tKey: ", key, " Source: ", value.Source, " Url: ", manifest.ExternalData[value.Source].Url)

				externalDataTasks = append(externalDataTasks, ExternalDataTask{
					Key: key,
					Url: manifest.ExternalData[value.Source].Url,
				})
			}

			tasks = append(tasks, PageTask{
				Title:             manifest.Title,
				InputFile:         path.Join(baseDir, page.MarkdownPath),
				OutputFile:        outPath,
				Url:               filepath.Join("/", sectionName, outputFilename),
				Template:          path.Join(baseDir, manifest.DefaultPageTemplate),
				LayoutTemplate:    path.Join(baseDir, manifest.DefaultLayoutTemplate),
				Metadata:          page.Metadata,
				Tags:              page.Tags,
				Section:           sectionName,
				Sections:          sections,
				ExternalDataTasks: externalDataTasks,
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
				Title:          manifest.Title,
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
