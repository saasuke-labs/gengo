package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/saasuke-labs/gengo/pkg/generator"

	"github.com/charmbracelet/lipgloss"
	"golang.org/x/term"
)

// Styling definitions
var (
	// Colors
	primaryColor = lipgloss.Color("#7D56F4")
	successColor = lipgloss.Color("#73F59F")
	warningColor = lipgloss.Color("#F2C94C")
	dangerColor  = lipgloss.Color("#F25757")
	infoColor    = lipgloss.Color("#2D9CDB")

	// Status styles
	statusIcons = map[generator.FileStatus]string{
		generator.Pending:   "○",
		generator.Started:   "◔",
		generator.Completed: "●",
		generator.Failed:    "✗",
	}

	statusStyles = map[generator.FileStatus]lipgloss.Style{
		generator.Pending:   lipgloss.NewStyle().Foreground(infoColor),
		generator.Started:   lipgloss.NewStyle().Foreground(warningColor),
		generator.Completed: lipgloss.NewStyle().Foreground(successColor),
		generator.Failed:    lipgloss.NewStyle().Foreground(dangerColor),
	}

	// Container styles
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(primaryColor).
			MarginBottom(1)

	fileStyle = lipgloss.NewStyle().
			PaddingLeft(2)

	progressBarStyle = lipgloss.NewStyle().
				Border(lipgloss.NormalBorder(), false, false, true, false).
				BorderForeground(primaryColor).
				MarginTop(1).
				MarginBottom(1).
				PaddingBottom(1)

	progressFillStyle = lipgloss.NewStyle().
				Background(primaryColor).
				Foreground(lipgloss.Color("#FFFFFF"))

	progressEmptyStyle = lipgloss.NewStyle().
				Background(lipgloss.Color("#333333"))

	percentageStyle = lipgloss.NewStyle().
			Foreground(primaryColor).
			Bold(true).
			PaddingLeft(2)
)

// renderFileList renders the list of files with their statuses
func renderFileList(files []string, fileStatuses map[string]generator.FileStatus) string {
	var result strings.Builder

	for _, filename := range files {
		status := fileStatuses[filename]
		icon := statusStyles[status].Render(statusIcons[status])
		line := fileStyle.Render(fmt.Sprintf("%s %s", icon, filename))
		result.WriteString(line + "\n")
	}

	return result.String()
}

// renderProgressBar renders a fancy progress bar
func RenderProgressBar(completed, total int) string {
	percent := float64(0)

	if total > 0 {
		percent = float64(completed) / float64(total)
	}
	percentInt := int(percent * 100)

	// Get terminal width for responsive progress bar
	width := 50 // Default width if we can't detect terminal
	if w, _, err := term.GetSize(int(os.Stdout.Fd())); err == nil && w > 20 {
		width = w - 20 // Leave room for percentage
	}

	filledWidth := int(float64(width) * percent)
	emptyWidth := width - filledWidth

	filled := progressFillStyle.Render(strings.Repeat("█", filledWidth))
	empty := progressEmptyStyle.Render(strings.Repeat("░", emptyWidth))

	bar := filled + empty
	percentage := percentageStyle.Render(fmt.Sprintf("%d%%", percentInt))

	return progressBarStyle.Render(bar + percentage)
}

func UpdateScreen(title string, files []string, fileStatuses map[string]generator.FileStatus, completed int, total int) {
	// Sample files to process

	// Set up terminal for rendering
	fmt.Print("\033[?25l")       // Hide cursor
	defer fmt.Print("\033[?25h") // Show cursor on exit

	// Clear screen and move cursor to top-left
	fmt.Print("\033[H\033[2J")

	// Render title
	fmt.Println(titleStyle.Render(title))

	// Render file list
	fmt.Print(renderFileList(files, fileStatuses))

	// Render progress bar
	fmt.Println(RenderProgressBar(completed, total))

}
