package editor

import (
	"fmt"
	"unsafe"

	"github.com/mappu/miqt/qt6"
)

const (
	appTitle       = "Word Processor"
	AppVersion     = "0.1.0"
	fileFilter     = "Rich Text (*.html *.htm);;Markdown (*.md *.markdown);;Plain Text (*.txt);;All Files (*)"
	defaultOpenDir = ""
)

type documentTab struct {
	editor            *qt6.QTextEdit
	filePath          string
	pageSetup         PageSetup
	drawings          []DrawingItem
	selectedDrawingID string

	dragMode   dragMode
	dragHandle resizeHandle
	dragIndex  int
	dragStartX float64
	dragStartY float64
	dragOrig   DrawingItem
}

// MainWindow wraps the primary Qt6 application window and tabbed rich-text editor.
type MainWindow struct {
	win         *qt6.QMainWindow
	tabs        *qt6.QTabWidget
	tabByWidget map[unsafe.Pointer]*documentTab
	format      *formatControls
	statsLabel  *qt6.QLabel
	cursorLabel *qt6.QLabel

	undoAct *qt6.QAction
	redoAct *qt6.QAction
	cutAct  *qt6.QAction
	copyAct *qt6.QAction

	settings       AppSettings
	settingsPath   string
	nextDrawingSeq int
}

// NewMainWindow constructs and wires the tabbed rich-text editor window.
func NewMainWindow() *MainWindow {
	win := qt6.NewQMainWindow2()
	win.Resize(1024, 720)

	tabs := qt6.NewQTabWidget(win.QWidget)
	tabs.SetTabPosition(qt6.QTabWidget__South)
	tabs.SetTabsClosable(true)
	tabs.SetMovable(true)
	tabs.SetDocumentMode(true)
	win.SetCentralWidget(tabs.QWidget)

	settingsPath, _ := DefaultSettingsPath()
	settings := DefaultAppSettings()
	if settingsPath != "" {
		if loaded, err := LoadSettings(settingsPath); err == nil {
			settings = loaded
		}
	}

	mw := &MainWindow{
		win:          win,
		tabs:         tabs,
		tabByWidget:  make(map[unsafe.Pointer]*documentTab),
		format:       newFormatControls(win),
		settings:     settings,
		settingsPath: settingsPath,
	}

	mw.setupToolbars()
	mw.setupStatusBar()
	mw.connectWindowSignals()
	mw.newFile()

	return mw
}

// Show displays the main window.
func (mw *MainWindow) Show() {
	mw.win.Show()
}

