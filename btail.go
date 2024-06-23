package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type Config struct {
	Filename string
	Lines    int
	Follow   bool
}

type Tail struct {
	Filename string
	Lines    chan Line
}

type Line struct {
	Text string
	Time time.Time
}

func TailFile(config Config) (Tail, error) {
	tail := Tail{
		Filename: config.Filename,
		Lines:    make(chan Line),
	}

	go func() {
		defer close(tail.Lines)

		file, err := os.Open(config.Filename)
		if err != nil {
			log.Fatalf("failed to open file: %v", err)
		}
		defer file.Close()

		fileInfo, err := file.Stat()
		if err != nil {
			log.Fatalf("failed to get file info: %v", err)
		}
		fileSize := fileInfo.Size()

		lines, offset, err := readLastNLines(file, fileSize, config.Lines)
		if err != nil {
			log.Fatalf("failed to read lines from file: %v", err)
		}

		for _, line := range lines {
			tail.Lines <- line
		}

		if config.Follow {
			if err := followFile(file, offset, tail.Lines); err != nil {
				log.Fatalf("failed to follow file: %v", err)
			}
		}
	}()
	return tail, nil
}

func readLastNLines(file *os.File, fileSize int64, lines int) ([]Line, int64, error) {
	var lineCount int
	var offset int64 = 0
	chunkSize := int64(4096)
	var result []Line
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
					result = append([]Line{{string(lineBuffer), time.Now()}}, result...)
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
		result = append([]Line{{string(lineBuffer), time.Now()}}, result...)
	}

	return result, fileSize - offset, nil
}

func followFile(file *os.File, offset int64, lines chan<- Line) error {
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
			lines <- Line{line, time.Now()}
			offset += int64(len(line))
		}
		time.Sleep(1 * time.Second)
	}
}
