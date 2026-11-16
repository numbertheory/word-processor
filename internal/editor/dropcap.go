package editor

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	stdhtml "html"
	"math"
	"regexp"
	"strings"
)

// DropCapStyle identifies the visual adornment style of a drop cap.
type DropCapStyle string

const (
	DropCapPlain       DropCapStyle = "plain"
	DropCapSimpleBox   DropCapStyle = "simple-box"
	DropCapDoubleBox   DropCapStyle = "double-box"
	DropCapOrnamental  DropCapStyle = "ornamental"
	DropCapFilledBadge DropCapStyle = "filled-badge"
	DropCapFloral      DropCapStyle = "floral"
	DropCapVineScroll  DropCapStyle = "vine-scroll"
	DropCapArtDeco     DropCapStyle = "art-deco"
)

// DropCapStyleOption describes a selectable drop cap style.
type DropCapStyleOption struct {
	Style DropCapStyle
	Label string
}

var dropCapStyles = []DropCapStyleOption{
	{Style: DropCapPlain, Label: "Plain Serif"},
	{Style: DropCapSimpleBox, Label: "Simple Box"},
	{Style: DropCapDoubleBox, Label: "Double Frame"},
	{Style: DropCapOrnamental, Label: "Ornamental Frame"},
	{Style: DropCapFilledBadge, Label: "Filled Badge"},
	{Style: DropCapFloral, Label: "Floral Petals"},
	{Style: DropCapVineScroll, Label: "Vine & Leaves"},
	{Style: DropCapArtDeco, Label: "Art Deco"},
}

// AvailableDropCapStyles returns all drop cap styles in display order.
func AvailableDropCapStyles() []DropCapStyleOption {
	out := make([]DropCapStyleOption, len(dropCapStyles))
	copy(out, dropCapStyles)
	return out
}

// DropCapItem represents a 6-line floating initial letter embedded in document text.
type DropCapItem struct {
	ID        string       `json:"id"`
	Letter    string       `json:"letter"`
	Style     DropCapStyle `json:"style"`
	TextColor string       `json:"textColor"`
	Width     float64      `json:"width"`
	Height    float64      `json:"height"`
}

// NormalizeDropCapLetter returns the first uppercase rune of s, defaulting to "A".
func NormalizeDropCapLetter(s string) string {
	s = strings.TrimSpace(s)
	for _, r := range s {
		return strings.ToUpper(string(r))
	}
	return "A"
}

// RenderDropCapSVG generates a standalone SVG for a drop cap item.
func RenderDropCapSVG(item DropCapItem) string {
	w := math.Max(item.Width, 24)
	h := math.Max(item.Height, 24)
	letter := NormalizeDropCapLetter(item.Letter)

	color := item.TextColor
	if !IsValidHexColor(color) {
		color = "#1e293b"
	}

	var b strings.Builder
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 %.1f %.1f" width="%.1f" height="%.1f">`, w, h, w, h)

	pad := 3.0
	iw := w - 2*pad
	ih := h - 2*pad

	switch item.Style {
	case DropCapSimpleBox:
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="%s" stroke-width="2.5" />`,
			pad, pad, iw, ih, color)

	case DropCapDoubleBox:
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" fill-opacity="0.06" stroke="%s" stroke-width="2.0" />`,
			pad, pad, iw, ih, color, color)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="%s" stroke-width="1.2" />`,
			pad+5, pad+5, iw-10, ih-10, color)

	case DropCapOrnamental:
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" fill-opacity="0.08" stroke="%s" stroke-width="2.0" />`,
			pad+4, pad+4, iw-8, ih-8, color, color)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="%s" stroke-width="1.0" stroke-dasharray="3,2" />`,
			pad+9, pad+9, iw-18, ih-18, color)
		// Corner diamond ornaments
		corners := [][2]float64{
			{pad + 4, pad + 4},
			{w - pad - 4, pad + 4},
			{w - pad - 4, h - pad - 4},
			{pad + 4, h - pad - 4},
		}
		for _, c := range corners {
			d := 4.5
			fmt.Fprintf(&b, `<polygon points="%.1f,%.1f %.1f,%.1f %.1f,%.1f %.1f,%.1f" fill="%s" />`,
				c[0], c[1]-d, c[0]+d, c[1], c[0], c[1]+d, c[0]-d, c[1], color)
		}

	case DropCapFilledBadge:
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" rx="6" ry="6" fill="%s" fill-opacity="0.16" stroke="%s" stroke-width="2.5" />`,
			pad, pad, iw, ih, color, color)

	case DropCapFloral:
		cx, cy := w/2, h/2
		outerR := math.Min(iw, ih) * 0.48
		innerR := outerR * 0.66
		// 12 radiating flower petals around a central medallion
		for i := 0; i < 12; i++ {
			angle := float64(i) * (2 * math.Pi / 12)
			px := cx + (outerR*0.76)*math.Cos(angle)
			py := cy + (outerR*0.76)*math.Sin(angle)
			deg := angle * 180 / math.Pi
			fmt.Fprintf(&b, `<ellipse cx="%.2f" cy="%.2f" rx="%.2f" ry="%.2f" transform="rotate(%.1f %.2f %.2f)" fill="%s" fill-opacity="0.16" stroke="%s" stroke-width="1.2" />`,
				px, py, outerR*0.24, outerR*0.13, deg, px, py, color, color)
		}
		fmt.Fprintf(&b, `<circle cx="%.2f" cy="%.2f" r="%.2f" fill="%s" fill-opacity="0.08" stroke="%s" stroke-width="2.0" />`,
			cx, cy, innerR, color, color)

	case DropCapVineScroll:
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" fill-opacity="0.05" stroke="%s" stroke-width="1.8" />`,
			pad+4, pad+4, iw-8, ih-8, color, color)
		// Botanical leaves at top/bottom corners
		leafOffsets := [][3]float64{
			{pad + 8, pad + 8, -35},
			{w - pad - 8, pad + 8, 35},
			{pad + 8, h - pad - 8, 35},
			{w - pad - 8, h - pad - 8, -35},
		}
		for _, lf := range leafOffsets {
			fmt.Fprintf(&b, `<ellipse cx="%.2f" cy="%.2f" rx="7.5" ry="3.8" transform="rotate(%.1f %.2f %.2f)" fill="%s" fill-opacity="0.28" stroke="%s" stroke-width="1.1" />`,
				lf[0], lf[1], lf[2], lf[0], lf[1], color, color)
		}

	case DropCapArtDeco:
		// Stepped geometric Art Deco frame
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="none" stroke="%s" stroke-width="2.0" />`,
			pad, pad, iw, ih, color)
		fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="%s" fill-opacity="0.08" stroke="%s" stroke-width="1.2" />`,
			pad+6, pad+6, iw-12, ih-12, color, color)
		// Vertical accent bars
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="2.5" />`,
			pad+11, pad+2, pad+11, h-pad-2, color)
		fmt.Fprintf(&b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="%s" stroke-width="2.5" />`,
			w-pad-11, pad+2, w-pad-11, h-pad-2, color)
	}

	fontSize := h * 0.64
	textY := h*0.5 + fontSize*0.35
	if item.Style == DropCapPlain {
		fontSize = h * 0.90
		textY = h * 0.68
	} else if item.Style == DropCapFloral {
		fontSize = h * 0.54
		textY = h*0.5 + fontSize*0.35
	}
	fmt.Fprintf(&b, `<text x="%.2f" y="%.2f" text-anchor="middle" font-family="serif" font-weight="bold" font-size="%.1f" fill="%s">%s</text>`,
		w/2, textY, fontSize, color, stdhtml.EscapeString(letter))

	b.WriteString(`</svg>`)
	return b.String()
}