func (mw *MainWindow) setupToolbars() {
	mainTB := mw.win.AddToolBarWithTitle("File & Edit")
	mainTB.SetMovable(false)
	mainTB.SetToolButtonStyle(qt6.ToolButtonIconOnly)
	mainTB.SetIconSize(qt6.NewQSize2(18, 18))

	// File actions
	newAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "+", "document-new"),
		"New",
		mw.win.QObject,
	)
	newAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__New))
	newAct.SetToolTip("New Tab (Ctrl+N)")
	newAct.OnTriggered(mw.newFile)
	mainTB.AddAction(newAct)

	openAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "O", "document-open"),
		"Open",
		mw.win.QObject,
	)
	openAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Open))
	openAct.SetToolTip("Open Document in New Tab (Ctrl+O)")
	openAct.OnTriggered(mw.openFile)
	mainTB.AddAction(openAct)

	saveAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "S", "document-save"),
		"Save",
		mw.win.QObject,
	)
	saveAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Save))
	saveAct.SetToolTip("Save Document (Ctrl+S)")
	saveAct.OnTriggered(func() { mw.saveCurrentTab() })
	mainTB.AddAction(saveAct)

	saveAsAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "S+", "document-save-as"),
		"Save As",
		mw.win.QObject,
	)
	saveAsAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__SaveAs))
	saveAsAct.SetToolTip("Save As... (Ctrl+Shift+S)")
	saveAsAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			mw.saveTabAs(tab)
		}
	})
	mainTB.AddAction(saveAsAct)

	quitAct := qt6.NewQAction5("Quit", mw.win.QObject)
	quitAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Quit))
	quitAct.OnTriggered(func() { mw.win.Close() })
	mw.win.AddAction(quitAct)

	// Edit actions
	mainTB.AddSeparator()

	mw.undoAct = qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "↶", "edit-undo"),
		"Undo",
		mw.win.QObject,
	)
	mw.undoAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Undo))
	mw.undoAct.SetToolTip("Undo (Ctrl+Z)")
	mw.undoAct.SetEnabled(false)
	mw.undoAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.Undo()
		}
	})
	mainTB.AddAction(mw.undoAct)

	mw.redoAct = qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "↷", "edit-redo"),
		"Redo",
		mw.win.QObject,
	)
	mw.redoAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Redo))
	mw.redoAct.SetToolTip("Redo (Ctrl+Y)")
	mw.redoAct.SetEnabled(false)
	mw.redoAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.Redo()
		}
	})
	mainTB.AddAction(mw.redoAct)

	mainTB.AddSeparator()

	mw.cutAct = qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "✂", "edit-cut"),
		"Cut",
		mw.win.QObject,
	)
	mw.cutAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Cut))
	mw.cutAct.SetToolTip("Cut (Ctrl+X)")
	mw.cutAct.SetEnabled(false)
	mw.cutAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.Cut()
		}
	})
	mainTB.AddAction(mw.cutAct)

	mw.copyAct = qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "C", "edit-copy"),
		"Copy",
		mw.win.QObject,
	)
	mw.copyAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Copy))
	mw.copyAct.SetToolTip("Copy (Ctrl+C)")
	mw.copyAct.SetEnabled(false)
	mw.copyAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.Copy()
		}
	})
	mainTB.AddAction(mw.copyAct)

	pasteAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "P", "edit-paste"),
		"Paste",
		mw.win.QObject,
	)
	pasteAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Paste))
	pasteAct.SetToolTip("Paste (Ctrl+V)")
	pasteAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.Paste()
		}
	})
	mainTB.AddAction(pasteAct)

	selectAllAct := qt6.NewQAction5("Select All", mw.win.QObject)
	selectAllAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__SelectAll))
	selectAllAct.OnTriggered(func() {
		if tab := mw.currentTab(); tab != nil {
			tab.editor.SelectAll()
		}
	})
	mw.win.AddAction(selectAllAct)

	// Page Setup, Application Settings, and Help
	mainTB.AddSeparator()

	pageSetupAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "▤", "document-properties", "configure"),
		"Page Setup",
		mw.win.QObject,
	)
	pageSetupAct.SetToolTip("Page Setup (Margins)")
	pageSetupAct.OnTriggered(mw.openPageSetup)
	mainTB.AddAction(pageSetupAct)

	settingsAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "⚙", "configure", "preferences-other", "preferences-system"),
		"Settings",
		mw.win.QObject,
	)
	settingsAct.SetToolTip("Application Settings")
	settingsAct.OnTriggered(mw.openSettings)
	mainTB.AddAction(settingsAct)

	helpAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "?", "help-about", "help-contents", "dialog-information"),
		"Help",
		mw.win.QObject,
	)
	helpAct.SetToolTip("About " + appTitle)
	helpAct.OnTriggered(func() {
		qt6.QMessageBox_About(
			mw.win.QWidget,
			"About "+appTitle,
			fmt.Sprintf("%s\nVersion %s\n\nA cross-platform rich-text editor built with Go and Qt 6.", appTitle, AppVersion),
		)
	})
	mainTB.AddAction(helpAct)

	// Formatting toolbar
	mw.win.AddToolBarBreak()
	formatTB := mw.win.AddToolBarWithTitle("Formatting")
	formatTB.SetMovable(false)
	mw.format.populateToolbar(formatTB)

	// Draw... dropdown button
	formatTB.AddSeparator()
	drawBtn := qt6.NewQToolButton(formatTB.QWidget)
	drawBtn.SetText("Draw...")
	drawBtn.SetToolTip("Insert SVG Shape")
	drawBtn.SetPopupMode(qt6.QToolButton__InstantPopup)

	drawMenu := qt6.NewQMenu(drawBtn.QWidget)
	shapes := []struct {
		label string
		shape ShapeType
	}{
		{"Star", ShapeStar},
		{"Circle", ShapeCircle},
		{"Rectangle", ShapeRectangle},
		{"Triangle", ShapeTriangle},
		{"Arrow", ShapeArrow},
	}
	for _, s := range shapes {
		shapeType := s.shape
		act := qt6.NewQAction5(s.label, drawMenu.QObject)
		act.OnTriggered(func() {
			mw.insertDrawing(shapeType)
		})
		drawMenu.AddAction(act)
	}
	drawBtn.SetMenu(drawMenu)
	formatTB.AddWidget(drawBtn.QWidget)
}

