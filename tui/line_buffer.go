package tui

import (
	"github.com/mitchellh/go-wordwrap"
	"strings"
	"sync"
)

type LineBuffer struct {
	mutex    sync.Mutex
	width    int
	messages []string
	lines    []string
}

func NewLineBuffer(width int) *LineBuffer {
	return &LineBuffer{
		width:    width,
		messages: []string{},
		lines:    []string{},
	}
}

func (s *LineBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	l := string(p)
	s.messages = append(s.messages, l)
	wl := s.wrap(l)
	s.lines = append(s.lines, wl...)

	return len(p), nil
}

func (s *LineBuffer) SetWidth(width int) {
	if width < 1 {
		return
	}

	s.mutex.Lock()
	defer s.mutex.Unlock()

	if width == s.width {
		return
	}

	s.width = width

	chunks := make([][]string, len(s.messages))

	var wg sync.WaitGroup
	wg.Add(len(s.messages))

	// parallelize this as it will get really slow the bigger the buffer is
	for i, l := range s.messages {
		go func(i int, l string) {
			defer wg.Done()

			wl := s.wrap(l)
			chunks[i] = append(chunks[i], wl...)
		}(i, l)
	}

	wg.Wait()

	s.lines = []string{}
	for _, c := range chunks {
		s.lines = append(s.lines, c...)
	}
}

func (s *LineBuffer) Lines(from, to int) []string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if from < 0 {
		from = 0
	}

	if to < 0 {
		to = 0
	}

	if to > len(s.lines) {
		to = len(s.lines)
	}

	return s.lines[from:to]
}

func (s *LineBuffer) Len() int {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return len(s.lines)
}

func (s *LineBuffer) wrap(line string) []string {
	wl := wordwrap.WrapString(line, uint(s.width))
	return strings.Split(wl, "\n")
}
