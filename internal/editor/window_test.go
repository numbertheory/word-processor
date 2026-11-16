package editor

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/mappu/miqt/qt6"
)

func TestWindowMarginsAndZoomGUI(t *testing.T) {
	if os.Getenv("WP_GUI_TEST") != "1" {
		t.Skip("set WP_GUI_TEST=1 to run GUI window test")
	}
	runtime.LockOSThread()
	qt6.NewQApplication(os.Args)

	mw := NewMainWindow()
	mw.settingsPath = filepath.Join(t.TempDir(), "settings.json")
	mw.setZoomPercent(DefaultZoomPercent)
	mw.Show()

	timer := qt6.NewQTimer2(mw.win.QObject)
	timer.SetSingleShot(true)
	timer.OnTimeout(func() {
		defer qt6.QCoreApplication_Quit()

		tab := mw.currentTab()
		if tab == nil {
			t.Fatal("expected active tab")
		}
		ed := tab.editor
		doc := ed.Document()
		vpW := float64(ed.Viewport().Width())
		if doc.TextWidth() != vpW {
			t.Errorf("startup doc.TextWidth = %.2f, want %.2f (viewport width)", doc.TextWidth(), vpW)
		}
		effL, effR := mw.effectiveHorizontalMargins(tab)
		ff := doc.RootFrame().FrameFormat()
		if ff.LeftMargin() != effL || ff.RightMargin() != effR {
			t.Errorf("startup frame margins = (%.2f, %.2f), want (%.2f, %.2f)",
				ff.LeftMargin(), ff.RightMargin(), effL, effR)
		}
		emptyRect := ed.CursorRect(ed.TextCursor())
		if emptyRect.X() < int(effL)-2 || emptyRect.X() > int(effL)+2 {
			t.Errorf("startup empty cursor X = %d, want near %.2f", emptyRect.X(), effL)
		}
		if emptyRect.Height() < 12 {
			t.Errorf("startup empty cursor height = %d, want >= 12", emptyRect.Height())
		}

		ed.SetHtml("<p>Hello world</p>")
		mw.applyMarginsToTab(tab)
		doc.SetModifiedWithBool(false)
		htmlBefore := doc.ToHtml()
		cursorRectBefore := ed.CursorRect(ed.TextCursor())
		cursorHBefore := cursorRectBefore.Height()

		mw.setZoomPercent(150)
		qt6.QCoreApplication_ProcessEvents()

		if doc.ToHtml() != htmlBefore {
			t.Errorf("zooming changed document HTML")
		}
		if doc.IsModified() {
			t.Errorf("zooming marked unmodified document as modified")
		}
		if got := mw.zoomLabelBtn.Text(); got != "150%" {
			t.Errorf("zoomLabelBtn.Text() = %q, want %q", got, "150%")
		}
		cursorRectAfter := ed.CursorRect(ed.TextCursor())
		cursorHAfter := cursorRectAfter.Height()
		if cursorHAfter <= cursorHBefore {
			t.Errorf("expected cursor height to increase on 150%% zoom: before=%d, after=%d", cursorHBefore, cursorHAfter)
		}
		expectedLeftX := int(effL*1.5 + 0.5)
		if cursorRectAfter.X() < expectedLeftX-2 || cursorRectAfter.X() > expectedLeftX+2 {
			t.Errorf("150%% cursor X = %d, want near scaled margin guide %d", cursorRectAfter.X(), expectedLeftX)
		}
		if doc.TextWidth() != vpW {
			t.Errorf("post-zoom doc.TextWidth = %.2f, want %.2f", doc.TextWidth(), vpW)
		}
	})
	timer.Start(300)
	qt6.QApplication_Exec()
}
