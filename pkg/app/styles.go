package app

import (
	"regexp"

	"github.com/charmbracelet/lipgloss"
)

var (
	blueColor       = lipgloss.Color("#00afd7")
	pinkColor       = lipgloss.Color("#ff5faf")
	yellowColor     = lipgloss.Color("#ffff87")
	orangeColor     = lipgloss.Color("#ff8700")
	redColor        = lipgloss.Color("#ff0000")
	tealColor       = lipgloss.Color("#5f87ff")
	lightGrayColor  = lipgloss.Color("#D3D3D3")
	darkerGrayColor = lipgloss.Color("#3a3a3a")
	silverColor     = lipgloss.Color("#1b1b1b")
	greenColor      = lipgloss.Color("#00ff00")
	lightGreenColor = lipgloss.Color("#90ee90")

	baseStyle = lipgloss.NewStyle().
			Margin(0).
			Padding(0).
			BorderStyle(lipgloss.NormalBorder()).
			BorderForeground(darkerGrayColor)

	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(yellowColor).
			Padding(0, 1)

	statusBarStyle = lipgloss.NewStyle().
			Bold(false).
			Foreground(pinkColor).
			Background(silverColor)

	searchInputStyle = lipgloss.NewStyle().
				Foreground(pinkColor)

	pinkStyle = lipgloss.NewStyle().
			Foreground(pinkColor)

	searchMatchStyle = lipgloss.NewStyle().
				Foreground(darkerGrayColor).
				Background(pinkColor).
				Bold(true)

	timeStyle = lipgloss.NewStyle().
			Foreground(lightGreenColor)

	ipStyle = lipgloss.NewStyle().
		Foreground(blueColor)

	urlStyle = lipgloss.NewStyle().
			Foreground(tealColor).
			Underline(true)

	methodStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(yellowColor)

	bracketsStyle = lipgloss.NewStyle().
			Foreground(yellowColor)

	searchModeStyle = lipgloss.NewStyle().
			Foreground(lightGreenColor)

	ipRegex      = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(:\d{1,5})?\b`)
	urlRegex     = regexp.MustCompile(`\b(?:https?|ftp|rtmp|smtp)://\S+`)
	methodRegex  = regexp.MustCompile(`\b(GET|POST|PUT|DELETE|PATCH|HEAD|OPTIONS)\b`)
	errorRegex   = regexp.MustCompile(`(?i)\b(error|exception|failed|failure|timeout|denied)\b`)
	warningRegex = regexp.MustCompile(`(?i)\b(warning|warn|deprecated)\b`)
)

func highlightPatterns(text string) string {
	text = ipRegex.ReplaceAllStringFunc(text, func(ip string) string {
		return ipStyle.Render(ip)
	})

	text = urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		return urlStyle.Render(url)
	})

	text = methodRegex.ReplaceAllStringFunc(text, func(method string) string {
		return methodStyle.Render(method)
	})

	text = errorRegex.ReplaceAllStringFunc(text, func(match string) string {
		return lipgloss.NewStyle().Foreground(redColor).Bold(true).Render(match)
	})

	text = warningRegex.ReplaceAllStringFunc(text, func(match string) string {
		return lipgloss.NewStyle().Foreground(orangeColor).Render(match)
	})

	return text
}
