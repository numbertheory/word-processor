# Word Processor

A cross-platform rich-text editor written in **Go** using **Qt 6** (`github.com/mappu/miqt/qt6`).

## Features

- **Multi-Tab Editing**: Bottom tab bar for editing multiple documents at once; opening an already-open file focuses its existing tab instead of opening a duplicate
- **File Operations & Formats**: New, Open, Save, Save As with unsaved-changes prompts; supports Rich Text (`.html` / `.htm`), Markdown (`.md` / `.markdown`), and Plain Text (`.txt`)
- **Edit & Formatting**: Undo, Redo, Cut, Copy, Paste, Select All, font family & size, Bold, Italic, Underline, Strikethrough, Text color, Highlight color, Clear formatting, Alignment (Left / Center / Right / Justify), and Bulleted / Numbered lists
- **Page Setup**:
  - Configurable Top, Bottom, Left, and Right margins (with subtle vertical margin guide lines in the editor)
  - Centered `10.2 in` page layout in both maximized editor windows and exported HTML viewed in a browser
  - Customizable **Page Background** and **Default Text Color** (overridable by inline HTML formatting)
  - Page setup metadata is saved in `.html` / `.htm` files and discarded for Markdown or plain text
- **SVG Drawing Tools (`Draw...`)**:
  - Insert **Star**, **Circle**, **Rectangle**, **Triangle**, and **Arrow** vector shapes anywhere on the page (independent of margins and text flow)
  - Select to drag-move or resize via 8 handles (`Delete`/`Backspace` to remove, `Esc` to deselect)
  - Right-click context menu to change **Fill Color**, **Border Color**, **Background Color** (or **Transparent Background**), and layer order (**Send to Front (Above Text)** / **Send to Back (Behind Text)**)
  - Saved in `.html` / `.htm` files (discarded for Markdown or plain text) and rendered accurately in browsers
- **Themes & Settings**: Persistent preferences in `~/.config/word-processor/settings.json` with **System (Default)**, **Light**, **Dark**, **Dracula**, **Tokyo Night**, **Nord**, **Gruvbox Dark**, **Catppuccin Mocha**, and **Solarized Light**
- **Status Bar**: Live word count, character count, and cursor line/column

## Building & Running

### CachyOS / Arch Linux

Ensure Go, a C++ compiler, `pkg-config`, and Qt 6 base + SVG packages are installed:

```bash
sudo pacman -S go gcc pkgconf qt6-base qt6-svg
```

Build, run, or install to `~/.local/bin`:

```bash
make run
# or install for the current user:
make install
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
pacman -S mingw-w64-ucrt-x86_64-go mingw-w64-ucrt-x86_64-gcc mingw-w64-ucrt-x86_64-pkgconf mingw-w64-ucrt-x86_64-qt6-base mingw-w64-ucrt-x86_64-qt6-svg
go build -ldflags="-H windowsgui" -o word-processor.exe .
```

## Tests

```bash
make test
```
