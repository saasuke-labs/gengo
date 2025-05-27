package generator

import (
	"fmt"
	"os"
	"path/filepath"
)

type Task interface {
	Execute() error
	Name() string
}

func getFullPath(baseDir, relativePath string) string {

	if relativePath == "" {
		return ""
	}

	return filepath.Join(baseDir, relativePath)
}

func scheduleTasks(manifest ManifestFile, baseDir, outDir string) []Task {
	tasks := make([]Task, 0)

	// Copy static files
	for _, asset := range manifest.StaticAssets {
		assetPath := getFullPath(baseDir, asset.Path)
		outPath := getFullPath(outDir, asset.Destination)
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
			Template:       getFullPath(baseDir, manifest.HomeTemplate),
			LayoutTemplate: getFullPath(baseDir, manifest.DefaultLayoutTemplate),
			Metadata:       manifest.Metadata,
		})
	}

	for sectionName, section := range manifest.Sections {
		tags := make(map[string][]Page)

		sectionBasePath := getFullPath(outDir, sectionName)

		// Do not generate the section page if there is no template configured
		if manifest.DefaultSectionTemplate != "" {
			os.MkdirAll(sectionBasePath, 0755)

			outFile := getFullPath(sectionBasePath, "index.html")

			tasks = append(tasks, &SectionTask{
				Title:          manifest.Title,
				Section:        sectionName,
				Sections:       sections,
				OutputFile:     outFile,
				Template:       getFullPath(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: getFullPath(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          section.Pages,
				Metadata:       merge(manifest.Metadata, section.Metadata),
			})
		}

		for _, page := range section.Pages {
			outputFilename := convertExtension(page.MarkdownPath, ".html")
			outPath := getFullPath(sectionBasePath, outputFilename)

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
				InputFile:         getFullPath(baseDir, page.MarkdownPath),
				OutputFile:        outPath,
				Url:               filepath.Join("/", sectionName, outputFilename),
				Template:          getFullPath(baseDir, manifest.DefaultPageTemplate),
				LayoutTemplate:    getFullPath(baseDir, manifest.DefaultLayoutTemplate),
				Metadata:          merge(manifest.Metadata, page.Metadata),
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
				Template:       getFullPath(baseDir, manifest.DefaultSectionTemplate),
				LayoutTemplate: getFullPath(baseDir, manifest.DefaultLayoutTemplate),
				Pages:          pages,
				Section:        sectionName,
				Sections:       sections,
				Metadata:       merge(manifest.Metadata, section.Metadata),
			})

		}
	}

	return tasks
}
