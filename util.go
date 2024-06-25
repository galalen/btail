package main

import (
	"fmt"
	"regexp"
)

var (
	ipRegex       = regexp.MustCompile(`\b(?:\d{1,3}\.){3}\d{1,3}(:\d{1,5})?\b`)
	urlRegex      = regexp.MustCompile(`\b(?:https?|ftp|rtmp|smtp)://\S+`)
	filePathRegex = regexp.MustCompile(`\b[A-Za-z]:\\\S+|\b/[^ ]+`)
)

func highlightPatterns(text string) string {
	text = ipRegex.ReplaceAllStringFunc(text, func(ip string) string {
		return fmt.Sprintf("[yellow::b]%s[#32CD32::-]", ip)
	})

	text = urlRegex.ReplaceAllStringFunc(text, func(url string) string {
		return fmt.Sprintf("[green::u]%s[#32CD32::-]", url)
	})

	text = filePathRegex.ReplaceAllStringFunc(text, func(path string) string {
		return fmt.Sprintf("[green::u]%s[#32CD32::-]", path)
	})

	return text
}

func parseLogLine(line string) []string {
	// TODO: write this function
	return nil
}
