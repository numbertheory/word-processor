package editor

import (
	"path/filepath"
	"testing"
)

func TestDetectFormat(t *testing.T) {
	tests := []struct {
		path string
		want DocumentFormat
	}{
		{"notes.html", FormatHTML},
		{"notes.HTM", FormatHTML},
		{"readme.md", FormatMarkdown},
		{"doc.MARKDOWN", FormatMarkdown},
		{"plain.txt", FormatPlainText},
		{"plain.TEXT", FormatPlainText},
		{"noext", FormatHTML},
		{"", FormatHTML},
	}

	for _, tc := range tests {
		if got := DetectFormat(tc.path); got != tc.want {
			t.Errorf("DetectFormat(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestDisplayName(t *testing.T) {
	if got := DisplayName(""); got != "Untitled" {
		t.Errorf("DisplayName(\"\") = %q, want Untitled", got)
	}
	if got := DisplayName("/tmp/docs/letter.html"); got != "letter.html" {
		t.Errorf("DisplayName(...) = %q, want letter.html", got)
	}
}

func TestComputeStats(t *testing.T) {
	tests := []struct {
		name string
		text string
		want DocumentStats
	}{
		{
			name: "empty",
			text: "",
			want: DocumentStats{Words: 0, Characters: 0, Lines: 1},
		},
		{
			name: "single line",
			text: "Hello, world!",
			want: DocumentStats{Words: 2, Characters: 13, Lines: 1},
		},
		{
			name: "multi line with unicode",
			text: "Café résumé\nSecond line\tthree",
			want: DocumentStats{Words: 5, Characters: 29, Lines: 2},
		},
		{
			name: "whitespace only",
			text: "   \n  \t ",
			want: DocumentStats{Words: 0, Characters: 8, Lines: 2},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got := ComputeStats(tc.text)
			if got != tc.want {
				t.Errorf("ComputeStats(%q) = %+v, want %+v", tc.text, got, tc.want)
			}
		})
	}
}

func TestReadWriteDocumentFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "sample.html")
	content := "<p><b>Hello</b> Qt6</p>"

	if err := WriteDocumentFile(path, content); err != nil {
		t.Fatalf("WriteDocumentFile failed: %v", err)
	}

	got, err := ReadDocumentFile(path)
	if err != nil {
		t.Fatalf("ReadDocumentFile failed: %v", err)
	}
	if got != content {
		t.Errorf("ReadDocumentFile = %q, want %q", got, content)
	}
}

func TestInjectAndExtractHTMLMargins(t *testing.T) {
	margins := PageMargins{Top: 0.75, Right: 1.25, Bottom: 0.50, Left: 1.50}
	baseHTML := "<!DOCTYPE HTML><html><head><meta charset=\"utf-8\" /></head><body><p>Hi</p></body></html>"

	withMeta := InjectHTMLMargins(baseHTML, margins)
	got, ok := ExtractHTMLMargins(withMeta)
	if !ok {
		t.Fatalf("ExtractHTMLMargins returned ok=false on injected HTML:\n%s", withMeta)
	}
	if got != margins {
		t.Errorf("ExtractHTMLMargins = %+v, want %+v", got, margins)
	}

	// Updating existing meta tag should not duplicate it.
	updated := PageMargins{Top: 2.0, Right: 2.0, Bottom: 2.0, Left: 2.0}
	withUpdated := InjectHTMLMargins(withMeta, updated)
	gotUpdated, ok := ExtractHTMLMargins(withUpdated)
	if !ok || gotUpdated != updated {
		t.Errorf("updated margins = %+v (ok=%v), want %+v", gotUpdated, ok, updated)
	}

	// HTML without <head> should still round-trip.
	noHead := InjectHTMLMargins("<p>Plain fragment</p>", margins)
	if gotNoHead, ok := ExtractHTMLMargins(noHead); !ok || gotNoHead != margins {
		t.Errorf("no-head margins = %+v (ok=%v), want %+v", gotNoHead, ok, margins)
	}

	// Missing or invalid meta returns defaults and ok=false.
	if def, ok := ExtractHTMLMargins("<p>No meta</p>"); ok || def != DefaultPageMargins() {
		t.Errorf("missing meta = %+v (ok=%v), want default with ok=false", def, ok)
	}
	if def, ok := ExtractHTMLMargins(`<meta name="word-processor-margins" content="bad" />`); ok || def != DefaultPageMargins() {
		t.Errorf("malformed meta = %+v (ok=%v), want default with ok=false", def, ok)
	}

	// Browser page style should be injected and cleanly stripped when reopening.
	if stripped := StripHTMLPageStyle(withUpdated); stripped == withUpdated {
		t.Errorf("StripHTMLPageStyle did not remove style block from:\n%s", withUpdated)
	} else if _, ok := ExtractHTMLMargins(stripped); !ok {
		t.Errorf("StripHTMLPageStyle unexpectedly removed margin meta tag")
	}
}

func TestInchesToPx(t *testing.T) {
	if got := InchesToPx(1.0); got != 96.0 {
		t.Errorf("InchesToPx(1.0) = %v, want 96.0", got)
	}
	if got := InchesToPx(-0.5); got != 0 {
		t.Errorf("InchesToPx(-0.5) = %v, want 0", got)
	}
}
