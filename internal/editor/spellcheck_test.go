package editor

import (
	"path/filepath"
	"slices"
	"testing"
)

func TestSpellCheckerFallbackAndSuggestions(t *testing.T) {
	sc := NewSpellChecker("")

	valid := []string{"Hello", "world", "don't", "English", "2026", "HTML", "a"}
	for _, w := range valid {
		if !sc.IsCorrect(w) {
			t.Errorf("expected %q to be marked correct", w)
		}
	}

	if sc.IsCorrect("wrold") {
		t.Errorf("expected misspelled word %q to be marked incorrect", "wrold")
	}

	suggestions := sc.Suggest("wrold", 5)
	if !slices.Contains(suggestions, "world") {
		t.Errorf("Suggest(\"wrold\") = %v, expected \"world\"", suggestions)
	}

	capSuggestions := sc.Suggest("Wrold", 5)
	if !slices.Contains(capSuggestions, "World") {
		t.Errorf("Suggest(\"Wrold\") = %v, expected \"World\"", capSuggestions)
	}
}

func TestSpellCheckerCustomDictionaryAndIgnore(t *testing.T) {
	dictPath := filepath.Join(t.TempDir(), "dictionary.txt")
	sc := NewSpellChecker(dictPath)

	if sc.IsCorrect("CachyOS") {
		t.Fatalf("expected CachyOS to be unknown before adding")
	}

	if err := sc.AddCustomWord("CachyOS"); err != nil {
		t.Fatalf("AddCustomWord failed: %v", err)
	}
	if !sc.IsCorrect("CachyOS") {
		t.Errorf("expected CachyOS to be correct after AddCustomWord")
	}

	// Reload from disk to verify persistence.
	sc2 := NewSpellChecker(dictPath)
	if !sc2.IsCorrect("CachyOS") {
		t.Errorf("expected CachyOS to persist across SpellChecker instances")
	}

	if sc2.IsCorrect("xyzzyword") {
		t.Fatalf("expected xyzzyword to be unknown")
	}
	sc2.IgnoreWord("xyzzyword")
	if !sc2.IsCorrect("xyzzyword") {
		t.Errorf("expected xyzzyword to be ignored after IgnoreWord")
	}
}

func TestTokenizeWordsUTF16Offsets(t *testing.T) {
	// 😀 is 2 UTF-16 code units.
	text := "Hello 😀 wrold!"
	spans := TokenizeWords(text)
	if len(spans) != 2 {
		t.Fatalf("len(spans) = %d, want 2 (%+v)", len(spans), spans)
	}
	if spans[0] != (WordSpan{Word: "Hello", Start: 0, End: 5}) {
		t.Errorf("span[0] = %+v", spans[0])
	}
	// "Hello " (6) + "😀" (2) + " " (1) = 9
	if spans[1] != (WordSpan{Word: "wrold", Start: 9, End: 14}) {
		t.Errorf("span[1] = %+v, want Start:9 End:14", spans[1])
	}
}
