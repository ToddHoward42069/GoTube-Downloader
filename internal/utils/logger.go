package utils

import (
	"strings"
	"sync"
)

// LogBuffer is a thread-safe circular buffer for log lines
type LogBuffer struct {
	lines    []string
	maxLines int
	mu       sync.Mutex
	dirty    bool // True if content changed since last read
}

func NewLogBuffer(max int) *LogBuffer {
	return &LogBuffer{
		lines:    make([]string, 0, max),
		maxLines: max,
	}
}

// Write adds a line to the buffer
func (l *LogBuffer) Write(text string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// Split incoming text by newlines to handle bulk updates
	parts := strings.Split(text, "\n")
	for _, p := range parts {
		if strings.TrimSpace(p) == "" {
			continue
		}

		if len(l.lines) >= l.maxLines {
			// Shift: Remove first element, append new
			l.lines = l.lines[1:]
		}
		l.lines = append(l.lines, p)
	}
	l.dirty = true
}

// String returns the full log content joined by newlines
func (l *LogBuffer) String() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

// HasChanged checks if we need to update the UI
func (l *LogBuffer) HasChanged() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.dirty
}

// MarkRead resets the dirty flag
func (l *LogBuffer) MarkRead() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.dirty = false
}

// Clear empties the buffer
func (l *LogBuffer) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = []string{}
	l.dirty = true
}
