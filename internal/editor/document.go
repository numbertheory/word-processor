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

const (
	pixelsPerInch      = 96.0
	MaxPageWidthInches = 10.2
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

// PageMargins stores page margins in inches.
type PageMargins struct {
	Top    float64
	Right  float64
	Bottom float64
	Left   float64
}

// PageSetup stores per-document page layout settings (margins and optional page colors).
type PageSetup struct {
	Margins         PageMargins
	BackgroundColor string // Hex color "#rrggbb", or "" for default theme background
	TextColor       string // Hex color "#rrggbb", or "" for default theme text color
}

// DefaultPageMargins returns standard 1-inch margins on all sides.
func DefaultPageMargins() PageMargins {
	return PageMargins{Top: 1.0, Right: 1.0, Bottom: 1.0, Left: 1.0}
}

// DefaultPageSetup returns default margins and no custom background color.
func DefaultPageSetup() PageSetup {
	return PageSetup{Margins: DefaultPageMargins()}
}

// InchesToPx converts inches to device-independent pixels (96 DPI).
func InchesToPx(inches float64) float64 {
	if inches < 0 {
		return 0
	}
	return inches * pixelsPerInch
}

var marginMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-margins["']\s+content=["']([^"']*)["']\s*/?>\s*`)
var bgMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-page-bg["']\s+content=["']([^"']*)["']\s*/?>\s*`)
var fgMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-page-fg["']\s+content=["']([^"']*)["']\s*/?>\s*`)
var hexColorRe = regexp.MustCompile(`^#[0-9a-fA-F]{6}$`)
var pageStyleRe = regexp.MustCompile(`(?is)<style\s+id=["']word-processor-page-style["'][^>]*>.*?</style>\s*`)
var headOpenRe = regexp.MustCompile(`(?i)<head[^>]*>`)

// IsValidHexColor reports whether s is a #rrggbb hex color string.
func IsValidHexColor(s string) bool {
	return hexColorRe.MatchString(s)
}

// InjectHTMLMargins embeds or updates the word-processor-margins meta tag
// and responsive centered-page CSS for browsers in an HTML document.
func InjectHTMLMargins(html string, m PageMargins) string {
	return InjectHTMLPageSetup(html, PageSetup{Margins: m})
}

// InjectHTMLPageSetup embeds or updates page margin and background-color meta
// tags plus browser page-layout CSS in an HTML document.
func InjectHTMLPageSetup(html string, setup PageSetup) string {
	m := setup.Margins
	metaTags := fmt.Sprintf(
		`<meta name="word-processor-margins" content="%.2f,%.2f,%.2f,%.2f" />`,
		m.Top, m.Right, m.Bottom, m.Left,
	)

	bgCSS := ""
	if IsValidHexColor(setup.BackgroundColor) {
		metaTags += fmt.Sprintf("\n"+`<meta name="word-processor-page-bg" content="%s" />`, setup.BackgroundColor)
		bgCSS = fmt.Sprintf(" html, body { background-color: %s !important; }", setup.BackgroundColor)
	}
	fgCSS := ""
	if IsValidHexColor(setup.TextColor) {
		metaTags += fmt.Sprintf("\n"+`<meta name="word-processor-page-fg" content="%s" />`, setup.TextColor)
		fgCSS = fmt.Sprintf(" color: %s !important;", setup.TextColor)
	}

	pageStyle := fmt.Sprintf(
		`<style id="word-processor-page-style">body { max-width: %.1fin; margin: 0 auto !important; padding: %.2fin %.2fin %.2fin %.2fin !important; box-sizing: border-box;%s } table[style*="-qt-table-type: root"] { margin: 0 !important; width: 100%%; }%s</style>`,
		MaxPageWidthInches, m.Top, m.Right, m.Bottom, m.Left, fgCSS, bgCSS,
	)

	cleaned := pageStyleRe.ReplaceAllString(html, "")
	cleaned = marginMetaRe.ReplaceAllString(cleaned, "")
	cleaned = bgMetaRe.ReplaceAllString(cleaned, "")
	cleaned = fgMetaRe.ReplaceAllString(cleaned, "")

	block := metaTags + "\n" + pageStyle
	if loc := headOpenRe.FindStringIndex(cleaned); loc != nil {
		return cleaned[:loc[1]] + "\n" + block + "\n" + cleaned[loc[1]:]
	}
	return fmt.Sprintf("<head>\n%s\n</head>\n%s", block, cleaned)
}

// ExtractHTMLPageSetup parses margin, page background-color, and default text-color meta tags from HTML.
func ExtractHTMLPageSetup(html string) PageSetup {
	margins, _ := ExtractHTMLMargins(html)
	bg := ""
	if matches := bgMetaRe.FindStringSubmatch(html); len(matches) >= 2 {
		candidate := strings.TrimSpace(matches[1])
		if IsValidHexColor(candidate) {
			bg = candidate
		}
	}
	fg := ""
	if matches := fgMetaRe.FindStringSubmatch(html); len(matches) >= 2 {
		candidate := strings.TrimSpace(matches[1])
		if IsValidHexColor(candidate) {
			fg = candidate
		}
	}
	return PageSetup{
		Margins:         margins,
		BackgroundColor: bg,
		TextColor:       fg,
	}
}

// StripHTMLPageStyle removes the browser-only page centering style block so
// QTextEdit does not double-apply margins when reopening an HTML file.
func StripHTMLPageStyle(html string) string {
	return pageStyleRe.ReplaceAllString(html, "")
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

// CanonicalPath normalizes a file path for reliable duplicate-tab detection.
func CanonicalPath(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return filepath.Clean(path)
	}
	if resolved, err := filepath.EvalSymlinks(abs); err == nil {
		return resolved
	}
	return abs
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
