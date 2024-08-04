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
	lines := flag.Int("n", 10, "number of lines to display")
	follow := flag.Bool("f", false, "follow the file for new lines")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Error: Please provide a filename")
		os.Exit(1)
	}
	filename := flag.Args()[0]

	config := Config{
		Lines:  *lines,
		Follow: *follow,
	}

	tail, err := TailFile(filename, config)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	// runWithoutUI(tail)
	NewBtailApp(tail).Run()
}
