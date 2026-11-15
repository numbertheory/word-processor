package editor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"math"
	"regexp"
	"strconv"
	"strings"
)

// ShapeType identifies a built-in vector drawing shape.
type ShapeType string

const (
	ShapeStar      ShapeType = "star"
	ShapeCircle    ShapeType = "circle"
	ShapeRectangle ShapeType = "rectangle"
	ShapeTriangle  ShapeType = "triangle"
	ShapeArrow     ShapeType = "arrow"
	ShapeNote      ShapeType = "note"
)

// DrawingItem represents an arbitrarily positioned SVG shape on a document page.
type DrawingItem struct {
	ID          string    `json:"id"`
	Shape       ShapeType `json:"shape"`
	X           float64   `json:"x"`
	Y           float64   `json:"y"`
	Width       float64   `json:"width"`
	Height      float64   `json:"height"`
	Rotation    float64   `json:"rotation,omitempty"`
	Text        string    `json:"text,omitempty"`
	FillColor   string    `json:"fillColor"`
	BorderColor string    `json:"borderColor"`
	Background  string    `json:"background"`
	BehindText  bool      `json:"behindText"`
}

// NewDrawingItem creates a DrawingItem with sensible defaults.
func NewDrawingItem(id string, shape ShapeType, x, y float64) DrawingItem {
	if shape == ShapeNote {
		return NewNoteItem(id, x, y)
	}
	return DrawingItem{
		ID:          id,
		Shape:       shape,
		X:           x,
		Y:           y,
		Width:       120,
		Height:      120,
		FillColor:   "#3b82f6",
		BorderColor: "#1d4ed8",
		Background:  "transparent",
		BehindText:  false,
	}
}

// NewNoteItem creates a fixed-size flat Post-it-style SVG note.
func NewNoteItem(id string, x, y float64) DrawingItem {
	return DrawingItem{
		ID:          id,
		Shape:       ShapeNote,
		X:           x,
		Y:           y,
		Width:       190,
		Height:      190,
		FillColor:   "#fef08a",
		BorderColor: "#eab308",
		Background:  "transparent",
		BehindText:  false,
	}
}

// RenderDrawingSVG generates a standalone SVG string for item.
func RenderDrawingSVG(item DrawingItem) string {
	w := math.Max(item.Width, 1)
	h := math.Max(item.Height, 1)
	strokeW := 2.0
	pad := strokeW + 1.0

	fill := item.FillColor
	if fill == "" {
		fill = "none"
	}
	stroke := item.BorderColor
	if stroke == "" {
		stroke = "none"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.1f %.1f" width="%.1f" height="%.1f" overflow="visible">`, w, h, w, h)

	if item.Background != "" && item.Background != "transparent" {
		fmt.Fprintf(&b, `<rect x="0" y="0" width="%.1f" height="%.1f" fill="%s" />`, w, h, item.Background)
	}

	iw := math.Max(w-2*pad, 1)
	ih := math.Max(h-2*pad, 1)

	if item.Rotation != 0 {
		fmt.Fprintf(&b, `<g transform="rotate(%.1f %.2f %.2f)">`, item.Rotation, w/2, h/2)
	}

	switch item.Shape {
	case ShapeCircle:
		fmt.Fprintf(&b, `<ellipse cx="%.2f" cy="%.2f" rx="%.2f" ry="%.2f" fill="%s" stroke="%s" stroke-width="%.1f" />`,
			w/2, h/2, iw/2, ih/2, fill, stroke, strokeW)
	case ShapeRectangle:
		fmt.Fprintf(&b, `<rect x="%.2f" y="%.2f" width="%.2f" height="%.2f" fill="%s" stroke="%s" stroke-width="%.1f" />`,
			pad, pad, iw, ih, fill, stroke, strokeW)
	case ShapeTriangle:
		p1 := fmt.Sprintf("%.2f,%.2f", w/2, pad)
		p2 := fmt.Sprintf("%.2f,%.2f", w-pad, h-pad)
		p3 := fmt.Sprintf("%.2f,%.2f", pad, h-pad)
		fmt.Fprintf(&b, `<polygon points="%s %s %s" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" />`,
			p1, p2, p3, fill, stroke, strokeW)
	case ShapeArrow:
		shaftTop := pad + ih*0.28
		shaftBottom := pad + ih*0.72
		headX := pad + iw*0.58
		pts := []string{
			fmt.Sprintf("%.2f,%.2f", pad, shaftTop),
			fmt.Sprintf("%.2f,%.2f", headX, shaftTop),
			fmt.Sprintf("%.2f,%.2f", headX, pad),
			fmt.Sprintf("%.2f,%.2f", w-pad, h/2),
			fmt.Sprintf("%.2f,%.2f", headX, h-pad),
			fmt.Sprintf("%.2f,%.2f", headX, shaftBottom),
			fmt.Sprintf("%.2f,%.2f", pad, shaftBottom),
		}
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" />`,
			strings.Join(pts, " "), fill, stroke, strokeW)
	case ShapeStar:
		cx, cy := w/2, h/2
		rxOuter, ryOuter := iw/2, ih/2
		rxInner, ryInner := rxOuter*0.40, ryOuter*0.40
		pts := make([]string, 0, 10)
		for i := 0; i < 10; i++ {
			angle := -math.Pi/2 + float64(i)*math.Pi/5
			rx, ry := rxOuter, ryOuter
			if i%2 == 1 {
				rx, ry = rxInner, ryInner
			}
			px := cx + rx*math.Cos(angle)
			py := cy + ry*math.Sin(angle)
			pts = append(pts, fmt.Sprintf("%.2f,%.2f", px, py))
		}
		fmt.Fprintf(&b, `<polygon points="%s" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" />`,
			strings.Join(pts, " "), fill, stroke, strokeW)
	case ShapeNote:
		renderNoteSVG(&b, w, h, pad, fill, stroke, strokeW, item.Text)
	}

	if item.Rotation != 0 {
		b.WriteString(`</g>`)
	}

	b.WriteString(`</svg>`)
	return b.String()
}

