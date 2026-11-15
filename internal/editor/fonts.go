package editor

import (
	"sort"
	"strings"
)

const maxRecentFonts = 5

var classicFontCandidates = []string{
	"Sans Serif",
	"Noto Sans",
	"Noto Serif",
	"Times New Roman",
	"Liberation Serif",
	"Georgia",
	"Arial",
	"Liberation Sans",
	"Helvetica",
	"Courier New",
	"Courier",
	"Liberation Mono",
	"JetBrains Mono",
	"Comic Sans MS",
	"Comic Sans",
}

// AddRecentFont prepends font to recent (case-insensitively deduplicated) and caps at maxItems.
func AddRecentFont(recent []string, font string, maxItems int) []string {
	font = strings.TrimSpace(font)
	if font == "" || maxItems <= 0 {
		return recent
	}
	out := make([]string, 0, maxItems)
	out = append(out, font)
	for _, r := range recent {
		if len(out) >= maxItems {
			break
		}
		if !strings.EqualFold(r, font) {
			out = append(out, r)
		}
	}
	return out
}

// BuildFontSections groups installed system fonts into three ordered sections:
// recent (preserving MRU order), classic (popular readable fonts), and all remaining fonts (sorted).
func BuildFontSections(installed, recent []string) (recentOut, classicOut, restOut []string) {
	byLower := make(map[string]string, len(installed))
	for _, f := range installed {
		trimmed := strings.TrimSpace(f)
		if trimmed == "" {
			continue
		}
		key := strings.ToLower(trimmed)
		if _, exists := byLower[key]; !exists {
			byLower[key] = trimmed
		}
	}

	seen := make(map[string]bool)

	for _, r := range recent {
		key := strings.ToLower(strings.TrimSpace(r))
		if actual, ok := byLower[key]; ok && !seen[key] {
			seen[key] = true
			recentOut = append(recentOut, actual)
		}
	}

	for _, c := range classicFontCandidates {
		key := strings.ToLower(c)
		if actual, ok := byLower[key]; ok && !seen[key] {
			seen[key] = true
			classicOut = append(classicOut, actual)
		}
	}

	for key, actual := range byLower {
		if !seen[key] {
			restOut = append(restOut, actual)
		}
	}
	sort.Slice(restOut, func(i, j int) bool {
		return strings.ToLower(restOut[i]) < strings.ToLower(restOut[j])
	})

	return recentOut, classicOut, restOut
}
