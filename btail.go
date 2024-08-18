package main

import (
	"bufio"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"time"

	"github.com/fsnotify/fsnotify"
)

type Config struct {
	Lines  int
	Follow bool
}

type Tail struct {
	Filename string
	Lines    chan Line
	Config   Config
	file     *os.File
	fileSize int64
	watcher  *fsnotify.Watcher
	done     chan struct{}
}

type Line struct {
	Text string
	Time time.Time
}

func TailFile(Filename string, config Config) (*Tail, error) {
	if config.Lines <= 0 {
		config.Lines = 5
	}

	t := &Tail{
		Filename: Filename,
		Lines:    make(chan Line),
		Config:   config,
		done:     make(chan struct{}),
	}
	var err error

	if err = t.openFile(); err != nil {
		return nil, err
	}

	if config.Follow {
		t.watcher, err = fsnotify.NewWatcher()
		if err != nil {
			t.file.Close()
			return nil, fmt.Errorf("failed to create watcher: %v", err)
		}

		err = t.watcher.Add(Filename)
		if err != nil {
			t.file.Close()
			t.watcher.Close()
			return nil, fmt.Errorf("failed to add file to watcher: %v", err)
		}
	}

	go t.tail()

	return t, nil
}

func (t *Tail) openFile() error {
	file, err := os.Open(t.Filename)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", t.Filename, err)
	}
	t.file = file

	info, err := file.Stat()
	if err != nil {
		t.file.Close()
		return fmt.Errorf("failed to stat file %s: %v", t.Filename, err)
	}
	t.fileSize = info.Size()

	return nil
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
	buffer := make([]byte, 1024*1024)
	offset := t.fileSize
	lineCount := 0
	lines := make([]Line, 0, t.Config.Lines)

	for lineCount < t.Config.Lines*2 && offset > 0 {
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
				if lineCount > t.Config.Lines*2 {
					offset += int64(i) + 1
					break
				}
			}
		}
	}

	_, err := t.file.Seek(offset, io.SeekStart)
	if err != nil {
		return nil, err
	}

	scanner := bufio.NewScanner(t.file)
	for scanner.Scan() && len(lines) < t.Config.Lines {
		trimmedText := strings.TrimSpace(scanner.Text())
		if trimmedText != "" {
			lines = append(lines, Line{Text: trimmedText, Time: time.Now()})
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}

func (t *Tail) readNewLines(reader *bufio.Reader) {
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if err == io.EOF {
				break
			}
			// show error in info area
			return
		}
		t.Lines <- Line{line, time.Now()}
		t.fileSize += int64(len(line))
	}
}

func (t *Tail) followFile() error {
	reader := bufio.NewReader(t.file)

	for {
		select {
		case event, ok := <-t.watcher.Events:
			if !ok {
				return nil
			}
			if event.Op&fsnotify.Write == fsnotify.Write {
				t.readNewLines(reader)
			}
		case _, ok := <-t.watcher.Errors:
			if !ok {
				return nil
			}
			// show error in info area
		case <-t.done:
			return nil
		}
	}
}

func (t *Tail) Stop() {
	close(t.done)
}
