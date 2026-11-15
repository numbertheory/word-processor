package editor

import (
	"fmt"
	"math"

	"github.com/mappu/miqt/qt6"
	"github.com/mappu/miqt/qt6/svg"
)

type dragMode int

const (
	dragNone dragMode = iota
	dragMove
	dragResize
)

type resizeHandle int

const (
	handleNone resizeHandle = iota
	handleNW
	handleN
	handleNE
	handleE
	handleSE
	handleS
	handleSW
	handleW
)

const (
	minDrawingSize  = 24.0
	handleSizePx    = 8.0
	handleHitRadius = 7.0
)

func (mw *MainWindow) pageOriginX(tab *documentTab) float64 {
	vpWidth := float64(tab.editor.Viewport().Width())
	maxPagePx := InchesToPx(MaxPageWidthInches)
	if vpWidth > maxPagePx {
		return (vpWidth - maxPagePx) / 2.0
	}
	return 0
}

func (mw *MainWindow) drawingScreenRect(tab *documentTab, item DrawingItem) (float64, float64, float64, float64) {
	x := mw.pageOriginX(tab) + item.X - float64(tab.editor.HorizontalScrollBar().Value())
	y := item.Y - float64(tab.editor.VerticalScrollBar().Value())
	return x, y, item.Width, item.Height
}

func (mw *MainWindow) insertDrawing(shape ShapeType) {
	tab := mw.currentTab()
	if tab == nil {
		return
	}

	vpWidth := float64(tab.editor.Viewport().Width())
	pageW := math.Min(vpWidth, InchesToPx(MaxPageWidthInches))
	y := float64(tab.editor.VerticalScrollBar().Value()) + 80

	mw.nextDrawingSeq++
	id := fmt.Sprintf("drawing-%d", mw.nextDrawingSeq)
	item := NewDrawingItem(id, shape, 0, y)
	item.X = math.Max(20, (pageW-item.Width)/2)
	tab.drawings = append(tab.drawings, item)
	tab.selectedDrawingID = item.ID
	tab.editor.Document().SetModifiedWithBool(true)
	tab.editor.Viewport().Update()
}

func (mw *MainWindow) paintViewportBackground(tab *documentTab) {
	vp := tab.editor.Viewport()
	painter := qt6.NewQPainter2(vp.QPaintDevice)
	defer painter.End()

	bg := mw.resolvedBackgroundColor(tab)
	painter.FillRect6(vp.Rect(), bg)
}

func (mw *MainWindow) resolvedBackgroundColor(tab *documentTab) *qt6.QColor {
	if IsValidHexColor(tab.pageSetup.BackgroundColor) {
		return qt6.NewQColor6(tab.pageSetup.BackgroundColor)
	}
	opt, _ := FindTheme(mw.settings.Theme)
	if opt.Background != "" {
		return qt6.NewQColor6(opt.Background)
	}
	return tab.editor.Palette().Base().Color()
}

func (mw *MainWindow) paintDrawings(tab *documentTab, behindText bool) {
	if len(tab.drawings) == 0 {
		return
	}
	vp := tab.editor.Viewport()
	painter := qt6.NewQPainter2(vp.QPaintDevice)
	defer painter.End()

	for _, item := range tab.drawings {
		if item.BehindText != behindText {
			continue
		}
		x, y, w, h := mw.drawingScreenRect(tab, item)
		renderer := svg.NewQSvgRenderer3([]byte(RenderDrawingSVG(item)))
		bounds := qt6.NewQRectF4(x, y, w, h)
		renderer.Render2(painter, bounds)
		renderer.Delete()
	}
}

