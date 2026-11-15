package editor

import (
	"os"
	"path/filepath"
	"strings"
	"unicode"
	"unicode/utf8"
)

// DocumentFormat represents the serialization format of a document.
type DocumentFormat int

const (
	FormatHTML DocumentFormat = iota
	FormatMarkdown
	FormatPlainText
)

// DocumentStats holds word, character, and line counts for a document.
type DocumentStats struct {
	Words      int
	Characters int
	Lines      int
}

// DetectFormat determines the document format from a file path's extension.
// Defaults to FormatHTML for unknown or empty extensions so rich text is preserved.
func DetectFormat(path string) DocumentFormat {
	switch strings.ToLower(filepath.Ext(path)) {
	case ".md", ".markdown":
		return FormatMarkdown
	case ".txt", ".text":
		return FormatPlainText
	case ".html", ".htm":
		return FormatHTML
	default:
		return FormatHTML
	}
}

// DisplayName returns a human-readable file name for window titles.
func DisplayName(path string) string {
	if path == "" {
		return "Untitled"
	}
	return filepath.Base(path)
}

// ComputeStats calculates word, character, and line counts from plain text.
func ComputeStats(text string) DocumentStats {
	if text == "" {
		return DocumentStats{Words: 0, Characters: 0, Lines: 1}
	}

	words := 0
	inWord := false
	lines := 1

	for _, r := range text {
		if r == '\n' {
			lines++
		}
		if unicode.IsSpace(r) {
			inWord = false
		} else if !inWord {
			inWord = true
			words++
		}
	}

	return DocumentStats{
		Words:      words,
		Characters: utf8.RuneCountInString(text),
		Lines:      lines,
	}
}

// ReadDocumentFile reads a document from disk as UTF-8 text.
func ReadDocumentFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// WriteDocumentFile writes document content to disk.
func WriteDocumentFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}
