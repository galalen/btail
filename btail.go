package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"time"
)

func Tail(filename string, lines int, follow bool) error {
	file, err := os.Open(filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	result, offset, err := readLastNLines(file, fileSize, lines)
	if err != nil {
		return err
	}

	for _, line := range result {
		fmt.Println(prettify(line))
	}

	if follow {
		if err := followFile(file, offset); err != nil {
			return err
		}
	}

	return nil
}

func readLastNLines(file *os.File, fileSize int64, lines int) ([]string, int64, error) {
	var lineCount int
	var offset int64 = 0
	chunkSize := int64(4096)
	var result []string
	var lineBuffer []byte

	for lineCount < lines && offset < fileSize {
		if offset+chunkSize > fileSize {
			chunkSize = fileSize - offset
		}
		offset += chunkSize
		file.Seek(-offset, io.SeekEnd)
		chunk := make([]byte, chunkSize)
		_, err := file.Read(chunk)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to read file: %w", err)
		}

		for i := len(chunk) - 1; i >= 0; i-- {
			if chunk[i] == '\n' {
				if len(lineBuffer) > 0 {
					lineCount++
					result = append([]string{string(lineBuffer)}, result...)
					lineBuffer = nil
				}
				if lineCount == lines {
					break
				}
			} else {
				lineBuffer = append([]byte{chunk[i]}, lineBuffer...)
			}
		}
	}

	if lineCount < lines && len(lineBuffer) > 0 {
		result = append([]string{string(lineBuffer)}, result...)
	}

	return result, fileSize - offset, nil
}

func followFile(file *os.File, offset int64) error {
	for {
		_, err := file.Seek(offset, 0)
		if err != nil {
			return fmt.Errorf("error seeking file: %w", err)
		}

		reader := bufio.NewReader(file)
		for {
			line, err := reader.ReadString('\n')
			if err != nil {
				break
			}
			fmt.Println(prettify(line))
			offset += int64(len(line))
		}
		time.Sleep(1 * time.Second)
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

	if err := Tail(*filename, *lines, *follow); err != nil {
		fmt.Printf("Error: %s\n", err)
		os.Exit(1)
	}
}
