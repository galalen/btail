package ui

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

var (
	primaryColor    = lipgloss.Color("39")  // Blue
	secondaryColor  = lipgloss.Color("205") // Pink
	highlightColor  = lipgloss.Color("228") // Yellow
	warningColor    = lipgloss.Color("208") // Orange
	errorColor      = lipgloss.Color("196") // Red
	infoColor       = lipgloss.Color("69")  // Teal
	subtleColor     = lipgloss.Color("244") // Light gray
	verySubtleColor = lipgloss.Color("237") // Darker gray
	backgroundColor = lipgloss.Color("235") // Very dark gray
	textColor       = lipgloss.Color("252") // Almost white
	brightTextColor = lipgloss.Color("255") // White

	baseStyle = lipgloss.NewStyle().
			Margin(0).
			Padding(0).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(subtleColor)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor).
			Background(verySubtleColor).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Background(verySubtleColor).
			Foreground(subtleColor)

	searchStyle = lipgloss.NewStyle().
			Foreground(brightTextColor).
			Background(secondaryColor).
			Bold(true)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(brightTextColor).
				Background(secondaryColor)

	promptStyle = lipgloss.NewStyle().
			Foreground(secondaryColor)

	inputTextStyle = lipgloss.NewStyle().
			Foreground(brightTextColor)

	statusMessageStyle = lipgloss.NewStyle().
				Background(verySubtleColor).
				Foreground(highlightColor)

	timeStyle = lipgloss.NewStyle().
			Foreground(subtleColor)

	ipStyle = lipgloss.NewStyle().
		Background(primaryColor).
		Foreground(brightTextColor)

	urlStyle = lipgloss.NewStyle().
			Foreground(infoColor).
			Underline(true)

	methodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(highlightColor)

	filePathStyle = lipgloss.NewStyle().
			Foreground(subtleColor).
			Italic(true)

	bracketsStyle = lipgloss.NewStyle().
			Foreground(highlightColor)

	ipRegex       = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(:\d{1,5})?\b`)
	urlRegex      = regexp.MustCompile(`\b(?:https?|ftp|rtmp|smtp)://\S+`)
	filePathRegex = regexp.MustCompile(`\b[A-Za-z]:\\\S+|\b/[^\s:]+`)
	methodRegex   = regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\b`)
	errorRegex    = regexp.MustCompile(`(?i)\b(error|exception|failed|failure|timeout|denied)\b`)
	warningRegex  = regexp.MustCompile(`(?i)\b(warning|warn|deprecated)\b`)
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

	text = methodRegex.ReplaceAllStringFunc(text, func(method string) string {
		return methodStyle.Render(method)
	})

	text = errorRegex.ReplaceAllStringFunc(text, func(match string) string {
		return lipgloss.NewStyle().Foreground(errorColor).Bold(true).Render(match)
	})

	text = warningRegex.ReplaceAllStringFunc(text, func(match string) string {
		return lipgloss.NewStyle().Foreground(warningColor).Render(match)
	})

	return text
}

func highlightSearch(content, term string) string {
	if term == "" {
		return content
	}

	re := regexp.MustCompile(`(?i)` + regexp.QuoteMeta(term))
	return re.ReplaceAllStringFunc(content, func(match string) string {
		return searchStyle.Render(match)
	})
}
