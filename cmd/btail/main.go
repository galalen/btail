package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/galalen/btail/pkg/app"
	"github.com/galalen/btail/pkg/config"
	"github.com/galalen/btail/pkg/tail"
)

func main() {
	lines := flag.Int("n", 10, "number of lines to display")
	follow := flag.Bool("f", false, "follow the file for new lines")
	flag.Parse()

	if len(flag.Args()) < 1 {
		fmt.Println("Error: Please provide a filename")
		os.Exit(1)
	}
	filename := flag.Args()[0]

	cfg := config.Config{
		Lines:        *lines,
		Follow:       *follow,
		UIBufferSize: 500,
		BufferSize:   500,
	}

	t, err := tail.TailFile(filename, cfg)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}

	if err := app.Run(t); err != nil {
		log.Fatalf("Error running UI: %v", err)
	}
}
