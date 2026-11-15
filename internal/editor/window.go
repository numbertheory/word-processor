package editor

import (
	"fmt"

	"github.com/mappu/miqt/qt6"
)

const (
	appTitle       = "Word Processor"
	AppVersion     = "0.1.0"
	fileFilter     = "Rich Text (*.html *.htm);;Markdown (*.md *.markdown);;Plain Text (*.txt);;All Files (*)"
	defaultOpenDir = ""
)

// MainWindow wraps the primary Qt6 application window and rich-text editor.
type MainWindow struct {
	win         *qt6.QMainWindow
	editor      *qt6.QTextEdit
	format      *formatControls
	statsLabel  *qt6.QLabel
	cursorLabel *qt6.QLabel

	currentPath string
}

// NewMainWindow constructs and wires the rich-text editor window.
func NewMainWindow() *MainWindow {
	win := qt6.NewQMainWindow2()
	win.Resize(1024, 720)

	editor := qt6.NewQTextEdit(win.QWidget)
	editor.SetAcceptRichText(true)
	defaultFont := qt6.NewQFont6("Sans Serif", 12)
	editor.SetFont(defaultFont)
	win.SetCentralWidget(editor.QWidget)

	mw := &MainWindow{
		win:    win,
		editor: editor,
		format: newFormatControls(win, editor),
	}

	mw.setupToolbars()
	mw.setupStatusBar()
	mw.connectSignals()
	mw.updateWindowTitle()
	mw.updateStatusBar()

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
	newAct.SetToolTip("New Document (Ctrl+N)")
	newAct.OnTriggered(mw.newFile)
	mainTB.AddAction(newAct)

	openAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "O", "document-open"),
		"Open",
		mw.win.QObject,
	)
	openAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Open))
	openAct.SetToolTip("Open Document (Ctrl+O)")
	openAct.OnTriggered(mw.openFile)
	mainTB.AddAction(openAct)

	saveAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "S", "document-save"),
		"Save",
		mw.win.QObject,
	)
	saveAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Save))
	saveAct.SetToolTip("Save Document (Ctrl+S)")
	saveAct.OnTriggered(func() { mw.saveFile() })
	mainTB.AddAction(saveAct)

	saveAsAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "S+", "document-save-as"),
		"Save As",
		mw.win.QObject,
	)
	saveAsAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__SaveAs))
	saveAsAct.SetToolTip("Save As... (Ctrl+Shift+S)")
	saveAsAct.OnTriggered(func() { mw.saveFileAs() })
	mainTB.AddAction(saveAsAct)

	quitAct := qt6.NewQAction5("Quit", mw.win.QObject)
	quitAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Quit))
	quitAct.OnTriggered(func() { mw.win.Close() })
	mw.win.AddAction(quitAct)

	// Edit actions
	mainTB.AddSeparator()

	undoAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "↶", "edit-undo"),
		"Undo",
		mw.win.QObject,
	)
	undoAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Undo))
	undoAct.SetToolTip("Undo (Ctrl+Z)")
	undoAct.SetEnabled(false)
	undoAct.OnTriggered(mw.editor.Undo)
	mw.editor.OnUndoAvailable(undoAct.SetEnabled)
	mainTB.AddAction(undoAct)

	redoAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "↷", "edit-redo"),
		"Redo",
		mw.win.QObject,
	)
	redoAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Redo))
	redoAct.SetToolTip("Redo (Ctrl+Y)")
	redoAct.SetEnabled(false)
	redoAct.OnTriggered(mw.editor.Redo)
	mw.editor.OnRedoAvailable(redoAct.SetEnabled)
	mainTB.AddAction(redoAct)

	mainTB.AddSeparator()

	cutAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "✂", "edit-cut"),
		"Cut",
		mw.win.QObject,
	)
	cutAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Cut))
	cutAct.SetToolTip("Cut (Ctrl+X)")
	cutAct.SetEnabled(false)
	cutAct.OnTriggered(mw.editor.Cut)
	mw.editor.OnCopyAvailable(cutAct.SetEnabled)
	mainTB.AddAction(cutAct)

	copyAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "C", "edit-copy"),
		"Copy",
		mw.win.QObject,
	)
	copyAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Copy))
	copyAct.SetToolTip("Copy (Ctrl+C)")
	copyAct.SetEnabled(false)
	copyAct.OnTriggered(mw.editor.Copy)
	mw.editor.OnCopyAvailable(copyAct.SetEnabled)
	mainTB.AddAction(copyAct)

	pasteAct := qt6.NewQAction6(
		loadActionIcon(mw.win.QWidget, "P", "edit-paste"),
		"Paste",
		mw.win.QObject,
	)
	pasteAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__Paste))
	pasteAct.SetToolTip("Paste (Ctrl+V)")
	pasteAct.OnTriggered(mw.editor.Paste)
	mainTB.AddAction(pasteAct)

	selectAllAct := qt6.NewQAction5("Select All", mw.win.QObject)
	selectAllAct.SetShortcut(qt6.NewQKeySequence6(qt6.QKeySequence__SelectAll))
	selectAllAct.OnTriggered(mw.editor.SelectAll)
	mw.win.AddAction(selectAllAct)

	// Help / About
	mainTB.AddSeparator()

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
}