func (mw *MainWindow) setupStatusBar() {
	sb := mw.win.StatusBar()
	mw.statsLabel = qt6.NewQLabel2()
	mw.cursorLabel = qt6.NewQLabel2()
	sb.AddWidget(mw.statsLabel.QWidget)
	sb.AddPermanentWidget(mw.cursorLabel.QWidget)
}

func (mw *MainWindow) connectWindowSignals() {
	mw.tabs.OnCurrentChanged(func(index int) {
		mw.onActiveTabChanged(index)
	})
	mw.tabs.OnTabCloseRequested(func(index int) {
		mw.closeTab(index)
	})

	mw.win.OnCloseEvent(func(super func(event *qt6.QCloseEvent), event *qt6.QCloseEvent) {
		for i := 0; i < mw.tabs.Count(); i++ {
			tab := mw.tabAt(i)
			if tab != nil && tab.editor.Document().IsModified() {
				mw.tabs.SetCurrentIndex(i)
				if !mw.maybeSaveTab(tab) {
					event.Ignore()
					return
				}
			}
		}
		event.Accept()
	})
}

func (mw *MainWindow) tabAt(index int) *documentTab {
	if index < 0 || index >= mw.tabs.Count() {
		return nil
	}
	w := mw.tabs.Widget(index)
	if w == nil {
		return nil
	}
	return mw.tabByWidget[w.UnsafePointer()]
}

func (mw *MainWindow) currentTab() *documentTab {
	return mw.tabAt(mw.tabs.CurrentIndex())
}

func (mw *MainWindow) findTabByPath(path string) (int, *documentTab) {
	target := CanonicalPath(path)
	if target == "" {
		return -1, nil
	}
	for i := 0; i < mw.tabs.Count(); i++ {
		tab := mw.tabAt(i)
		if tab != nil && tab.filePath != "" && CanonicalPath(tab.filePath) == target {
			return i, tab
		}
	}
	return -1, nil
}

func (mw *MainWindow) createTab(filePath string, loadContent func(*qt6.QTextEdit), pageSetup PageSetup, drawings []DrawingItem) *documentTab {
	ed := qt6.NewQTextEdit(mw.tabs.QWidget)
	ed.SetAcceptRichText(true)
	ed.SetFont(qt6.NewQFont6("Sans Serif", 12))
	ed.Viewport().SetAutoFillBackground(false)

	tab := &documentTab{
		editor:    ed,
		filePath:  CanonicalPath(filePath),
		pageSetup: pageSetup,
		drawings:  drawings,
		dragIndex: -1,
	}
	mw.tabByWidget[ed.QWidget.UnsafePointer()] = tab

	mw.applyThemeToTab(tab)
	if loadContent != nil {
		loadContent(ed)
	}
	mw.applyMarginsToTab(tab)
	ed.Document().SetModifiedWithBool(false)

	ed.OnTextChanged(func() {
		if mw.currentTab() == tab {
			mw.updateStatusBar()
		}
	})
	ed.OnCursorPositionChanged(func() {
		if mw.currentTab() == tab {
			mw.updateStatusBar()
			mw.format.syncFromCursor()
		}
	})
	ed.OnCurrentCharFormatChanged(func(fmt *qt6.QTextCharFormat) {
		if mw.currentTab() == tab {
			mw.format.syncFromFormat(fmt)
		}
	})
	ed.Document().OnModificationChanged(func(changed bool) {
		mw.updateTabTitle(tab)
		if mw.currentTab() == tab {
			mw.updateWindowTitle()
		}
	})
	ed.OnUndoAvailable(func(avail bool) {
		if mw.currentTab() == tab {
			mw.undoAct.SetEnabled(avail)
		}
	})
	ed.OnRedoAvailable(func(avail bool) {
		if mw.currentTab() == tab {
			mw.redoAct.SetEnabled(avail)
		}
	})
	ed.OnCopyAvailable(func(avail bool) {
		if mw.currentTab() == tab {
			mw.cutAct.SetEnabled(avail)
			mw.copyAct.SetEnabled(avail)
		}
	})
	ed.OnPaintEvent(func(super func(e *qt6.QPaintEvent), e *qt6.QPaintEvent) {
		mw.paintViewportBackground(tab)
		mw.paintDrawings(tab, true)
		super(e)
		mw.paintDrawings(tab, false)
		mw.paintMarginGuides(tab)
		mw.paintDrawingSelection(tab)
	})
	ed.OnMousePressEvent(func(super func(e *qt6.QMouseEvent), e *qt6.QMouseEvent) {
		if mw.handleDrawingMousePress(tab, e) {
			return
		}
		super(e)
	})
	ed.OnMouseMoveEvent(func(super func(e *qt6.QMouseEvent), e *qt6.QMouseEvent) {
		if mw.handleDrawingMouseMove(tab, e) {
			return
		}
		super(e)
	})
	ed.OnMouseReleaseEvent(func(super func(e *qt6.QMouseEvent), e *qt6.QMouseEvent) {
		if mw.handleDrawingMouseRelease(tab, e) {
			return
		}
		super(e)
	})
	ed.OnContextMenuEvent(func(super func(e *qt6.QContextMenuEvent), e *qt6.QContextMenuEvent) {
		if mw.handleDrawingContextMenu(tab, e) {
			return
		}
		super(e)
	})
	ed.OnKeyPressEvent(func(super func(e *qt6.QKeyEvent), e *qt6.QKeyEvent) {
		if mw.handleDrawingKeyPress(tab, e) {
			return
		}
		super(e)
	})
	ed.OnResizeEvent(func(super func(e *qt6.QResizeEvent), e *qt6.QResizeEvent) {
		super(e)
		mw.applyMarginsToTab(tab)
	})

	idx := mw.tabs.AddTab(ed.QWidget, DisplayName(tab.filePath))
	if tab.filePath != "" {
		mw.tabs.SetTabToolTip(idx, tab.filePath)
	}
	mw.tabs.SetCurrentIndex(idx)
	ed.SetFocus()
	return tab
}

