package editor

import (
	"fmt"
	"math"

	"github.com/mappu/miqt/qt6"
)

func newMarginSpinBox(parent *qt6.QWidget, value float64) *qt6.QDoubleSpinBox {
	sb := qt6.NewQDoubleSpinBox(parent)
	sb.SetRange(0.0, 4.0)
	sb.SetSingleStep(0.25)
	sb.SetDecimals(2)
	sb.SetSuffix(" in")
	sb.SetValue(value)
	return sb
}

func updateColorButton(btn *qt6.QPushButton, hexColor string) {
	if IsValidHexColor(hexColor) {
		c := qt6.NewQColor6(hexColor)
		labelColor := "#000000"
		if (0.299*float64(c.Red()) + 0.587*float64(c.Green()) + 0.114*float64(c.Blue())) < 140 {
			labelColor = "#ffffff"
		}
		btn.SetText(hexColor)
		btn.SetStyleSheet(fmt.Sprintf("QPushButton { background-color: %s; color: %s; padding: 4px 10px; }", hexColor, labelColor))
	} else {
		btn.SetText("Theme Default")
		btn.SetStyleSheet("")
	}
}

func newColorPickerRow(dlg *qt6.QDialog, title, defaultHex string, value *string) *qt6.QWidget {
	row := qt6.NewQWidget(dlg.QWidget)
	rowLayout := qt6.NewQHBoxLayout(row)
	rowLayout.SetContentsMargins(0, 0, 0, 0)

	colorBtn := qt6.NewQPushButton(row)
	updateColorButton(colorBtn, *value)
	colorBtn.OnClicked(func() {
		initial := qt6.NewQColor6(defaultHex)
		if IsValidHexColor(*value) {
			initial = qt6.NewQColor6(*value)
		}
		chosen := qt6.QColorDialog_GetColor3(initial, dlg.QWidget, title)
		if chosen.IsValid() {
			*value = chosen.Name()
			updateColorButton(colorBtn, *value)
		}
	})

	resetBtn := qt6.NewQPushButton5("Reset", row)
	resetBtn.OnClicked(func() {
		*value = ""
		updateColorButton(colorBtn, *value)
	})

	rowLayout.AddWidget(colorBtn.QWidget)
	rowLayout.AddWidget(resetBtn.QWidget)
	return row
}

func showPageSetupDialog(parent *qt6.QWidget, current PageSetup) (PageSetup, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Page Setup")
	dlg.SetModal(true)

	layout := qt6.NewQFormLayout(dlg.QWidget)

	topSpin := newMarginSpinBox(dlg.QWidget, current.Margins.Top)
	bottomSpin := newMarginSpinBox(dlg.QWidget, current.Margins.Bottom)
	leftSpin := newMarginSpinBox(dlg.QWidget, current.Margins.Left)
	rightSpin := newMarginSpinBox(dlg.QWidget, current.Margins.Right)

	layout.AddRow3("Top Margin:", topSpin.QWidget)
	layout.AddRow3("Bottom Margin:", bottomSpin.QWidget)
	layout.AddRow3("Left Margin:", leftSpin.QWidget)
	layout.AddRow3("Right Margin:", rightSpin.QWidget)

	bgColor := current.BackgroundColor
	fgColor := current.TextColor
	layout.AddRow3("Page Background:", newColorPickerRow(dlg, "Page Background Color", "#ffffff", &bgColor))
	layout.AddRow3("Default Text Color:", newColorPickerRow(dlg, "Default Text Color", "#1e1e1e", &fgColor))

	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	layout.AddRowWithWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		return current, false
	}

	return PageSetup{
		Margins: PageMargins{
			Top:    topSpin.Value(),
			Right:  rightSpin.Value(),
			Bottom: bottomSpin.Value(),
			Left:   leftSpin.Value(),
		},
		BackgroundColor: bgColor,
		TextColor:       fgColor,
	}, true
}

func showSettingsDialog(parent *qt6.QWidget, current AppSettings) (AppSettings, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Application Settings")
	dlg.SetModal(true)

	layout := qt6.NewQFormLayout(dlg.QWidget)

	themeCombo := qt6.NewQComboBox(dlg.QWidget)
	themes := AvailableThemes()
	selectedIdx := 0
	for i, opt := range themes {
		themeCombo.AddItem(opt.Label)
		if opt.Mode == current.Theme {
			selectedIdx = i
		}
	}
	themeCombo.SetCurrentIndex(selectedIdx)

	layout.AddRow3("Editor Theme:", themeCombo.QWidget)

	spellCheck := qt6.NewQCheckBox3("Enable English Spellcheck")
	spellCheck.SetChecked(current.Spellcheck)
	layout.AddRow3("Spellcheck:", spellCheck.QWidget)

	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	layout.AddRowWithWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		return current, false
	}

	idx := themeCombo.CurrentIndex()
	mode := ThemeSystem
	if idx >= 0 && idx < len(themes) {
		mode = themes[idx].Mode
	}

	return AppSettings{Theme: mode, Spellcheck: spellCheck.IsChecked()}, true
}

