package generator

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestScheduleTasks_PageTemplatePrecedence(t *testing.T) {
	manifest := ManifestFile{
		Title:                 "Site",
		DefaultLayoutTemplate: "layout.html",
		DefaultPageTemplate:   "default-page.html",
		Sections: map[string]Section{
			"cities": {
				PageTemplate: "section-page.html",
				Pages: []Page{
					{Title: "Barcelona", MarkdownPath: "cities/barcelona.md", Template: "frontmatter-page.html"},
					{Title: "Lisbon", MarkdownPath: "cities/lisbon.md"},
				},
			},
		},
	}

	tasks := scheduleTasks(manifest, "/base", "/out")
	require.Len(t, tasks, 2)

	firstPageTask, ok := tasks[0].(PageTask)
	require.True(t, ok)
	assert.Equal(t, "/base/frontmatter-page.html", firstPageTask.Template)
	assert.Equal(t, "Barcelona", firstPageTask.Title)

	secondPageTask, ok := tasks[1].(PageTask)
	require.True(t, ok)
	assert.Equal(t, "/base/section-page.html", secondPageTask.Template)
	assert.Equal(t, "Lisbon", secondPageTask.Title)
}
