package editor

import (
	"github.com/mappu/miqt/qt6"
)

// formatControls holds the formatting actions and widgets so their state
// stays synchronized with the editor's current cursor format.
type formatControls struct {
	editor *qt6.QTextEdit

	fontCombo *qt6.QFontComboBox
	sizeSpin  *qt6.QSpinBox

	boldAction      *qt6.QAction
	italicAction    *qt6.QAction
	underlineAction *qt6.QAction
	strikeAction    *qt6.QAction

	textColorAction      *qt6.QAction
	highlightColorAction *qt6.QAction
	clearFormatAction    *qt6.QAction

	alignLeftAction    *qt6.QAction
	alignCenterAction  *qt6.QAction
	alignRightAction   *qt6.QAction
	alignJustifyAction *qt6.QAction

	bulletListAction   *qt6.QAction
	numberedListAction *qt6.QAction

	syncing bool
}

func newFormatControls(parent *qt6.QMainWindow) *formatControls {
	fc := &formatControls{}

	fc.fontCombo = qt6.NewQFontComboBox(parent.QWidget)
	fc.fontCombo.SetToolTip("Font Family")
	fc.fontCombo.SetMaximumWidth(200)
	fc.fontCombo.OnCurrentFontChanged(func(font *qt6.QFont) {
		if fc.syncing {
			return
		}
		fmt := qt6.NewQTextCharFormat()
		fmt.SetFontFamilies([]string{font.Family()})
		fc.mergeCharFormat(fmt)
		fc.editor.SetFocus()
	})

	fc.sizeSpin = qt6.NewQSpinBox(parent.QWidget)
	fc.sizeSpin.SetToolTip("Font Size (pt)")
	fc.sizeSpin.SetRange(6, 144)
	fc.sizeSpin.SetValue(12)
	fc.sizeSpin.SetSuffix(" pt")
	fc.sizeSpin.OnValueChanged(func(size int) {
		if fc.syncing {
			return
		}
		fmt := qt6.NewQTextCharFormat()
		fmt.SetFontPointSize(float64(size))
		fc.mergeCharFormat(fmt)
	})

	fc.boldAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "B", "format-text-bold"),
		"Bold",
		parent.QObject,
	)
	fc.boldAction.SetCheckable(true)
	fc.boldAction.SetShortcut(qt6.QKeySequence_FromString("Ctrl+B"))
	fc.boldAction.SetToolTip("Bold (Ctrl+B)")
	fc.boldAction.OnTriggered(func() {
		fmt := qt6.NewQTextCharFormat()
		if fc.boldAction.IsChecked() {
			fmt.SetFontWeight(int(qt6.QFont__Bold))
		} else {
			fmt.SetFontWeight(int(qt6.QFont__Normal))
		}
		fc.mergeCharFormat(fmt)
	})

	fc.italicAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "I", "format-text-italic"),
		"Italic",
		parent.QObject,
	)
	fc.italicAction.SetCheckable(true)
	fc.italicAction.SetShortcut(qt6.QKeySequence_FromString("Ctrl+I"))
	fc.italicAction.SetToolTip("Italic (Ctrl+I)")
	fc.italicAction.OnTriggered(func() {
		fmt := qt6.NewQTextCharFormat()
		fmt.SetFontItalic(fc.italicAction.IsChecked())
		fc.mergeCharFormat(fmt)
	})

	fc.underlineAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "U", "format-text-underline"),
		"Underline",
		parent.QObject,
	)
	fc.underlineAction.SetCheckable(true)
	fc.underlineAction.SetShortcut(qt6.QKeySequence_FromString("Ctrl+U"))
	fc.underlineAction.SetToolTip("Underline (Ctrl+U)")
	fc.underlineAction.OnTriggered(func() {
		fmt := qt6.NewQTextCharFormat()
		fmt.SetFontUnderline(fc.underlineAction.IsChecked())
		fc.mergeCharFormat(fmt)
	})

	fc.strikeAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "S", "format-text-strikethrough"),
		"Strikethrough",
		parent.QObject,
	)
	fc.strikeAction.SetCheckable(true)
	fc.strikeAction.SetToolTip("Strikethrough")
	fc.strikeAction.OnTriggered(func() {
		fmt := qt6.NewQTextCharFormat()
		fmt.SetFontStrikeOut(fc.strikeAction.IsChecked())
		fc.mergeCharFormat(fmt)
	})

	fc.textColorAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "A", "format-text-color"),
		"Text Color",
		parent.QObject,
	)
	fc.textColorAction.SetToolTip("Text Color")
	fc.textColorAction.OnTriggered(func() {
		initial := fc.editor.TextColor()
		color := qt6.QColorDialog_GetColor3(initial, parent.QWidget, "Text Color")
		if color.IsValid() {
			fc.editor.SetTextColor(color)
		}
	})

	fc.highlightColorAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "H", "format-fill-color", "fill-color"),
		"Highlight Color",
		parent.QObject,
	)
	fc.highlightColorAction.SetToolTip("Highlight Color")
	fc.highlightColorAction.OnTriggered(func() {
		initial := fc.editor.TextBackgroundColor()
		color := qt6.QColorDialog_GetColor3(initial, parent.QWidget, "Highlight Color")
		if color.IsValid() {
			fc.editor.SetTextBackgroundColor(color)
		}
	})

	fc.clearFormatAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "Tx", "edit-clear", "draw-eraser"),
		"Clear Formatting",
		parent.QObject,
	)
	fc.clearFormatAction.SetToolTip("Clear Formatting")
	fc.clearFormatAction.OnTriggered(func() {
		cursor := fc.editor.TextCursor()
		cursor.SetCharFormat(qt6.NewQTextCharFormat())
		fc.editor.SetTextCursor(cursor)
	})

	alignGroup := qt6.NewQActionGroup(parent.QObject)
	alignGroup.SetExclusive(true)

	fc.alignLeftAction = fc.newAlignAction(parent, alignGroup, "Align Left", "Align Left", "L", "format-justify-left", qt6.AlignLeft)
	fc.alignCenterAction = fc.newAlignAction(parent, alignGroup, "Center", "Align Center", "C", "format-justify-center", qt6.AlignCenter)
	fc.alignRightAction = fc.newAlignAction(parent, alignGroup, "Align Right", "Align Right", "R", "format-justify-right", qt6.AlignRight)
	fc.alignJustifyAction = fc.newAlignAction(parent, alignGroup, "Justify", "Justify", "J", "format-justify-fill", qt6.AlignJustify)
	fc.alignLeftAction.SetChecked(true)

	fc.bulletListAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "•", "format-list-unordered"),
		"Bulleted List",
		parent.QObject,
	)
	fc.bulletListAction.SetCheckable(true)
	fc.bulletListAction.SetToolTip("Bulleted List")
	fc.bulletListAction.OnTriggered(func() {
		fc.toggleList(qt6.QTextListFormat__ListDisc)
	})

	fc.numberedListAction = qt6.NewQAction6(
		loadActionIcon(parent.QWidget, "1.", "format-list-ordered"),
		"Numbered List",
		parent.QObject,
	)
	fc.numberedListAction.SetCheckable(true)
	fc.numberedListAction.SetToolTip("Numbered List")
	fc.numberedListAction.OnTriggered(func() {
		fc.toggleList(qt6.QTextListFormat__ListDecimal)
	})

	return fc
}