func (mw *MainWindow) paintDrawingSelection(tab *documentTab) {
	item, _ := mw.findDrawingByID(tab, tab.selectedDrawingID)
	if item == nil {
		return
	}

	vp := tab.editor.Viewport()
	painter := qt6.NewQPainter2(vp.QPaintDevice)
	defer painter.End()

	x, y, w, h := mw.drawingScreenRect(tab, *item)
	selColor := qt6.NewQColor6("#3b82f6")
	pen := qt6.NewQPen3(selColor)
	pen.SetStyle(qt6.DashLine)
	pen.SetWidth(1)
	painter.SetPenWithPen(pen)
	painter.DrawRect(qt6.NewQRectF4(x, y, w, h))

	if item.Shape == ShapeNote {
		return
	}

	// Draw 8 resize handles
	solidPen := qt6.NewQPen3(selColor)
	solidPen.SetWidth(1)
	painter.SetPenWithPen(solidPen)
	white := qt6.NewQColor6("#ffffff")

	for _, pt := range handlePoints(x, y, w, h) {
		r := qt6.NewQRectF4(pt.x-handleSizePx/2, pt.y-handleSizePx/2, handleSizePx, handleSizePx)
		painter.FillRect4(r, white)
		painter.DrawRect(r)
	}
}

type handlePoint struct {
	handle resizeHandle
	x, y   float64
}

func handlePoints(x, y, w, h float64) []handlePoint {
	mx := x + w/2
	my := y + h/2
	return []handlePoint{
		{handleNW, x, y},
		{handleN, mx, y},
		{handleNE, x + w, y},
		{handleE, x + w, my},
		{handleSE, x + w, y + h},
		{handleS, mx, y + h},
		{handleSW, x, y + h},
		{handleW, x, my},
	}
}

func (mw *MainWindow) findDrawingByID(tab *documentTab, id string) (*DrawingItem, int) {
	if id == "" {
		return nil, -1
	}
	for i := range tab.drawings {
		if tab.drawings[i].ID == id {
			return &tab.drawings[i], i
		}
	}
	return nil, -1
}

func (mw *MainWindow) hitTestHandle(tab *documentTab, px, py float64) resizeHandle {
	item, _ := mw.findDrawingByID(tab, tab.selectedDrawingID)
	if item == nil || item.Shape == ShapeNote {
		return handleNone
	}
	x, y, w, h := mw.drawingScreenRect(tab, *item)
	for _, pt := range handlePoints(x, y, w, h) {
		if math.Abs(px-pt.x) <= handleHitRadius && math.Abs(py-pt.y) <= handleHitRadius {
			return pt.handle
		}
	}
	return handleNone
}

func (mw *MainWindow) hitTestDrawing(tab *documentTab, px, py float64) int {
	// Search front-to-back: first above-text items in reverse order, then behind-text items in reverse order.
	for _, behind := range []bool{false, true} {
		for i := len(tab.drawings) - 1; i >= 0; i-- {
			item := tab.drawings[i]
			if item.BehindText != behind {
				continue
			}
			x, y, w, h := mw.drawingScreenRect(tab, item)
			if px >= x && px <= x+w && py >= y && py <= y+h {
				return i
			}
		}
	}
	return -1
}

func (mw *MainWindow) handleDrawingMousePress(tab *documentTab, e *qt6.QMouseEvent) bool {
	if e.Button() != qt6.LeftButton {
		return false
	}
	pos := e.Pos()
	px, py := float64(pos.X()), float64(pos.Y())

	if h := mw.hitTestHandle(tab, px, py); h != handleNone {
		if item, idx := mw.findDrawingByID(tab, tab.selectedDrawingID); item != nil {
			tab.dragMode = dragResize
			tab.dragHandle = h
			tab.dragIndex = idx
			tab.dragStartX = px
			tab.dragStartY = py
			tab.dragOrig = *item
			return true
		}
	}

	if idx := mw.hitTestDrawing(tab, px, py); idx >= 0 {
		item := tab.drawings[idx]
		tab.selectedDrawingID = item.ID
		tab.dragMode = dragMove
		tab.dragIndex = idx
		tab.dragStartX = px
		tab.dragStartY = py
		tab.dragOrig = item
		tab.editor.Viewport().Update()
		return true
	}

	if tab.selectedDrawingID != "" {
		tab.selectedDrawingID = ""
		tab.editor.Viewport().Update()
	}
	return false
}

func cursorForHandle(h resizeHandle) qt6.CursorShape {
	switch h {
	case handleNW, handleSE:
		return qt6.SizeFDiagCursor
	case handleNE, handleSW:
		return qt6.SizeBDiagCursor
	case handleN, handleS:
		return qt6.SizeVerCursor
	case handleE, handleW:
		return qt6.SizeHorCursor
	default:
		return qt6.IBeamCursor
	}
}

