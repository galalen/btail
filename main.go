package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

func runWithoutUI(tail *Tail) {
	rowCount := 1
	for line := range tail.Lines {
		timestamp := line.Time.Format(time.RFC3339)
		fmt.Printf("[%d] %s - %s", rowCount, timestamp, line.Text)

		rowCount++
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
		Lines:  *lines,
		Follow: *follow,
	}

	tail, err := TailFile(*filename, config)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	// runWithoutUI(tail)
	NewBtailApp(tail).Run()
}
