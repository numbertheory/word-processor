package editor

import (
	"github.com/mappu/miqt/qt6"
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