func (mw *MainWindow) updateHoverCursor(tab *documentTab, px, py float64) {
	vp := tab.editor.Viewport()
	if h := mw.hitTestHandle(tab, px, py); h != handleNone {
		vp.SetCursor(qt6.NewQCursor2(cursorForHandle(h)))
		return
	}
	if mw.hitTestDrawing(tab, px, py) >= 0 {
		vp.SetCursor(qt6.NewQCursor2(qt6.SizeAllCursor))
		return
	}
	vp.SetCursor(qt6.NewQCursor2(qt6.IBeamCursor))
}

func (mw *MainWindow) handleDrawingMouseMove(tab *documentTab, e *qt6.QMouseEvent) bool {
	pos := e.Pos()
	px, py := float64(pos.X()), float64(pos.Y())

	if tab.dragMode == dragNone || tab.dragIndex < 0 || tab.dragIndex >= len(tab.drawings) {
		mw.updateHoverCursor(tab, px, py)
		return false
	}
	dx := px - tab.dragStartX
	dy := py - tab.dragStartY

	item := &tab.drawings[tab.dragIndex]
	orig := tab.dragOrig

	if tab.dragMode == dragMove {
		item.X = orig.X + dx
		item.Y = orig.Y + dy
	} else if tab.dragMode == dragResize {
		x, y, w, h := orig.X, orig.Y, orig.Width, orig.Height

		switch tab.dragHandle {
		case handleNW, handleW, handleSW:
			newW := math.Max(minDrawingSize, w-dx)
			x = x + (w - newW)
			w = newW
		case handleNE, handleE, handleSE:
			w = math.Max(minDrawingSize, w+dx)
		}

		switch tab.dragHandle {
		case handleNW, handleN, handleNE:
			newH := math.Max(minDrawingSize, h-dy)
			y = y + (h - newH)
			h = newH
		case handleSW, handleS, handleSE:
			h = math.Max(minDrawingSize, h+dy)
		}

		item.X, item.Y, item.Width, item.Height = x, y, w, h
	}

	tab.editor.Document().SetModifiedWithBool(true)
	tab.editor.Viewport().Update()
	return true
}

func (mw *MainWindow) handleDrawingMouseRelease(tab *documentTab, e *qt6.QMouseEvent) bool {
	if tab.dragMode != dragNone {
		tab.dragMode = dragNone
		tab.dragHandle = handleNone
		tab.dragIndex = -1
		pos := e.Pos()
		mw.updateHoverCursor(tab, float64(pos.X()), float64(pos.Y()))
		return true
	}
	return false
}

func (mw *MainWindow) handleDrawingKeyPress(tab *documentTab, e *qt6.QKeyEvent) bool {
	if tab.selectedDrawingID == "" {
		return false
	}
	key := e.Key()
	if key == int(qt6.Key_Escape) {
		tab.selectedDrawingID = ""
		tab.editor.Viewport().Update()
		return true
	}
	if key == int(qt6.Key_Delete) || key == int(qt6.Key_Backspace) {
		if _, idx := mw.findDrawingByID(tab, tab.selectedDrawingID); idx >= 0 {
			tab.drawings = append(tab.drawings[:idx], tab.drawings[idx+1:]...)
			tab.selectedDrawingID = ""
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
			return true
		}
	}
	return false
}