func (fc *formatControls) newAlignAction(
	parent *qt6.QMainWindow,
	group *qt6.QActionGroup,
	label string,
	tooltip string,
	fallbackGlyph string,
	themeIcon string,
	alignment qt6.AlignmentFlag,
) *qt6.QAction {
	action := qt6.NewQAction6(
		loadActionIcon(parent.QWidget, fallbackGlyph, themeIcon),
		label,
		parent.QObject,
	)
	action.SetCheckable(true)
	action.SetToolTip(tooltip)
	group.AddAction(action)
	action.OnTriggered(func() {
		fc.editor.SetAlignment(alignment)
	})
	return action
}

func (fc *formatControls) mergeCharFormat(fmt *qt6.QTextCharFormat) {
	cursor := fc.editor.TextCursor()
	if !cursor.HasSelection() {
		cursor.Select(qt6.QTextCursor__WordUnderCursor)
	}
	cursor.MergeCharFormat(fmt)
	fc.editor.MergeCurrentCharFormat(fmt)
}

func (fc *formatControls) toggleList(style qt6.QTextListFormat__Style) {
	cursor := fc.editor.TextCursor()
	currentList := cursor.CurrentList()
	if currentList != nil && currentList.Format().Style() == style {
		// Remove the block from its current list by resetting object index.
		blockFmt := qt6.NewQTextBlockFormat()
		blockFmt.SetObjectIndex(-1)
		cursor.SetBlockFormat(blockFmt)
	} else {
		listFmt := qt6.NewQTextListFormat()
		listFmt.SetStyle(style)
		cursor.CreateList(listFmt)
	}
	fc.syncFromCursor()
}

