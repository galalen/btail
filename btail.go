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

	lines, err := t.readLastNLines()
	if err != nil {
		log.Printf("failed to read lines from file: %v", err)
		return
	}

	for _, line := range lines {
		t.Lines <- line
	}

	if t.Config.Follow {
		if err := t.followFile(); err != nil {
			log.Printf("failed to follow file: %v", err)
		}
	}
}

func (t *Tail) readLastNLines() ([]Line, error) {
	stat, err := t.file.Stat()
	if err != nil {
		return nil, err
	}

	fileSize := stat.Size()
	buffer := make([]byte, 1024*1024) // 1MB buffer
	offset := fileSize
	lineCount := 0
	lines := make([]Line, 0, t.Config.Lines)

	for lineCount < t.Config.Lines && offset > 0 {
		readSize := int64(len(buffer))
		if offset < readSize {
			readSize = offset
		}
		offset -= readSize

		_, err := t.file.Seek(offset, io.SeekStart)
		if err != nil {
			return nil, err
		}

		bytesRead, err := t.file.Read(buffer[:readSize])
		if err != nil && err != io.EOF {
			return nil, err
		}

		for i := bytesRead - 1; i >= 0; i-- {
			if buffer[i] == '\n' {
				lineCount++
				if lineCount > t.Config.Lines {
					offset += int64(i) + 1
					break
				}
			}
		}
	}

	_, err = t.file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(t.file)
	for scanner.Scan() && len(lines) < t.Config.Lines {
		lines = append(lines, Line{Text: scanner.Text(), Time: time.Now()})
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (t *Tail) followFile() error {
	_, err := t.file.Seek(0, io.SeekEnd)
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