func (mw *MainWindow) handleDrawingContextMenu(tab *documentTab, e *qt6.QContextMenuEvent) bool {
	pos := e.Pos()
	idx := mw.hitTestDrawing(tab, float64(pos.X()), float64(pos.Y()))
	if idx < 0 {
		return false
	}

	tab.selectedDrawingID = tab.drawings[idx].ID
	tab.editor.Viewport().Update()

	if tab.drawings[idx].Shape == ShapeNote {
		updated, deleted, ok := showNoteDialog(mw.win.QWidget, tab.drawings[idx])
		if ok {
			if deleted {
				tab.drawings = append(tab.drawings[:idx], tab.drawings[idx+1:]...)
				tab.selectedDrawingID = ""
			} else {
				wasBehind := tab.drawings[idx].BehindText
				tab.drawings[idx] = updated
				if updated.BehindText != wasBehind {
					item := tab.drawings[idx]
					without := append(tab.drawings[:idx], tab.drawings[idx+1:]...)
					if item.BehindText {
						tab.drawings = append([]DrawingItem{item}, without...)
					} else {
						tab.drawings = append(without, item)
					}
				}
			}
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
		}
		e.Accept()
		return true
	}

	menu := qt6.NewQMenu(tab.editor.QWidget)

	fillAct := qt6.NewQAction5("Fill Color...", menu.QObject)
	fillAct.OnTriggered(func() {
		item := &tab.drawings[idx]
		initial := qt6.NewQColor6(item.FillColor)
		chosen := qt6.QColorDialog_GetColor3(initial, mw.win.QWidget, "Shape Fill Color")
		if chosen.IsValid() {
			item.FillColor = chosen.Name()
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
		}
	})
	menu.AddAction(fillAct)

	borderAct := qt6.NewQAction5("Border Color...", menu.QObject)
	borderAct.OnTriggered(func() {
		item := &tab.drawings[idx]
		initial := qt6.NewQColor6(item.BorderColor)
		chosen := qt6.QColorDialog_GetColor3(initial, mw.win.QWidget, "Shape Border Color")
		if chosen.IsValid() {
			item.BorderColor = chosen.Name()
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
		}
	})
	menu.AddAction(borderAct)

	bgAct := qt6.NewQAction5("Background Color...", menu.QObject)
	bgAct.OnTriggered(func() {
		item := &tab.drawings[idx]
		initial := qt6.NewQColor6("#ffffff")
		if IsValidHexColor(item.Background) {
			initial = qt6.NewQColor6(item.Background)
		}
		chosen := qt6.QColorDialog_GetColor3(initial, mw.win.QWidget, "Shape Background Color")
		if chosen.IsValid() {
			item.Background = chosen.Name()
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
		}
	})
	menu.AddAction(bgAct)

	clearBgAct := qt6.NewQAction5("Transparent Background", menu.QObject)
	clearBgAct.OnTriggered(func() {
		tab.drawings[idx].Background = "transparent"
		tab.editor.Document().SetModifiedWithBool(true)
		tab.editor.Viewport().Update()
	})
	menu.AddAction(clearBgAct)

	rotateAct := qt6.NewQAction5("Rotation...", menu.QObject)
	rotateAct.OnTriggered(func() {
		origDeg := tab.drawings[idx].Rotation
		newDeg, ok := showRotationDialog(mw.win.QWidget, origDeg, func(deg float64) {
			tab.drawings[idx].Rotation = deg
			tab.editor.Viewport().Update()
		})
		if ok && newDeg != origDeg {
			tab.drawings[idx].Rotation = newDeg
			tab.editor.Document().SetModifiedWithBool(true)
			tab.editor.Viewport().Update()
		}
	})
	menu.AddAction(rotateAct)

	menu.AddSeparator()

	frontAct := qt6.NewQAction5("Send to Front (Above Text)", menu.QObject)
	frontAct.OnTriggered(func() {
		item := tab.drawings[idx]
		item.BehindText = false
		tab.drawings = append(append(tab.drawings[:idx], tab.drawings[idx+1:]...), item)
		tab.editor.Document().SetModifiedWithBool(true)
		tab.editor.Viewport().Update()
	})
	menu.AddAction(frontAct)

	backAct := qt6.NewQAction5("Send to Back (Behind Text)", menu.QObject)
	backAct.OnTriggered(func() {
		item := tab.drawings[idx]
		item.BehindText = true
		without := append(tab.drawings[:idx], tab.drawings[idx+1:]...)
		tab.drawings = append([]DrawingItem{item}, without...)
		tab.editor.Document().SetModifiedWithBool(true)
		tab.editor.Viewport().Update()
	})
	menu.AddAction(backAct)

	menu.AddSeparator()

	delAct := qt6.NewQAction5("Delete Shape", menu.QObject)
	delAct.OnTriggered(func() {
		tab.drawings = append(tab.drawings[:idx], tab.drawings[idx+1:]...)
		tab.selectedDrawingID = ""
		tab.editor.Document().SetModifiedWithBool(true)
		tab.editor.Viewport().Update()
	})
	menu.AddAction(delAct)

	menu.ExecWithPos(e.GlobalPos())
	e.Accept()
	return true
}
