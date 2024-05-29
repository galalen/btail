package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type TailConfig struct {
	Filename string
	Lines    int
	Follow   bool
}

type TailResponse struct {
	Lines chan string
}

func Tail(config TailConfig) (TailResponse, error) {
	res := TailResponse{make(chan string)}

	go func() {
		defer close(res.Lines)

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
			res.Lines <- line
		}

		if config.Follow {
			if err := followFile(file, offset, res.Lines); err != nil {
				log.Fatalf("failed to follow file: %v", err)
			}
		}
	}()
	return res, nil
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

func followFile(file *os.File, offset int64, lines chan<- string) error {
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
			lines <- line
			offset += int64(len(line))
		}
		time.Sleep(1 * time.Second)
	}
}
