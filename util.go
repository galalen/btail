package main

import (
	"fmt"
	"regexp"
	"strings"
)

const primaryColor = "#90EE90"

var (
	ipRegex       = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(:\d{1,5})?\b`)
	urlRegex      = regexp.MustCompile(`\b(?:https?|ftp|rtmp|smtp)://\S+`)
	filePathRegex = regexp.MustCompile(`\b[A-Za-z]:\\\S+|\b/[^ ]+`)
	// TODO: add json parsing
)

func highlightKeyword(text, keyword string) string {
	return strings.ReplaceAll(text, keyword, fmt.Sprintf("[red]%s[%s]", keyword, primaryColor))
}

func highlightPatterns(text string) string {
	text = ipRegex.ReplaceAllStringFunc(text, func(ip string) string {
		return fmt.Sprintf("[yellow::b]%s[%s::-]", ip, primaryColor)
	})

	text = urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		return fmt.Sprintf("[green::u]%s[%s::-]", url, primaryColor)
	})

	text = filePathRegex.ReplaceAllStringFunc(text, func(path string) string {
		return fmt.Sprintf("[green::u]%s[%s::-]", path, primaryColor)
	})

	return text
}

func parseLogLine(line string) []string {
	// TODO: write this function
	return nil
}