func renderNoteSVG(b *strings.Builder, w, h, pad float64, fill, stroke string, strokeW float64, text string) {
	// Flat note card with an obvious dog-ear folded lower-left corner and stepped drop shadow.
	x0, y0 := pad, pad
	x1, y1 := w-pad-6, h-pad-6
	fold := 26.0

	cardPath := fmt.Sprintf(
		"M %.2f,%.2f L %.2f,%.2f L %.2f,%.2f L %.2f,%.2f L %.2f,%.2f Z",
		x0, y0,
		x1, y0,
		x1, y1,
		x0+fold, y1,
		x0, y1-fold,
	)

	// Stepped flat drop shadow (stronger around the lifted lower-left corner).
	fmt.Fprintf(b, `<path d="%s" transform="translate(5,7)" fill="rgba(0,0,0,0.11)" />`, cardPath)
	fmt.Fprintf(b, `<path d="%s" transform="translate(2,4)" fill="rgba(0,0,0,0.16)" />`, cardPath)

	// Main note card body.
	fmt.Fprintf(b, `<path d="%s" fill="%s" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" />`,
		cardPath, fill, stroke, strokeW)

	// Folded lower-left flap (base fill + subtle shade + border).
	flapPath := fmt.Sprintf(
		"M %.2f,%.2f L %.2f,%.2f L %.2f,%.2f Z",
		x0, y1-fold,
		x0+fold, y1-fold,
		x0+fold, y1,
	)
	fmt.Fprintf(b, `<path d="%s" fill="%s" />`, flapPath, fill)
	fmt.Fprintf(b, `<path d="%s" fill="rgba(0,0,0,0.15)" stroke="%s" stroke-width="%.1f" stroke-linejoin="round" />`,
		flapPath, stroke, math.Max(strokeW-0.5, 1.0))

	if strings.TrimSpace(text) == "" {
		return
	}

	lines := wrapNoteText(text, 18, 8)
	textColor := contrastingNoteTextColor(fill)
	textX := x0 + 12
	textY := y0 + 24
	lineH := 17.0

	for i, line := range lines {
		y := textY + float64(i)*lineH
		fmt.Fprintf(b, `<text x="%.2f" y="%.2f" font-family="Sans Serif" font-size="13" fill="%s">%s</text>`,
			textX, y, textColor, stdhtml.EscapeString(line))
	}
}

func wrapNoteText(text string, maxCharsPerLine, maxLines int) []string {
	var out []string
	paragraphs := strings.Split(strings.ReplaceAll(text, "\r\n", "\n"), "\n")
	for _, p := range paragraphs {
		if len(out) >= maxLines {
			break
		}
		words := strings.Fields(p)
		if len(words) == 0 {
			out = append(out, "")
			continue
		}
		cur := ""
		for _, w := range words {
			for len([]rune(w)) > maxCharsPerLine {
				if cur != "" {
					out = append(out, cur)
					cur = ""
					if len(out) >= maxLines {
						return out
					}
				}
				r := []rune(w)
				out = append(out, string(r[:maxCharsPerLine]))
				w = string(r[maxCharsPerLine:])
				if len(out) >= maxLines {
					return out
				}
			}
			if cur == "" {
				cur = w
			} else if len([]rune(cur))+1+len([]rune(w)) <= maxCharsPerLine {
				cur += " " + w
			} else {
				out = append(out, cur)
				cur = w
				if len(out) >= maxLines {
					return out
				}
			}
		}
		if cur != "" && len(out) < maxLines {
			out = append(out, cur)
		}
	}
	return out
}

