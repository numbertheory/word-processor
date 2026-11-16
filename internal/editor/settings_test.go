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
	want := AppSettings{Theme: ThemeLight, Spellcheck: true, ZoomPercent: 150, RecentFonts: []string{"Noto Sans", "Georgia"}}

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
		want := AppSettings{Theme: opt.Mode, Spellcheck: true, ZoomPercent: DefaultZoomPercent}
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

func TestClampZoomPercent(t *testing.T) {
	cases := []struct {
		in   int
		want int
	}{
		{0, DefaultZoomPercent},
		{-20, DefaultZoomPercent},
		{10, MinZoomPercent},
		{50, 50},
		{150, 150},
		{300, 300},
		{500, MaxZoomPercent},
	}
	for _, tc := range cases {
		if got := ClampZoomPercent(tc.in); got != tc.want {
			t.Errorf("ClampZoomPercent(%d) = %d, want %d", tc.in, got, tc.want)
		}
	}
}