var dropCapMetaRe = regexp.MustCompile(`(?i)<meta\s+name=["']word-processor-dropcaps["']\s+content=["']([^"']*)["']\s*/?>\s*`)
var dropCapBrowserImgRe = regexp.MustCompile(`(?is)<img([^>]*?)src=["']data:image/svg\+xml;base64,[^"']*["']([^>]*?)data-dropcap-id=["']([^"']+)["']([^>]*)>`)

// InjectHTMLDropCaps embeds drop cap metadata in <head> and replaces dropcap://
// image sources with inline data-URI SVGs so browsers render them cleanly.
func InjectHTMLDropCaps(html string, items []DropCapItem) string {
	cleaned := dropCapMetaRe.ReplaceAllString(html, "")
	if len(items) == 0 {
		return cleaned
	}

	raw, err := json.Marshal(items)
	if err != nil {
		return cleaned
	}
	encoded := base64.StdEncoding.EncodeToString(raw)
	metaTag := fmt.Sprintf(`<meta name="word-processor-dropcaps" content="%s" />`, encoded)

	if loc := headOpenRe.FindStringIndex(cleaned); loc != nil {
		cleaned = cleaned[:loc[1]] + "\n" + metaTag + cleaned[loc[1]:]
	} else {
		cleaned = fmt.Sprintf("<head>\n%s\n</head>\n%s", metaTag, cleaned)
	}

	for _, item := range items {
		svgB64 := base64.StdEncoding.EncodeToString([]byte(RenderDropCapSVG(item)))
		oldSrc := fmt.Sprintf(`src="dropcap://%s"`, item.ID)
		newSrc := fmt.Sprintf(`src="data:image/svg+xml;base64,%s" data-dropcap-id="%s"`, svgB64, item.ID)
		cleaned = strings.ReplaceAll(cleaned, oldSrc, newSrc)
	}

	return cleaned
}

// PrepareHTMLDropCapsForEditor extracts drop cap items from HTML and restores
// dropcap://<id> image sources so QTextDocument can bind its QImage resources.
func PrepareHTMLDropCapsForEditor(html string) (string, []DropCapItem) {
	var items []DropCapItem
	if matches := dropCapMetaRe.FindStringSubmatch(html); len(matches) >= 2 {
		if decoded, err := base64.StdEncoding.DecodeString(strings.TrimSpace(matches[1])); err == nil {
			_ = json.Unmarshal(decoded, &items)
		}
	}

	restored := dropCapBrowserImgRe.ReplaceAllString(html, `<img${1}src="dropcap://${3}"${2}${4}>`)
	return restored, items
}
