package main

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

var (
	baseStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(lipgloss.Color("240"))

	titleStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("228")).
			Background(lipgloss.Color("63")).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("241")).
			Background(lipgloss.Color("237"))

	searchStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("205")).
			Background(lipgloss.Color("237"))

	searchInputStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("205")).
				Background(lipgloss.Color("237")).
				Padding(0, 1)

	statusMessageStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("241")).
				Background(lipgloss.Color("237")).
				Padding(0, 1)

	ipStyle = lipgloss.NewStyle().
		Background(lipgloss.Color("25")).
		Foreground(lipgloss.Color("231"))

	urlStyle = lipgloss.NewStyle().
			Background(lipgloss.Color("90")).
			Foreground(lipgloss.Color("231"))

	filePathStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("231")).
			Underline(true)

	ipRegex       = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(:\d{1,5})?\b`)
	urlRegex      = regexp.MustCompile(`\b(?:https?|ftp|rtmp|smtp)://\S+`)
	filePathRegex = regexp.MustCompile(`\b[A-Za-z]:\\\S+|\b/[^ ]+`)
)

func highlightPatterns(text string) string {
	text = ipRegex.ReplaceAllStringFunc(text, func(ip string) string {
		return ipStyle.Render(ip)
	})

	text = urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		return urlStyle.Render(url)
	})

	text = filePathRegex.ReplaceAllStringFunc(text, func(path string) string {
		return filePathStyle.Render(path)
	})

	return text
}

func highlightSearch(content, term string) string {
	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(term))
	return re.ReplaceAllStringFunc(content, func(match string) string {
		return searchStyle.Render(match)
	})
}