func (mw *MainWindow) onActiveTabChanged(index int) {
	tab := mw.tabAt(index)
	if tab == nil {
		mw.format.setEditor(nil)
		mw.undoAct.SetEnabled(false)
		mw.redoAct.SetEnabled(false)
		mw.cutAct.SetEnabled(false)
		mw.copyAct.SetEnabled(false)
		mw.updateWindowTitle()
		mw.updateStatusBar()
		return
	}

	mw.format.setEditor(tab.editor)
	mw.format.syncFromFormat(tab.editor.CurrentCharFormat())
	mw.format.syncFromCursor()

	doc := tab.editor.Document()
	mw.undoAct.SetEnabled(doc.IsUndoAvailable())
	mw.redoAct.SetEnabled(doc.IsRedoAvailable())
	hasSel := tab.editor.TextCursor().HasSelection()
	mw.cutAct.SetEnabled(hasSel)
	mw.copyAct.SetEnabled(hasSel)

	mw.updateWindowTitle()
	mw.updateStatusBar()
	tab.editor.SetFocus()
}

func (mw *MainWindow) closeTab(index int) {
	tab := mw.tabAt(index)
	if tab == nil {
		return
	}
	if !mw.maybeSaveTab(tab) {
		return
	}
	delete(mw.tabByWidget, tab.editor.QWidget.UnsafePointer())
	mw.tabs.RemoveTab(index)

	if mw.tabs.Count() == 0 {
		mw.newFile()
	}
}

func (mw *MainWindow) updateTabTitle(tab *documentTab) {
	idx := mw.tabs.IndexOf(tab.editor.QWidget)
	if idx < 0 {
		return
	}
	title := DisplayName(tab.filePath)
	if tab.editor.Document().IsModified() {
		title += "*"
	}
	mw.tabs.SetTabText(idx, title)
	if tab.filePath != "" {
		mw.tabs.SetTabToolTip(idx, tab.filePath)
	}
}

func (mw *MainWindow) updateWindowTitle() {
	tab := mw.currentTab()
	if tab == nil {
		mw.win.SetWindowTitle(appTitle)
		return
	}
	name := DisplayName(tab.filePath)
	modified := ""
	if tab.editor.Document().IsModified() {
		modified = "*"
	}
	mw.win.SetWindowTitle(fmt.Sprintf("%s%s — %s", name, modified, appTitle))
}