func contrastingNoteTextColor(hex string) string {
	if !IsValidHexColor(hex) {
		return "#1e1e1e"
	}
	r, err1 := strconv.ParseInt(hex[1:3], 16, 64)
	g, err2 := strconv.ParseInt(hex[3:5], 16, 64)
	b, err3 := strconv.ParseInt(hex[5:7], 16, 64)
	if err1 != nil || err2 != nil || err3 != nil {
		return "#1e1e1e"
	}
	luma := 0.299*float64(r) + 0.587*float64(g) + 0.114*float64(b)
	if luma < 140 {
		return "#f8fafc"
	}
	return "#1e1e1e"
}

var drawingsMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-drawings["']\s+content=["']([^"']*)["']\s*/?>\s*`)
var drawingsLayerRe = regexp.MustCompile(`(?is)<div\s+id=["']word-processor-drawings-layer["'][^>]*>.*?</div>\s*</div>\s*|<div\s+class=["']wp-drawing["'][^>]*>.*?</div>\s*`)
var bodyOpenRe = regexp.MustCompile(`(?i)<body[^>]*>`)
var bodyCloseRe = regexp.MustCompile(`(?i)</body>`)

// InjectHTMLDrawings embeds drawing metadata in <head> and positioned SVG elements in <body>.
func InjectHTMLDrawings(html string, items []DrawingItem) string {
	cleaned := StripHTMLDrawings(html)
	cleaned = drawingsMetaRe.ReplaceAllString(cleaned, "")

	if len(items) == 0 {
		return cleaned
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return cleaned
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	metaTag := fmt.Sprintf(`<meta name="word-processor-drawings" content="%s" />`, encoded)

	if loc := headOpenRe.FindStringIndex(cleaned); loc != nil {
		cleaned = cleaned[:loc[1]] + "\n" + metaTag + cleaned[loc[1]:]
	} else {
		cleaned = fmt.Sprintf("<head>\n%s\n</head>\n%s", metaTag, cleaned)
	}

	// Ensure body is positioned relatively and the Qt root table sits at z-index 1.
	extraCSS := `<style id="word-processor-drawing-style">body { position: relative; } table[style*="-qt-table-type: root"] { position: relative; z-index: 1; } .wp-drawing, .wp-drawing svg { overflow: visible; }</style>`
	if loc := headOpenRe.FindStringIndex(cleaned); loc != nil {
		cleaned = cleaned[:loc[1]] + "\n" + extraCSS + cleaned[loc[1]:]
	}

	var layer strings.Builder
	for _, item := range items {
		z := 10
		if item.BehindText {
			z = 0
		}
		fmt.Fprintf(
			&layer,
			`<div class="wp-drawing" style="position: absolute; left: %.1fpx; top: %.1fpx; width: %.1fpx; height: %.1fpx; z-index: %d; pointer-events: none;">%s</div>`+"\n",
			item.X, item.Y, item.Width, item.Height, z, RenderDrawingSVG(item),
		)
	}

	if loc := bodyCloseRe.FindStringIndex(cleaned); loc != nil {
		return cleaned[:loc[0]] + layer.String() + cleaned[loc[0]:]
	}
	return cleaned + "\n" + layer.String()
}

var drawingStyleRe = regexp.MustCompile(`(?is)<style\s+id=["']word-processor-drawing-style["'][^>]*>.*?</style>\s*`)
var wpDrawingDivRe = regexp.MustCompile(`(?is)<div\s+class=["']wp-drawing["'][^>]*>.*?</svg>\s*</div>\s*`)

// StripHTMLDrawings removes browser-only SVG overlay divs and styles before loading into QTextEdit.
func StripHTMLDrawings(html string) string {
	cleaned := drawingStyleRe.ReplaceAllString(html, "")
	cleaned = wpDrawingDivRe.ReplaceAllString(cleaned, "")
	return cleaned
}

// ExtractHTMLDrawings parses the word-processor-drawings meta tag from HTML.
func ExtractHTMLDrawings(html string) []DrawingItem {
	matches := drawingsMetaRe.FindStringSubmatch(html)
	if len(matches) < 2 {
		return nil
	}
	decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(matches[1]))
	if err != nil {
		return nil
	}
	var items []DrawingItem
	if err := json.Unmarshal(decoded, &items); err != nil {
		return nil
	}
	return items
}
