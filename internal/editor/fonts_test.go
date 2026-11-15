package editor

import (
	"reflect"
	"testing"
)

func TestAddRecentFont(t *testing.T) {
	var recent []string
	recent = AddRecentFont(recent, "Noto Sans", 3)
	recent = AddRecentFont(recent, "Georgia", 3)
	recent = AddRecentFont(recent, "Courier", 3)
	recent = AddRecentFont(recent, "noto sans", 3) // moves to front without duplicating

	want := []string{"noto sans", "Courier", "Georgia"}
	if !reflect.DeepEqual(recent, want) {
		t.Errorf("got %v, want %v", recent, want)
	}

	recent = AddRecentFont(recent, "JetBrains Mono", 3)
	want = []string{"JetBrains Mono", "noto sans", "Courier"}
	if !reflect.DeepEqual(recent, want) {
		t.Errorf("after cap got %v, want %v", recent, want)
	}
}

func TestBuildFontSections(t *testing.T) {
	installed := []string{
		"Zebra",
		"Arial",
		"Noto Sans",
		"JetBrains Mono",
		"Alpha",
		"Times New Roman",
	}
	recent := []string{"JetBrains Mono", "NonExistentFont"}

	recentOut, classicOut, restOut := BuildFontSections(installed, recent)

	if !reflect.DeepEqual(recentOut, []string{"JetBrains Mono"}) {
		t.Errorf("recentOut = %v, want [JetBrains Mono]", recentOut)
	}
	if !reflect.DeepEqual(classicOut, []string{"Noto Sans", "Times New Roman", "Arial"}) {
		t.Errorf("classicOut = %v, want [Noto Sans, Times New Roman, Arial]", classicOut)
	}
	if !reflect.DeepEqual(restOut, []string{"Alpha", "Zebra"}) {
		t.Errorf("restOut = %v, want [Alpha, Zebra]", restOut)
	}
}
