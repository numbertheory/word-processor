package editor

import (
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

func showPageSetupDialog(parent *qt6.QWidget, current PageMargins) (PageMargins, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Page Setup")
	dlg.SetModal(true)

	layout := qt6.NewQFormLayout(dlg.QWidget)

	topSpin := newMarginSpinBox(dlg.QWidget, current.Top)
	bottomSpin := newMarginSpinBox(dlg.QWidget, current.Bottom)
	leftSpin := newMarginSpinBox(dlg.QWidget, current.Left)
	rightSpin := newMarginSpinBox(dlg.QWidget, current.Right)

	layout.AddRow3("Top Margin:", topSpin.QWidget)
	layout.AddRow3("Bottom Margin:", bottomSpin.QWidget)
	layout.AddRow3("Left Margin:", leftSpin.QWidget)
	layout.AddRow3("Right Margin:", rightSpin.QWidget)

	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	layout.AddRowWithWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		return current, false
	}

	return PageMargins{
		Top:    topSpin.Value(),
		Right:  rightSpin.Value(),
		Bottom: bottomSpin.Value(),
		Left:   leftSpin.Value(),
	}, true
}

func showSettingsDialog(parent *qt6.QWidget, current AppSettings) (AppSettings, bool) {
	dlg := qt6.NewQDialog(parent)
	dlg.SetWindowTitle("Application Settings")
	dlg.SetModal(true)

	layout := qt6.NewQFormLayout(dlg.QWidget)

	themeCombo := qt6.NewQComboBox(dlg.QWidget)
	themeCombo.AddItem("System (Default)")
	themeCombo.AddItem("Light")
	themeCombo.AddItem("Dark")

	switch current.Theme {
	case ThemeLight:
		themeCombo.SetCurrentIndex(1)
	case ThemeDark:
		themeCombo.SetCurrentIndex(2)
	default:
		themeCombo.SetCurrentIndex(0)
	}

	layout.AddRow3("Editor Theme:", themeCombo.QWidget)

	buttons := qt6.NewQDialogButtonBox4(qt6.QDialogButtonBox__Ok | qt6.QDialogButtonBox__Cancel)
	buttons.OnAccepted(dlg.Accept)
	buttons.OnRejected(dlg.Reject)
	layout.AddRowWithWidget(buttons.QWidget)

	if dlg.Exec() != int(qt6.QDialog__Accepted) {
		return current, false
	}

	var mode ThemeMode
	switch themeCombo.CurrentIndex() {
	case 1:
		mode = ThemeLight
	case 2:
		mode = ThemeDark
	default:
		mode = ThemeSystem
	}

	return AppSettings{Theme: mode}, true
}
