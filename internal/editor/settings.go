package editor

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
)

// ThemeMode controls the editor's light/dark appearance.
type ThemeMode string

const (
	ThemeSystem ThemeMode = "system"
	ThemeLight  ThemeMode = "light"
	ThemeDark   ThemeMode = "dark"
)

// AppSettings holds persistent application preferences.
type AppSettings struct {
	Theme ThemeMode `json:"theme"`
}

// DefaultAppSettings returns the default settings (obey system theme).
func DefaultAppSettings() AppSettings {
	return AppSettings{Theme: ThemeSystem}
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

	var s AppSettings
	if err := json.Unmarshal(data, &s); err != nil {
		return defaults, err
	}
	switch s.Theme {
	case ThemeSystem, ThemeLight, ThemeDark:
	default:
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
