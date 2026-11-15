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
