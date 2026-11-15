package editor

import (
	"strings"
	"testing"
)

func TestRenderDrawingSVGAllShapes(t *testing.T) {
	shapes := []ShapeType{ShapeStar, ShapeCircle, ShapeRectangle, ShapeTriangle, ShapeArrow}
	for _, s := range shapes {
		item := NewDrawingItem("id-1", s, 100, 150)
		item.Background = "#fef08a"
		svg := RenderDrawingSVG(item)
		if !strings.HasPrefix(svg, "<svg ") || !strings.HasSuffix(svg, "</svg>") {
			t.Errorf("shape %q produced invalid SVG wrapper: %s", s, svg)
		}
		if !strings.Contains(svg, `fill="#fef08a"`) {
			t.Errorf("shape %q missing background rect: %s", s, svg)
		}
	}

	rotated := NewDrawingItem("id-rot", ShapeArrow, 50, 50)
	rotated.Rotation = -45
	svgRot := RenderDrawingSVG(rotated)
	if !strings.Contains(svgRot, `transform="rotate(-45.0 60.00 60.00)"`) {
		t.Errorf("rotated SVG missing expected transform: %s", svgRot)
	}

	note := NewNoteItem("id-note", 40, 60)
	note.FillColor = "#1e293b"
	note.BorderColor = "#38bdf8"
	note.Text = "Remember to check <margins> & fonts!"
	noteSVG := RenderDrawingSVG(note)
	if !strings.Contains(noteSVG, "Remember to check") || !strings.Contains(noteSVG, "&lt;margins&gt;") {
		t.Errorf("note SVG missing escaped text: %s", noteSVG)
	}
	if !strings.Contains(noteSVG, `fill="#f8fafc"`) {
		t.Errorf("dark note background should use light contrasting text: %s", noteSVG)
	}
}

func TestInjectStripExtractHTMLDrawings(t *testing.T) {
	items := []DrawingItem{
		{
			ID:          "d1",
			Shape:       ShapeStar,
			X:           120,
			Y:           80,
			Width:       140,
			Height:      140,
			FillColor:   "#f59e0b",
			BorderColor: "#b45309",
			Background:  "transparent",
			BehindText:  false,
		},
		{
			ID:          "d2",
			Shape:       ShapeArrow,
			X:           200,
			Y:           260,
			Width:       160,
			Height:      90,
			FillColor:   "#10b981",
			BorderColor: "#047857",
			Background:  "#ecfdf5",
			BehindText:  true,
		},
	}

	baseHTML := "<html><head></head><body><p>Hello</p></body></html>"
	saved := InjectHTMLDrawings(baseHTML, items)

	if !strings.Contains(saved, `class="wp-drawing"`) {
		t.Fatalf("saved HTML missing wp-drawing divs:\n%s", saved)
	}
	if !strings.Contains(saved, "z-index: 10") || !strings.Contains(saved, "z-index: 0") {
		t.Errorf("saved HTML missing expected z-index for front/back layers:\n%s", saved)
	}

	extracted := ExtractHTMLDrawings(saved)
	if len(extracted) != len(items) {
		t.Fatalf("ExtractHTMLDrawings len = %d, want %d", len(extracted), len(items))
	}
	for i := range items {
		if extracted[i] != items[i] {
			t.Errorf("item[%d] = %+v, want %+v", i, extracted[i], items[i])
		}
	}

	stripped := StripHTMLDrawings(saved)
	if strings.Contains(stripped, `class="wp-drawing"`) || strings.Contains(stripped, "<svg") {
		t.Errorf("StripHTMLDrawings did not remove SVG divs:\n%s", stripped)
	}
	if !strings.Contains(stripped, "<p>Hello</p>") {
		t.Errorf("StripHTMLDrawings damaged document body:\n%s", stripped)
	}
}