func (mw *MainWindow) updateStatusBar() {
	tab := mw.currentTab()
	if tab == nil {
		mw.statsLabel.SetText("")
		mw.cursorLabel.SetText("")
		return
	}
	stats := ComputeStats(tab.editor.ToPlainText())
	mw.statsLabel.SetText(fmt.Sprintf("Words: %d   Characters: %d", stats.Words, stats.Characters))

	cursor := tab.editor.TextCursor()
	mw.cursorLabel.SetText(fmt.Sprintf("Ln %d, Col %d", cursor.BlockNumber()+1, cursor.PositionInBlock()+1))
}

func (mw *MainWindow) maybeSaveTab(tab *documentTab) bool {
	if tab == nil || !tab.editor.Document().IsModified() {
		return true
	}

	res := qt6.QMessageBox_Warning5(
		mw.win.QWidget,
		appTitle,
		fmt.Sprintf("%q has unsaved changes.\nDo you want to save your changes?", DisplayName(tab.filePath)),
		qt6.QMessageBox__Save|qt6.QMessageBox__Discard|qt6.QMessageBox__Cancel,
	)

	switch res {
	case qt6.QMessageBox__Save:
		return mw.saveTab(tab)
	case qt6.QMessageBox__Discard:
		return true
	default:
		return false
	}
}

func (mw *MainWindow) openPageSetup() {
	tab := mw.currentTab()
	if tab == nil {
		return
	}
	updated, ok := showPageSetupDialog(mw.win.QWidget, tab.pageSetup)
	if !ok {
		return
	}
	if updated != tab.pageSetup {
		tab.pageSetup = updated
		mw.applyMarginsToTab(tab)
		mw.applyThemeToTab(tab)
		tab.editor.Document().SetModifiedWithBool(true)
	}
}

func (mw *MainWindow) openSettings() {
	updated, ok := showSettingsDialog(mw.win.QWidget, mw.settings)
	if !ok {
		return
	}
	mw.settings = updated
	mw.applyEditorThemeAll()
	if mw.settingsPath != "" {
		if err := SaveSettings(mw.settingsPath, mw.settings); err != nil {
			qt6.QMessageBox_Warning(mw.win.QWidget, "Settings Error", fmt.Sprintf("Could not save settings:\n%v", err))
		}
	}
}

func (mw *MainWindow) effectiveHorizontalMargins(tab *documentTab) (float64, float64) {
	left := InchesToPx(tab.pageSetup.Margins.Left)
	right := InchesToPx(tab.pageSetup.Margins.Right)

	vpWidth := float64(tab.editor.Viewport().Width())
	maxPagePx := InchesToPx(MaxPageWidthInches)
	if vpWidth > maxPagePx {
		extra := (vpWidth - maxPagePx) / 2.0
		left += extra
		right += extra
	}
	return left, right
}

func (mw *MainWindow) applyMarginsToTab(tab *documentTab) {
	doc := tab.editor.Document()
	wasModified := doc.IsModified()

	left, right := mw.effectiveHorizontalMargins(tab)
	root := doc.RootFrame()
	frameFmt := root.FrameFormat()
	frameFmt.SetTopMargin(InchesToPx(tab.pageSetup.Margins.Top))
	frameFmt.SetRightMargin(right)
	frameFmt.SetBottomMargin(InchesToPx(tab.pageSetup.Margins.Bottom))
	frameFmt.SetLeftMargin(left)
	root.SetFrameFormat(frameFmt)

	if !wasModified {
		doc.SetModifiedWithBool(false)
	}
	tab.editor.Viewport().Update()
}

func (mw *MainWindow) paintMarginGuides(tab *documentTab) {
	vp := tab.editor.Viewport()
	w := vp.Width()
	h := vp.Height()
	if w <= 0 || h <= 0 {
		return
	}

	leftMargin, rightMargin := mw.effectiveHorizontalMargins(tab)
	scrollX := tab.editor.HorizontalScrollBar().Value()
	leftX := int(leftMargin) - scrollX
	rightX := w - int(rightMargin) - scrollX

	painter := qt6.NewQPainter2(vp.QPaintDevice)
	defer painter.End()

	guideColor := qt6.NewQColor11(128, 128, 128, 90)
	painter.SetPen(guideColor)

	if leftX > 0 && leftX < w {
		painter.DrawLine2(leftX, 0, leftX, h)
	}
	if rightX > 0 && rightX < w && rightX != leftX {
		painter.DrawLine2(rightX, 0, rightX, h)
	}
}

