package editor

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ThemeMode controls the editor's color theme.
type ThemeMode string

const (
	ThemeSystem          ThemeMode = "system"
	ThemeLight           ThemeMode = "light"
	ThemeDark            ThemeMode = "dark"
	ThemeDracula         ThemeMode = "dracula"
	ThemeTokyoNight      ThemeMode = "tokyo-night"
	ThemeNord            ThemeMode = "nord"
	ThemeGruvboxDark     ThemeMode = "gruvbox-dark"
	ThemeCatppuccinMocha ThemeMode = "catppuccin-mocha"
	ThemeSolarizedLight  ThemeMode = "solarized-light"
)

// ThemeOption describes a selectable editor color theme.
type ThemeOption struct {
	Mode        ThemeMode
	Label       string
	Background  string
	Foreground  string
	SelectionBg string
	SelectionFg string
}

var themeOptions = []ThemeOption{
	{Mode: ThemeSystem, Label: "System (Default)"},
	{Mode: ThemeLight, Label: "Light", Background: "#ffffff", Foreground: "#1e1e1e", SelectionBg: "#3b82f6", SelectionFg: "#ffffff"},
	{Mode: ThemeDark, Label: "Dark", Background: "#1e1e1e", Foreground: "#e6e6e6", SelectionBg: "#2563eb", SelectionFg: "#ffffff"},
	{Mode: ThemeDracula, Label: "Dracula", Background: "#282a36", Foreground: "#f8f8f2", SelectionBg: "#44475a", SelectionFg: "#f8f8f2"},
	{Mode: ThemeTokyoNight, Label: "Tokyo Night", Background: "#1a1b26", Foreground: "#c0caf5", SelectionBg: "#33467c", SelectionFg: "#c0caf5"},
	{Mode: ThemeNord, Label: "Nord", Background: "#2e3440", Foreground: "#d8dee9", SelectionBg: "#434c5e", SelectionFg: "#eceff4"},
	{Mode: ThemeGruvboxDark, Label: "Gruvbox Dark", Background: "#282828", Foreground: "#ebdbb2", SelectionBg: "#504945", SelectionFg: "#fbf1c7"},
	{Mode: ThemeCatppuccinMocha, Label: "Catppuccin Mocha", Background: "#1e1e2e", Foreground: "#cdd6f4", SelectionBg: "#45475a", SelectionFg: "#cdd6f4"},
	{Mode: ThemeSolarizedLight, Label: "Solarized Light", Background: "#fdf6e3", Foreground: "#657b83", SelectionBg: "#eee8d5", SelectionFg: "#586e75"},
}

// AvailableThemes returns all supported editor themes in display order.
func AvailableThemes() []ThemeOption {
	out := make([]ThemeOption, len(themeOptions))
	copy(out, themeOptions)
	return out
}

// FindTheme looks up a ThemeOption by ThemeMode, defaulting to ThemeSystem.
func FindTheme(mode ThemeMode) (ThemeOption, bool) {
	for _, opt := range themeOptions {
		if opt.Mode == mode {
			return opt, true
		}
	}
	return themeOptions[0], false
}

// AppSettings holds persistent application preferences.
type AppSettings struct {
	Theme       ThemeMode `json:"theme"`
	Spellcheck  bool      `json:"spellcheck"`
	RecentFonts []string  `json:"recentFonts,omitempty"`
}

// DefaultAppSettings returns the default settings (obey system theme, spellcheck enabled).
func DefaultAppSettings() AppSettings {
	return AppSettings{Theme: ThemeSystem, Spellcheck: true}
}

// DefaultSettingsPath returns the platform-appropriate settings file path
// (~/.config/word-processor/settings.json on Linux).
func DefaultSettingsPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "word-processor", "settings.json"), nil
}

// LoadSettings reads settings from path, returning defaults if the file does not exist.
func LoadSettings(path string) (AppSettings, error) {
	defaults := DefaultAppSettings()
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return defaults, nil
		}
		return defaults, err
	}

	s := DefaultAppSettings()
	if err := json.Unmarshal(data, &s); err != nil {
		return defaults, err
	}
	if _, ok := FindTheme(s.Theme); !ok {
		s.Theme = ThemeSystem
	}
	return s, nil
}

// SaveSettings writes settings to path, creating parent directories if needed.
func SaveSettings(path string, s AppSettings) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0o644)
}
