package editor

import (
	"strings"
	"testing"
)

func TestRenderDropCapSVGAllStyles(t *testing.T) {
	for _, opt := range AvailableDropCapStyles() {
		item := DropCapItem{
			ID:        "dc-1",
			Letter:    "c",
			Style:     opt.Style,
			TextColor: "#7c2d12",
			Width:     110,
			Height:    120,
		}
		svg := RenderDropCapSVG(item)
		if !strings.HasPrefix(svg, "<svg ") || !strings.HasSuffix(svg, "</svg>") {
			t.Errorf("style %q invalid SVG: %s", opt.Style, svg)
		}
		if !strings.Contains(svg, ">C</text>") {
			t.Errorf("style %q did not uppercase letter C: %s", opt.Style, svg)
		}
		if !strings.Contains(svg, `fill="#7c2d12"`) {
			t.Errorf("style %q missing custom text color: %s", opt.Style, svg)
		}
	}
}

func TestInjectAndPrepareHTMLDropCaps(t *testing.T) {
	items := []DropCapItem{
		{
			ID:        "dc-1",
			Letter:    "O",
			Style:     DropCapOrnamental,
			TextColor: "#1e3a8a",
			Width:     110,
			Height:    120,
		},
	}

	editorHTML := `<html><head></head><body><table style="float: left;"><tr><td><img src="dropcap://dc-1" width="110" height="120" /></td></tr></table><p>nce upon a time...</p></body></html>`
	saved := InjectHTMLDropCaps(editorHTML, items)

	if !strings.Contains(saved, `data:image/svg+xml;base64,`) || !strings.Contains(saved, `data-dropcap-id="dc-1"`) {
		t.Fatalf("saved HTML missing browser SVG data URI:\n%s", saved)
	}

	restored, gotItems := PrepareHTMLDropCapsForEditor(saved)
	if len(gotItems) != 1 || gotItems[0] != items[0] {
		t.Fatalf("PrepareHTMLDropCapsForEditor items = %+v, want %+v", gotItems, items)
	}
	if !strings.Contains(restored, `src="dropcap://dc-1"`) {
		t.Errorf("restored HTML did not restore dropcap://dc-1:\n%s", restored)
	}
}
