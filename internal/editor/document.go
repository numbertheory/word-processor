package editor

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

const pixelsPerInch = 96.0

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

// PageMargins stores page margins in inches.
type PageMargins struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// DefaultPageMargins returns standard 1-inch margins on all sides.
func DefaultPageMargins() PageMargins {
	return PageMargins{Top: 1.0, Right: 1.0, Bottom: 1.0, Left: 1.0}
}

// InchesToPx converts inches to device-independent pixels (96 DPI).
func InchesToPx(inches float64) float64 {
	if inches < 0 {
		return 0
	}
	return inches * pixelsPerInch
}

var marginMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-margins["']\s+content=["']([^"']*)["']\s*/?>`)
var headOpenRe = regexp.MustCompile(`(?i)<head[^>]*>`)

// InjectHTMLMargins embeds or updates the word-processor-margins meta tag in an HTML document.
func InjectHTMLMargins(html string, m PageMargins) string {
	metaTag := fmt.Sprintf(
		`<meta name="word-processor-margins" content="%.2f,%.2f,%.2f,%.2f" />`,
		m.Top, m.Right, m.Bottom, m.Left,
	)

	if marginMetaRe.MatchString(html) {
		return marginMetaRe.ReplaceAllString(html, metaTag)
	}
	if loc := headOpenRe.FindStringIndex(html); loc != nil {
		return html[:loc[1]] + "\n" + metaTag + html[loc[1]:]
	}
	return fmt.Sprintf("<head>\n%s\n</head>\n%s", metaTag, html)
}

// ExtractHTMLMargins parses the word-processor-margins meta tag from HTML, if present.
func ExtractHTMLMargins(html string) (PageMargins, bool) {
	matches := marginMetaRe.FindStringSubmatch(html)
	if len(matches) < 2 {
		return DefaultPageMargins(), false
	}
	parts := strings.Split(matches[1], ",")
	if len(parts) != 4 {
		return DefaultPageMargins(), false
	}

	vals := make([]float64, 4)
	for i, p := range parts {
		v, err := strconv.ParseFloat(strings.TrimSpace(p), 64)
		if err != nil || v < 0 {
			return DefaultPageMargins(), false
		}
		vals[i] = v
	}
	return PageMargins{
		Top:    vals[0],
		Right:  vals[1],
		Bottom: vals[2],
		Left:   vals[3],
	}, true
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