func (mw *MainWindow) applyThemeToTab(tab *documentTab) {
	opt, _ := FindTheme(mw.settings.Theme)
	fg := opt.Foreground
	if IsValidHexColor(tab.pageSetup.TextColor) {
		fg = tab.pageSetup.TextColor
	}

	// Keep QTextEdit background transparent so paintViewportBackground + behind-text SVGs
	// render cleanly underneath the document text.
	if opt.Mode == ThemeSystem {
		if fg != "" {
			tab.editor.SetStyleSheet(fmt.Sprintf("QTextEdit { background-color: transparent; color: %s; }", fg))
		} else {
			tab.editor.SetStyleSheet("QTextEdit { background-color: transparent; }")
		}
	} else {
		tab.editor.SetStyleSheet(fmt.Sprintf(
			"QTextEdit { background-color: transparent; color: %s; selection-background-color: %s; selection-color: %s; }",
			fg, opt.SelectionBg, opt.SelectionFg,
		))
	}
	tab.editor.Viewport().Update()
}

func (mw *MainWindow) applyEditorThemeAll() {
	for i := 0; i < mw.tabs.Count(); i++ {
		if tab := mw.tabAt(i); tab != nil {
			mw.applyThemeToTab(tab)
		}
	}
}

func (mw *MainWindow) newFile() {
	mw.createTab("", nil, DefaultPageSetup(), nil)
}

func (mw *MainWindow) openFile() {
	path := qt6.QFileDialog_GetOpenFileName4(mw.win.QWidget, "Open Document", defaultOpenDir, fileFilter)
	if path == "" {
		return
	}

	if idx, existing := mw.findTabByPath(path); existing != nil {
		mw.tabs.SetCurrentIndex(idx)
		existing.editor.SetFocus()
		return
	}

	content, err := ReadDocumentFile(path)
	if err != nil {
		qt6.QMessageBox_Warning(mw.win.QWidget, "Open Failed", fmt.Sprintf("Could not open %s:\n%v", path, err))
		return
	}

	pageSetup := DefaultPageSetup()
	var drawings []DrawingItem
	format := DetectFormat(path)
	var loader func(*qt6.QTextEdit)

	switch format {
	case FormatMarkdown:
		loader = func(ed *qt6.QTextEdit) { ed.SetMarkdown(content) }
	case FormatPlainText:
		loader = func(ed *qt6.QTextEdit) { ed.SetPlainText(content) }
	default:
		pageSetup = ExtractHTMLPageSetup(content)
		drawings = ExtractHTMLDrawings(content)
		cleanHTML := StripHTMLDrawings(StripHTMLPageStyle(content))
		loader = func(ed *qt6.QTextEdit) { ed.SetHtml(cleanHTML) }
	}

	mw.createTab(path, loader, pageSetup, drawings)
}

func (mw *MainWindow) saveCurrentTab() bool {
	tab := mw.currentTab()
	if tab == nil {
		return false
	}
	return mw.saveTab(tab)
}

func (mw *MainWindow) saveTab(tab *documentTab) bool {
	if tab.filePath == "" {
		return mw.saveTabAs(tab)
	}
	return mw.writeTabToPath(tab, tab.filePath)
}

func (mw *MainWindow) saveTabAs(tab *documentTab) bool {
	path := qt6.QFileDialog_GetSaveFileName4(mw.win.QWidget, "Save Document", defaultOpenDir, fileFilter)
	if path == "" {
		return false
	}
	return mw.writeTabToPath(tab, path)
}

func (mw *MainWindow) writeTabToPath(tab *documentTab, path string) bool {
	var content string
	switch DetectFormat(path) {
	case FormatMarkdown:
		content = tab.editor.ToMarkdown()
	case FormatPlainText:
		content = tab.editor.ToPlainText()
	default:
		htmlWithSetup := InjectHTMLPageSetup(tab.editor.ToHtml(), tab.pageSetup)
		content = InjectHTMLDrawings(htmlWithSetup, tab.drawings)
	}

	if err := WriteDocumentFile(path, content); err != nil {
		qt6.QMessageBox_Warning(mw.win.QWidget, "Save Failed", fmt.Sprintf("Could not save %s:\n%v", path, err))
		return false
	}

	tab.filePath = CanonicalPath(path)
	tab.editor.Document().SetModifiedWithBool(false)
	mw.updateTabTitle(tab)
	if mw.currentTab() == tab {
		mw.updateWindowTitle()
	}
	return true
}
