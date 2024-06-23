package main

import (
	"flag"
	"fmt"
	"os"
)

func runWithoutUI(tail Tail) {
	for line := range tail.Lines {
		fmt.Println(prettify(line.Text))
	}
}

func main() {
	filename := flag.String("file", "", "File to tail")
	lines := flag.Int("lines", 10, "Number of lines to display")
	follow := flag.Bool("follow", false, "Follow the file for new lines (similar to tail -f)")

	flag.Parse()

	if *filename == "" {
		fmt.Println("Usage: go run main.go --file <filename> [--lines <number_of_lines>] [--follow]")
		os.Exit(1)
	}

	config := Config{
		Filename: *filename,
		Lines:    *lines,
		Follow:   *follow,
	}

	tail, err := TailFile(config)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	//runWithoutUI(tail)
	appUI(tail)
}