func (mw *MainWindow) setupStatusBar() {
	sb := mw.win.StatusBar()
	mw.statsLabel = qt6.NewQLabel2()
	mw.cursorLabel = qt6.NewQLabel2()
	sb.AddWidget(mw.statsLabel.QWidget)
	sb.AddPermanentWidget(mw.cursorLabel.QWidget)
}

func (mw *MainWindow) connectSignals() {
	mw.editor.OnTextChanged(mw.updateStatusBar)
	mw.editor.OnCursorPositionChanged(func() {
		mw.updateStatusBar()
		mw.format.syncFromCursor()
	})
	mw.editor.OnCurrentCharFormatChanged(func(fmt *qt6.QTextCharFormat) {
		mw.format.syncFromFormat(fmt)
	})
	mw.editor.Document().OnModificationChanged(func(changed bool) {
		mw.updateWindowTitle()
	})

	mw.win.OnCloseEvent(func(super func(event *qt6.QCloseEvent), event *qt6.QCloseEvent) {
		if mw.maybeSave() {
			event.Accept()
			return
		}
		event.Ignore()
	})
}

func (mw *MainWindow) updateWindowTitle() {
	name := DisplayName(mw.currentPath)
	modified := ""
	if mw.editor.Document().IsModified() {
		modified = "*"
	}
	mw.win.SetWindowTitle(fmt.Sprintf("%s%s — %s", name, modified, appTitle))
}

func (mw *MainWindow) updateStatusBar() {
	stats := ComputeStats(mw.editor.ToPlainText())
	mw.statsLabel.SetText(fmt.Sprintf("Words: %d   Characters: %d", stats.Words, stats.Characters))

	cursor := mw.editor.TextCursor()
	mw.cursorLabel.SetText(fmt.Sprintf("Ln %d, Col %d", cursor.BlockNumber()+1, cursor.PositionInBlock()+1))
}

func (mw *MainWindow) maybeSave() bool {
	if !mw.editor.Document().IsModified() {
		return true
	}

	res := qt6.QMessageBox_Warning5(
		mw.win.QWidget,
		appTitle,
		"The document has unsaved changes.\nDo you want to save your changes?",
		qt6.QMessageBox__Save|qt6.QMessageBox__Discard|qt6.QMessageBox__Cancel,
	)

	switch res {
	case qt6.QMessageBox__Save:
		return mw.saveFile()
	case qt6.QMessageBox__Discard:
		return true
	default:
		return false
	}
}

func (mw *MainWindow) newFile() {
	if !mw.maybeSave() {
		return
	}
	mw.editor.Clear()
	mw.currentPath = ""
	mw.editor.Document().SetModifiedWithBool(false)
	mw.updateWindowTitle()
}

func (mw *MainWindow) openFile() {
	if !mw.maybeSave() {
		return
	}
	path := qt6.QFileDialog_GetOpenFileName4(mw.win.QWidget, "Open Document", defaultOpenDir, fileFilter)
	if path == "" {
		return
	}

	content, err := ReadDocumentFile(path)
	if err != nil {
		qt6.QMessageBox_Warning(mw.win.QWidget, "Open Failed", fmt.Sprintf("Could not open %s:\n%v", path, err))
		return
	}

	switch DetectFormat(path) {
	case FormatMarkdown:
		mw.editor.SetMarkdown(content)
	case FormatPlainText:
		mw.editor.SetPlainText(content)
	default:
		mw.editor.SetHtml(content)
	}

	mw.currentPath = path
	mw.editor.Document().SetModifiedWithBool(false)
	mw.updateWindowTitle()
}

func (mw *MainWindow) saveFile() bool {
	if mw.currentPath == "" {
		return mw.saveFileAs()
	}
	return mw.writeToPath(mw.currentPath)
}

func (mw *MainWindow) saveFileAs() bool {
	path := qt6.QFileDialog_GetSaveFileName4(mw.win.QWidget, "Save Document", defaultOpenDir, fileFilter)
	if path == "" {
		return false
	}
	return mw.writeToPath(path)
}

func (mw *MainWindow) writeToPath(path string) bool {
	var content string
	switch DetectFormat(path) {
	case FormatMarkdown:
		content = mw.editor.ToMarkdown()
	case FormatPlainText:
		content = mw.editor.ToPlainText()
	default:
		content = mw.editor.ToHtml()
	}

	if err := WriteDocumentFile(path, content); err != nil {
		qt6.QMessageBox_Warning(mw.win.QWidget, "Save Failed", fmt.Sprintf("Could not save %s:\n%v", path, err))
		return false
	}

	mw.currentPath = path
	mw.editor.Document().SetModifiedWithBool(false)
	mw.updateWindowTitle()
	return true
}