func (fc *formatControls) populateToolbar(tb *qt6.QToolBar) {
	tb.SetToolButtonStyle(qt6.ToolButtonIconOnly)
	tb.SetIconSize(qt6.NewQSize2(18, 18))

	tb.AddWidget(fc.fontCombo.QWidget)
	tb.AddWidget(fc.sizeSpin.QWidget)
	tb.AddSeparator()

	tb.AddAction(fc.boldAction)
	tb.AddAction(fc.italicAction)
	tb.AddAction(fc.underlineAction)
	tb.AddAction(fc.strikeAction)
	tb.AddSeparator()

	tb.AddAction(fc.textColorAction)
	tb.AddAction(fc.highlightColorAction)
	tb.AddAction(fc.clearFormatAction)
	tb.AddSeparator()

	tb.AddAction(fc.alignLeftAction)
	tb.AddAction(fc.alignCenterAction)
	tb.AddAction(fc.alignRightAction)
	tb.AddAction(fc.alignJustifyAction)
	tb.AddSeparator()

	tb.AddAction(fc.bulletListAction)
	tb.AddAction(fc.numberedListAction)
}

func (fc *formatControls) setEditor(editor *qt6.QTextEdit) {
	fc.editor = editor
}

func (fc *formatControls) syncFromFormat(fmt *qt6.QTextCharFormat) {
	fc.syncing = true
	defer func() { fc.syncing = false }()

	fc.boldAction.SetChecked(fmt.FontWeight() >= int(qt6.QFont__Bold))
	fc.italicAction.SetChecked(fmt.FontItalic())
	fc.underlineAction.SetChecked(fmt.FontUnderline())
	fc.strikeAction.SetChecked(fmt.FontStrikeOut())

	if family := fmt.Font().Family(); family != "" {
		fc.fontCombo.SetCurrentFont(qt6.NewQFont2(family))
	}
	if size := int(fmt.FontPointSize()); size > 0 {
		fc.sizeSpin.SetValue(size)
	}
}

func (fc *formatControls) syncFromCursor() {
	fc.syncing = true
	defer func() { fc.syncing = false }()

	align := fc.editor.Alignment()
	switch {
	case align&qt6.AlignCenter != 0:
		fc.alignCenterAction.SetChecked(true)
	case align&qt6.AlignRight != 0:
		fc.alignRightAction.SetChecked(true)
	case align&qt6.AlignJustify != 0:
		fc.alignJustifyAction.SetChecked(true)
	default:
		fc.alignLeftAction.SetChecked(true)
	}

	cursor := fc.editor.TextCursor()
	currentList := cursor.CurrentList()
	if currentList == nil {
		fc.bulletListAction.SetChecked(false)
		fc.numberedListAction.SetChecked(false)
		return
	}
	style := currentList.Format().Style()
	fc.bulletListAction.SetChecked(style == qt6.QTextListFormat__ListDisc)
	fc.numberedListAction.SetChecked(style == qt6.QTextListFormat__ListDecimal)
}
