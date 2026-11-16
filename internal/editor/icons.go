package editor

import (
	"strings"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/svg"
)

// loadActionIcon returns the first available freedesktop theme icon from
// themeNames, falling back to a crisp palette-aware glyph icon for platforms
// without an icon theme (macOS, Windows, or minimal Linux environments).
func loadActionIcon(widget *qt6.QWidget, fallbackGlyph string, themeNames ...string) *qt6.QIcon {
	for _, name := range themeNames {
		if icon := qt6.QIcon_FromTheme(name); !icon.IsNull() {
			return icon
		}
	}
	return drawGlyphIcon(widget, fallbackGlyph)
}

func drawGlyphIcon(widget *qt6.QWidget, glyph string) *qt6.QIcon {
	const size = 22
	pixmap := qt6.NewQPixmap2(size, size)
	pixmap.FillWithFillColor(qt6.NewQColor2(qt6.Transparent))

	painter := qt6.NewQPainter2(pixmap.QPaintDevice)
	painter.SetRenderHint(qt6.QPainter__Antialiasing)
	painter.SetRenderHint(qt6.QPainter__TextAntialiasing)

	fg := widget.Palette().WindowText().Color()
	painter.SetPen(fg)

	font := qt6.NewQFont6("Sans Serif", 11)
	font.SetBold(true)
	painter.SetFont(font)

	rect := qt6.NewQRectF4(0, 0, float64(size), float64(size))
	painter.DrawText5(rect, int(qt6.AlignCenter), glyph)
	painter.End()

	return qt6.NewQIcon2(pixmap)
}

func drawSVGIcon(widget *qt6.QWidget, svgTemplate string) *qt6.QIcon {
	fg := widget.Palette().WindowText().Color().Name()
	markup := strings.ReplaceAll(svgTemplate, "currentColor", fg)
	renderer := svg.NewQSvgRenderer3([]byte(markup))

	icon := qt6.NewQIcon()
	for _, scale := range []int{1, 2} {
		const logicalSize = 22
		px := logicalSize * scale
		pixmap := qt6.NewQPixmap2(px, px)
		pixmap.FillWithFillColor(qt6.NewQColor2(qt6.Transparent))
		pixmap.SetDevicePixelRatio(float64(scale))

		painter := qt6.NewQPainter2(pixmap.QPaintDevice)
		painter.SetRenderHint(qt6.QPainter__Antialiasing)
		renderer.Render2(painter, qt6.NewQRectF4(0, 0, logicalSize, logicalSize))
		painter.End()

		icon.AddPixmap(pixmap)
	}
	return icon
}

const (
	// Blank document page with folded corner and '+' for New.
	iconNewDocumentSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
  <path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"/>
  <polyline points="14 2 14 8 20 8"/>
  <line x1="12" y1="12" x2="12" y2="18"/>
  <line x1="9" y1="15" x2="15" y2="15"/>
</svg>`

	// Classic open folder for Open.
	iconOpenFolderSVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
  <path d="M3 7a2 2 0 0 1 2-2h4.2l1.8 2H19a2 2 0 0 1 2 2v1"/>
  <path d="M3 10h16.5a1.5 1.5 0 0 1 1.46 1.85l-1.5 6.5A2 2 0 0 1 17.51 20H5a2 2 0 0 1-2-2V7z"/>
</svg>`

	// Classic 3.5" floppy disk for Save.
	iconSaveFloppySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
  <path d="M19 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v11a2 2 0 0 1-2 2z"/>
  <polyline points="17 21 17 13 7 13 7 21"/>
  <polyline points="7 3 7 8 15 8"/>
</svg>`

	// Classic 3.5" floppy disk with pencil badge for Save As.
	iconSaveAsFloppySVG = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.75" stroke-linecap="round" stroke-linejoin="round">
  <path d="M12 21H5a2 2 0 0 1-2-2V5a2 2 0 0 1 2-2h11l5 5v4"/>
  <polyline points="7 21 7 13 13 13"/>
  <polyline points="7 3 7 8 15 8"/>
  <path d="M16.5 21.5l4.8-4.8a1.2 1.2 0 0 0-1.7-1.7l-4.8 4.8L14 22l2.5-.5z"/>
</svg>`
)
