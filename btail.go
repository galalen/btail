package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Filename string
	Lines    int
	Follow   bool
}

type Tail struct {
	Filename string
	Lines    chan Line
	Config   Config
	file     *os.File
	watcher  *fsnotify.Watcher
	done     chan struct{}
}

type Line struct {
	Text string
	Time time.Time
}

func TailFile(config Config) (*Tail, error) {
	t := &Tail{
		Filename: config.Filename,
		Lines:    make(chan Line),
		Config:   config,
		done:     make(chan struct{}),
	}

	var err error
	t.file, err = os.Open(config.Filename)
	if err != nil {
		return nil, fmt.Errorf("failed to open file %s: %v", config.Filename, err)
	}

	if config.Follow {
		t.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			t.file.Close()
			return nil, fmt.Errorf("failed to create watcher: %v", err)
		}
		err = t.watcher.Add(config.Filename)
		if err != nil {
			t.file.Close()
			t.watcher.Close()
			return nil, fmt.Errorf("failed to add file to watcher: %v", err)
		}
	}

	go t.tail()

	return t, nil
}

func (t *Tail) tail() {
	defer close(t.Lines)
	defer t.file.Close()
	if t.watcher != nil {
		defer t.watcher.Close()
	}

	lines, offset, err := t.readLastNLines()
	if err != nil {
		log.Printf("failed to read lines from file: %v", err)
		return
	}

	for _, line := range lines {
		t.Lines <- line
	}

	if t.Config.Follow {
		if err := t.followFile(offset); err != nil {
			log.Printf("failed to follow file: %v", err)
		}
	}
}

func (t *Tail) readLastNLines() ([]Line, int64, error) {
	fileInfo, err := t.file.Stat()
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get file info: %v", err)
	}

	fileSize := fileInfo.Size()

	var lineCount int
	var offset int64 = 0
	chunkSize := int64(4096)
	var result []Line
	var lineBuffer []byte

	for lineCount < t.Config.Lines && offset < fileSize {
		if offset+chunkSize > fileSize {
			chunkSize = fileSize - offset
		}

		offset += chunkSize
		t.file.Seek(-offset, io.SeekEnd)
		chunk := make([]byte, chunkSize)
		_, err := t.file.Read(chunk)
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
				if lineCount == t.Config.Lines {
					break
				}
			} else {
				lineBuffer = append([]byte{chunk[i]}, lineBuffer...)
			}
		}
	}

	if lineCount < t.Config.Lines && len(lineBuffer) > 0 {
		result = append([]Line{{string(lineBuffer), time.Now()}}, result...)
	}

	return result, fileSize - offset, nil
}

func (t *Tail) followFile(offset int64) error {
	_, err := t.file.Seek(offset, io.SeekStart)
	if err != nil {
		return fmt.Errorf("error seeking file: %w", err)
	}

	reader := bufio.NewReader(t.file)

	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				for {
					line, err := reader.ReadString('\n')
					if err != nil {
						if err == io.EOF {
							break
						}
						return fmt.Errorf("error reading file: %w", err)
					}
					t.Lines <- Line{line, time.Now()}
				}
			}
		case err, ok := <-t.watcher.Errors:
			if !ok {
				return nil
			}
			log.Printf("error from watcher: %v", err)
		case <-t.done:
			return nil
		}
	}
}

func (t *Tail) Stop() {
	close(t.done)
}
