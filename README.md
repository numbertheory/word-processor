# Word Processor

A cross-platform rich-text editor written in **Go** using **Qt 6** (`github.com/mappu/miqt/qt6`).

## Features (Starter Set)

- **File operations**: New, Open, Save, Save As with unsaved-changes prompts
- **Formats**: Rich Text (`.html` / `.htm`), Markdown (`.md` / `.markdown`), and Plain Text (`.txt`)
- **Edit**: Undo, Redo, Cut, Copy, Paste, Select All
- **Formatting**: Font family & size, Bold, Italic, Underline, Strikethrough, Text color, Highlight color, Clear formatting
- **Paragraphs**: Align Left / Center / Right / Justify, Bulleted & Numbered lists
- **Status bar**: Live word count, character count, and cursor line/column

## Building & Running

### CachyOS / Arch Linux

Ensure Go, a C++ compiler, `pkg-config`, and Qt 6 base packages are installed:

```bash
sudo pacman -S go gcc pkgconf qt6-base
```

Build and run:

```bash
make run
# or directly:
go build -o word-processor . && ./word-processor
```

### macOS

Install Go and Qt 6 via Homebrew, then ensure `pkg-config` can find Qt 6:

```bash
brew install go qt@6 pkg-config
export PKG_CONFIG_PATH="$(brew --prefix qt@6)/lib/pkgconfig:${PKG_CONFIG_PATH}"
go build -o word-processor .
./word-processor
```

### Windows

Using **MSYS2 (MinGW64 / UCRT64)** with Qt 6 and Go:

```powershell
pacman -S mingw-w64-ucrt-x86_64-go mingw-w64-ucrt-x86_64-gcc mingw-w64-ucrt-x86_64-pkgconf mingw-w64-ucrt-x86_64-qt6-base
go build -ldflags="-H windowsgui" -o word-processor.exe .
```

## Tests

```bash
go test -v ./...
```
