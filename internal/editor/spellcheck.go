package editor

import (
	"bufio"
	_ "embed"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"unicode"
	"unicode/utf16"
)

//go:embed data/en_us_fallback.txt
var fallbackEnglishDict string

// WordSpan locates a word inside a document using Qt UTF-16 code unit offsets.
type WordSpan struct {
	Word  string
	Start int
	End   int
}

type affixRule struct {
	strip string
	add   string
	cond  *regexp.Regexp
}

type affixGroup struct {
	isPrefix bool
	cross    bool
	rules    []affixRule
}

// SpellChecker provides English spellchecking, suggestions, and a persistent custom dictionary.
type SpellChecker struct {
	mu         sync.RWMutex
	words      map[string]struct{}
	lowerWords map[string]string
	custom     map[string]struct{}
	ignored    map[string]struct{}
	customPath string
}

// NewSpellChecker creates a SpellChecker loaded with the embedded fallback dictionary,
// and optionally loads a custom user dictionary from customPath.
func NewSpellChecker(customPath string) *SpellChecker {
	sc := &SpellChecker{
		words:      make(map[string]struct{}, 1024),
		lowerWords: make(map[string]string, 1024),
		custom:     make(map[string]struct{}),
		ignored:    make(map[string]struct{}),
		customPath: customPath,
	}
	sc.loadWordList(strings.NewReader(fallbackEnglishDict))
	sc.loadCustomDictionary()
	return sc
}

// DefaultCustomDictionaryPath returns ~/.config/word-processor/dictionary.txt on Linux.
func DefaultCustomDictionaryPath() (string, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(configDir, "word-processor", "dictionary.txt"), nil
}

// LoadSystemEnglishDictionary attempts to load Hunspell en_US from standard system paths.
// Safe to run in a background goroutine.
func (sc *SpellChecker) LoadSystemEnglishDictionary() bool {
	candidates := [][2]string{
		{"/usr/share/hunspell/en_US.dic", "/usr/share/hunspell/en_US.aff"},
		{"/usr/share/myspell/en_US.dic", "/usr/share/myspell/en_US.aff"},
		{"/usr/share/myspell/dicts/en_US.dic", "/usr/share/myspell/dicts/en_US.aff"},
		{"/Library/Spelling/en_US.dic", "/Library/Spelling/en_US.aff"},
	}
	for _, pair := range candidates {
		if _, err := os.Stat(pair[0]); err == nil {
			if sc.loadHunspell(pair[0], pair[1]) == nil {
				return true
			}
		}
	}
	return false
}

func (sc *SpellChecker) addWordLocked(w string) {
	w = normalizeApostrophe(strings.TrimSpace(w))
	if w == "" {
		return
	}
	sc.words[w] = struct{}{}
	lower := strings.ToLower(w)
	if _, exists := sc.lowerWords[lower]; !exists {
		sc.lowerWords[lower] = w
	}
}

func (sc *SpellChecker) loadWordList(r interface{ Read([]byte) (int, error) }) {
	scanner := bufio.NewScanner(r)
	sc.mu.Lock()
	defer sc.mu.Unlock()
	for scanner.Scan() {
		sc.addWordLocked(scanner.Text())
	}
}

func (sc *SpellChecker) loadCustomDictionary() {
	if sc.customPath == "" {
		return
	}
	f, err := os.Open(sc.customPath)
	if err != nil {
		return
	}
	defer f.Close()

	sc.mu.Lock()
	defer sc.mu.Unlock()
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		w := normalizeApostrophe(strings.TrimSpace(scanner.Text()))
		if w != "" {
			sc.custom[strings.ToLower(w)] = struct{}{}
			sc.addWordLocked(w)
		}
	}
}

