package editor

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLoadSettingsMissingFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nonexistent", "settings.json")
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings missing file returned error: %v", err)
	}
	if !reflect.DeepEqual(got, DefaultAppSettings()) {
		t.Errorf("got %+v, want %+v", got, DefaultAppSettings())
	}
}

func TestSaveAndLoadSettings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "word-processor", "settings.json")
	want := AppSettings{Theme: ThemeLight, RecentFonts: []string{"Noto Sans", "Georgia"}}

	if err := SaveSettings(path, want); err != nil {
		t.Fatalf("SaveSettings failed: %v", err)
	}

	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings failed: %v", err)
	}
	if !reflect.DeepEqual(got, want) {
		t.Errorf("LoadSettings = %+v, want %+v", got, want)
	}
}

func TestLoadSettingsInvalidThemeFallsBackToSystem(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	if err := os.WriteFile(path, []byte(`{"theme":"sepia"}`), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := LoadSettings(path)
	if err != nil {
		t.Fatalf("LoadSettings returned error: %v", err)
	}
	if got.Theme != ThemeSystem {
		t.Errorf("got Theme=%q, want %q", got.Theme, ThemeSystem)
	}
}

func TestSaveAndLoadAllThemes(t *testing.T) {
	path := filepath.Join(t.TempDir(), "settings.json")
	for _, opt := range AvailableThemes() {
		want := AppSettings{Theme: opt.Mode}
		if err := SaveSettings(path, want); err != nil {
			t.Fatalf("SaveSettings(%q) failed: %v", opt.Mode, err)
		}
		got, err := LoadSettings(path)
		if err != nil {
			t.Fatalf("LoadSettings(%q) failed: %v", opt.Mode, err)
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("theme %q round-trip: got %+v, want %+v", opt.Mode, got, want)
		}
	}
}