func showRotationDialog(parent *qt6.QWidget, currentDeg float64, onPreview func(float64)) (float64, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Rotation")
	dlg.SetModal(true)
	dlg.SetMinimumWidth(320)

	layout := qt6.NewQVBoxLayout(dlg.QWidget)

	row := qt6.NewQWidget(dlg.QWidget)
	rowLayout := qt6.NewQHBoxLayout(row)
	rowLayout.SetContentsMargins(0, 0, 0, 0)

	slider := qt6.NewQSlider4(qt6.Horizontal, row)
	slider.SetRange(-180, 180)
	slider.SetTickPosition(qt6.QSlider__TicksBelow)
	slider.SetTickInterval(45)

	initial := int(math.Round(currentDeg))
	if initial < -180 {
		initial = -180
	} else if initial > 180 {
		initial = 180
	}
	slider.SetValue(initial)

	degLabel := qt6.NewQLabel3(fmt.Sprintf("%d°", initial))
	degLabel.SetMinimumWidth(48)
	degLabel.SetAlignment(qt6.AlignRight | qt6.AlignVCenter)

	slider.OnValueChanged(func(val int) {
		degLabel.SetText(fmt.Sprintf("%d°", val))
		if onPreview != nil {
			onPreview(float64(val))
		}
	})

	rowLayout.AddWidget2(slider.QWidget, 1)
	rowLayout.AddWidget(degLabel.QWidget)
	layout.AddWidget(row)

	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel | qt6.QDialogButtonBox__Reset)
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	if resetBtn := buttons.Button(qt6.QDialogButtonBox__Reset); resetBtn != nil {
		resetBtn.OnClicked(func() {
			slider.SetValue(0)
		})
	}
	layout.AddWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		if onPreview != nil {
			onPreview(currentDeg)
		}
		return currentDeg, false
	}

	return float64(slider.Value()), true
}

func showNoteDialog(parent *qt6.QWidget, current DrawingItem) (DrawingItem, bool, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Edit Note")
	dlg.SetModal(true)
	dlg.SetMinimumWidth(360)

	layout := qt6.NewQFormLayout(dlg.QWidget)

	textEdit := qt6.NewQPlainTextEdit(dlg.QWidget)
	textEdit.SetPlainText(current.Text)
	textEdit.SetMinimumHeight(120)
	layout.AddRow3("Note Text:", textEdit.QWidget)

	bgColor := current.FillColor
	if !IsValidHexColor(bgColor) {
		bgColor = "#fef08a"
	}
	bgBtn := qt6.NewQPushButton(dlg.QWidget)
	updateColorButton(bgBtn, bgColor)
	bgBtn.OnClicked(func() {
		chosen := qt6.QColorDialog_GetColor3(qt6.NewQColor6(bgColor), dlg.QWidget, "Note Background Color")
		if chosen.IsValid() {
			bgColor = chosen.Name()
			updateColorButton(bgBtn, bgColor)
		}
	})
	layout.AddRow3("Background:", bgBtn.QWidget)

	borderColor := current.BorderColor
	if !IsValidHexColor(borderColor) {
		borderColor = "#eab308"
	}
	borderBtn := qt6.NewQPushButton(dlg.QWidget)
	updateColorButton(borderBtn, borderColor)
	borderBtn.OnClicked(func() {
		chosen := qt6.QColorDialog_GetColor3(qt6.NewQColor6(borderColor), dlg.QWidget, "Note Border Color")
		if chosen.IsValid() {
			borderColor = chosen.Name()
			updateColorButton(borderBtn, borderColor)
		}
	})
	layout.AddRow3("Border:", borderBtn.QWidget)

	layerCombo := qt6.NewQComboBox(dlg.QWidget)
	layerCombo.AddItem("Above Text")
	layerCombo.AddItem("Behind Text")
	if current.BehindText {
		layerCombo.SetCurrentIndex(1)
	}
	layout.AddRow3("Layer:", layerCombo.QWidget)

	deleted := false
	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	delBtn := buttons.AddButton2("Delete Note", qt6.QDialogButtonBox__DestructiveRole)
	delBtn.OnClicked(func() {
		deleted = true
		dlg.Accept()
	})
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	layout.AddRowWithWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		return current, false, false
	}
	if deleted {
		return current, true, true
	}

	updated := current
	updated.Text = textEdit.ToPlainText()
	updated.FillColor = bgColor
	updated.BorderColor = borderColor
	updated.BehindText = layerCombo.CurrentIndex() == 1
	return updated, false, true
}