// AddCustomWord adds word to the in-memory and persistent user dictionary.
func (sc *SpellChecker) AddCustomWord(word string) error {
	word = normalizeApostrophe(strings.TrimSpace(word))
	if word == "" {
		return nil
	}

	sc.mu.Lock()
	sc.custom[strings.ToLower(word)] = struct{}{}
	sc.addWordLocked(word)
	path := sc.customPath
	sc.mu.Unlock()

	if path == "" {
		return nil
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(word + "\n")
	return err
}

// IgnoreWord ignores word for the current session.
func (sc *SpellChecker) IgnoreWord(word string) {
	word = normalizeApostrophe(strings.TrimSpace(word))
	if word == "" {
		return
	}
	sc.mu.Lock()
	defer sc.mu.Unlock()
	sc.ignored[strings.ToLower(word)] = struct{}{}
}

// IsCorrect reports whether word is valid English, in the custom dictionary, or ignored.
func (sc *SpellChecker) IsCorrect(word string) bool {
	word = normalizeApostrophe(strings.TrimSpace(word))
	if shouldSkipWord(word) {
		return true
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	if sc.isKnownLocked(word) {
		return true
	}

	// Possessives like "Alice's" or "dogs'".
	if strings.HasSuffix(word, "'s") || strings.HasSuffix(word, "'S") {
		base := word[:len(word)-2]
		if sc.isKnownLocked(base) {
			return true
		}
	}

	// Hyphenated compounds like "well-known".
	if strings.Contains(word, "-") {
		parts := strings.Split(word, "-")
		allOK := true
		for _, p := range parts {
			if p != "" && !shouldSkipWord(p) && !sc.isKnownLocked(p) {
				allOK = false
				break
			}
		}
		if allOK {
			return true
		}
	}

	return false
}

func (sc *SpellChecker) isKnownLocked(word string) bool {
	lower := strings.ToLower(word)
	if _, ok := sc.ignored[lower]; ok {
		return true
	}
	if _, ok := sc.custom[lower]; ok {
		return true
	}
	if _, ok := sc.words[word]; ok {
		return true
	}
	_, ok := sc.lowerWords[lower]
	return ok
}

// Suggest returns up to maxWords ranked spelling suggestions for word.
func (sc *SpellChecker) Suggest(word string, maxWords int) []string {
	word = normalizeApostrophe(strings.TrimSpace(word))
	if word == "" || maxWords <= 0 {
		return nil
	}

	lower := strings.ToLower(word)
	wasCap := isCapitalized(word)

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	seen := make(map[string]bool)
	var out []string
	addCandidate := func(cand string) bool {
		if seen[cand] {
			return false
		}
		canonical, ok := sc.lowerWords[cand]
		if !ok {
			return false
		}
		seen[cand] = true
		if wasCap {
			canonical = capitalizeFirst(canonical)
		}
		out = append(out, canonical)
		return len(out) >= maxWords
	}

	edits1 := generateEdits1(lower)
	for _, e := range edits1 {
		if addCandidate(e) {
			return out
		}
	}

	// Edit distance 2 (bounded to keep interactive context menus instantaneous).
	limit := len(edits1)
	if limit > 120 {
		limit = 120
	}
	for i := 0; i < limit; i++ {
		for _, e2 := range generateEdits1(edits1[i]) {
			if addCandidate(e2) {
				return out
			}
		}
	}

	return out
}

func generateEdits1(word string) []string {
	const alphabet = "abcdefghijklmnopqrstuvwxyz'"
	n := len(word)
	results := make([]string, 0, n*54+26)

	// Deletes
	for i := 0; i < n; i++ {
		results = append(results, word[:i]+word[i+1:])
	}
	// Transposes
	for i := 0; i < n-1; i++ {
		results = append(results, word[:i]+string(word[i+1])+string(word[i])+word[i+2:])
	}
	// Replaces
	for i := 0; i < n; i++ {
		for j := 0; j < len(alphabet); j++ {
			if alphabet[j] != word[i] {
				results = append(results, word[:i]+string(alphabet[j])+word[i+1:])
			}
		}
	}
	// Inserts
	for i := 0; i <= n; i++ {
		for j := 0; j < len(alphabet); j++ {
			results = append(results, word[:i]+string(alphabet[j])+word[i:])
		}
	}
	return results
}

func shouldSkipWord(word string) bool {
	if len([]rune(word)) <= 1 {
		return true
	}
	if strings.Contains(word, "://") || strings.Contains(word, "@") {
		return true
	}
	hasLetter := false
	allUpper := true
	for _, r := range word {
		if unicode.IsLetter(r) {
			hasLetter = true
			if !unicode.IsUpper(r) {
				allUpper = false
			}
		} else if unicode.IsDigit(r) {
			// Skip words containing digits (e.g. "2026", "1st", "v2")
			return true
		}
	}
	if !hasLetter || allUpper {
		return true
	}
	return false
}

func normalizeApostrophe(s string) string {
	return strings.ReplaceAll(s, "’", "'")
}

func isCapitalized(s string) bool {
	for _, r := range s {
		return unicode.IsUpper(r)
	}
	return false
}

func capitalizeFirst(s string) string {
	for i, r := range s {
		return string(unicode.ToUpper(r)) + s[i+len(string(r)):]
	}
	return s
}

// TokenizeWords splits text into word spans using UTF-16 code unit offsets.
func TokenizeWords(text string) []WordSpan {
	var spans []WordSpan
	utf16Pos := 0
	inWord := false
	wordStart := 0
	var wordBuf strings.Builder

	flush := func() {
		if !inWord {
			return
		}
		w := wordBuf.String()
		// Trim leading/trailing apostrophes or hyphens.
		trimLeft := 0
		for _, r := range w {
			if r == '\'' || r == '’' || r == '-' {
				trimLeft += utf16Len(r)
			} else {
				break
			}
		}
		trimmed := strings.Trim(w, "'’-")
		if trimmed != "" {
			start := wordStart + trimLeft
			end := start + utf16StringLen(trimmed)
			spans = append(spans, WordSpan{Word: trimmed, Start: start, End: end})
		}
		wordBuf.Reset()
		inWord = false
	}

	for _, r := range text {
		rLen := utf16Len(r)
		if isWordRune(r) {
			if !inWord {
				inWord = true
				wordStart = utf16Pos
			}
			wordBuf.WriteRune(r)
		} else {
			flush()
		}
		utf16Pos += rLen
	}
	flush()
	return spans
}

func isWordRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '\'' || r == '’' || r == '-'
}

func utf16Len(r rune) int {
	if r > 0xFFFF {
		return 2
	}
	return 1
}

func utf16StringLen(s string) int {
	return len(utf16.Encode([]rune(s)))
}

func (sc *SpellChecker) loadHunspell(dicPath, affPath string) error {
	affixes, err := parseAffixFile(affPath)
	if err != nil {
		affixes = nil
	}

	f, err := os.Open(dicPath)
	if err != nil {
		return err
	}
	defer f.Close()

	type entry struct {
		word  string
		flags string
	}
	var entries []entry

	scanner := bufio.NewScanner(f)
	first := true
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if first {
			first = false
			if _, err := strconv.Atoi(line); err == nil {
				continue
			}
		}
		if idx := strings.IndexByte(line, '\t'); idx >= 0 {
			line = line[:idx]
		}
		word, flags, hasSlash := strings.Cut(line, "/")
		word = strings.TrimSpace(word)
		if word == "" {
			continue
		}
		if !hasSlash {
			flags = ""
		}
		entries = append(entries, entry{word: word, flags: flags})
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	for _, e := range entries {
		sc.addWordLocked(e.word)
		if e.flags == "" || len(affixes) == 0 {
			continue
		}
		var prefixes, suffixes []string
		for _, flag := range e.flags {
			grp, ok := affixes[flag]
			if !ok {
				continue
			}
			for _, rule := range grp.rules {
				if rule.cond != nil && !rule.cond.MatchString(e.word) {
					continue
				}
				derived := applyAffixRule(e.word, rule, grp.isPrefix)
				if derived == "" {
					continue
				}
				sc.addWordLocked(derived)
				if grp.cross {
					if grp.isPrefix {
						prefixes = append(prefixes, derived)
					} else {
						suffixes = append(suffixes, derived)
					}
				}
			}
		}
		if len(prefixes) > 0 && len(suffixes) > 0 {
			for _, flag := range e.flags {
				grp, ok := affixes[flag]
				if !ok || !grp.cross || grp.isPrefix {
					continue
				}
				for _, prefWord := range prefixes {
					for _, rule := range grp.rules {
						if rule.cond != nil && !rule.cond.MatchString(prefWord) {
							continue
						}
						if derived := applyAffixRule(prefWord, rule, false); derived != "" {
							sc.addWordLocked(derived)
						}
					}
				}
			}
		}
	}
	return nil
}

func applyAffixRule(word string, rule affixRule, isPrefix bool) string {
	if isPrefix {
		if rule.strip != "" {
			if !strings.HasPrefix(word, rule.strip) {
				return ""
			}
			word = word[len(rule.strip):]
		}
		return rule.add + word
	}
	if rule.strip != "" {
		if !strings.HasSuffix(word, rule.strip) {
			return ""
		}
		word = word[:len(word)-len(rule.strip)]
	}
	return word + rule.add
}

func parseAffixFile(path string) (map[rune]*affixGroup, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	groups := make(map[rune]*affixGroup)
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}
		kind := fields[0]
		if kind != "PFX" && kind != "SFX" {
			continue
		}
		flagRunes := []rune(fields[1])
		if len(flagRunes) != 1 {
			continue
		}
		flag := flagRunes[0]
		isPrefix := kind == "PFX"

		if len(fields) == 4 && (fields[2] == "Y" || fields[2] == "N") {
			groups[flag] = &affixGroup{
				isPrefix: isPrefix,
				cross:    fields[2] == "Y",
			}
			continue
		}

		if len(fields) >= 5 {
			grp, ok := groups[flag]
			if !ok {
				continue
			}
			strip := fields[2]
			if strip == "0" {
				strip = ""
			}
			add := fields[3]
			if slash := strings.IndexByte(add, '/'); slash >= 0 {
				add = add[:slash]
			}
			if add == "0" {
				add = ""
			}
			condStr := fields[4]
			var condRe *regexp.Regexp
			if condStr != "." {
				pattern := condStr + "$"
				if isPrefix {
					pattern = "^" + condStr
				}
				condRe, _ = regexp.Compile(pattern)
			}
			grp.rules = append(grp.rules, affixRule{
				strip: strip,
				add:   add,
				cond:  condRe,
			})
		}
	}
	return groups, scanner.Err()
}
