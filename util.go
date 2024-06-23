package main

import (
	"regexp"
	"strings"

	"github.com/fatih/color"
)

var (
	errorColor   = color.New(color.FgRed, color.Bold).SprintFunc()
	warningColor = color.New(color.FgYellow, color.Bold).SprintFunc()
	infoColor    = color.New(color.FgBlue, color.Bold).SprintFunc()
	ipColor      = color.New(color.FgCyan, color.Bold).SprintFunc()
	urlColor     = color.New(color.FgGreen, color.Bold).SprintFunc()

	ipRegex  = regexp.MustCompile(`\b(?:[0-9]{1,3}\.){3}[0-9]{1,3}\b`)
	urlRegex = regexp.MustCompile(`\bhttps?://\S+\b`)
)

type ColorFunc func(a ...interface{}) string

func convertForRegex(fn func(...interface{}) string) func(string) string {
	return func(s string) string {
		return fn([]interface{}{s})
	}
}

// Apply syntax highlighting
func prettify(line string) string {
	highlightedLine := strings.ReplaceAll(line, "ERROR", errorColor("ERROR"))
	highlightedLine = strings.ReplaceAll(highlightedLine, "WARNING", warningColor("WARNING"))
	highlightedLine = strings.ReplaceAll(highlightedLine, "INFO", infoColor("INFO"))

	highlightedLine = ipRegex.ReplaceAllStringFunc(highlightedLine, convertForRegex(ipColor))
	highlightedLine = urlRegex.ReplaceAllStringFunc(highlightedLine, convertForRegex(urlColor))

	return highlightedLine
}

func parseLogLine(line string) []string {
	// TODO: write this function
	return nil
}
