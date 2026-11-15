package main

import (
	"os"
	"runtime"

	"github.com/mappu/miqt/qt6"

	"word-processor/internal/editor"
)

func main() {
	// Qt GUI operations must run on the main OS thread.
	runtime.LockOSThread()

	qt6.NewQApplication(os.Args)
	qt6.QCoreApplication_SetApplicationName("Word Processor")
	qt6.QCoreApplication_SetApplicationVersion(editor.AppVersion)
	qt6.QCoreApplication_SetOrganizationName("word-processor")

	win := editor.NewMainWindow()
	win.Show()

	os.Exit(qt6.QApplication_Exec())
}
