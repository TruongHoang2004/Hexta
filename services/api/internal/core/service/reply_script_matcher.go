package service

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/TruongHoang2004/Hexta/services/api/internal/core/types"
	"golang.org/x/text/unicode/norm"
)

// MatchScript is the deterministic first router step (MVP 5 FR-5.2): it picks the active script whose trigger
// phrase best matches a customer message. It is pure and documented so it can be tested and explained to owners:
//
//  1. Normalize the message and every trigger phrase: Unicode NFC (keyboards and channels send composed or
//     decomposed Vietnamese), lower-case, every character that is not a letter, mark or
//     digit becomes a space, runs of spaces collapse. Vietnamese diacritics are kept ("giá" and "gia" differ).
//  2. A phrase matches when its normalized words appear as a contiguous whole-word run in the message
//     ("giá áo" matches "cho mình hỏi giá áo này", but "giá" does not match "giáp").
//  3. The longest matching phrase (in runes) wins. If phrases of equal best length come from different scripts,
//     the match is ambiguous and no script is chosen; the router then asks the AI service instead.
//
// Whether a model should decide that a script is good enough is an OPEN owner question; that evaluation belongs
// behind the AI service seam (the active scripts are passed to it), never here.
func MatchScript(message string, versions []types.ActiveScriptVersion) (types.ActiveScriptVersion, bool) {
	text := " " + normalizeForMatch(message) + " "
	if strings.TrimSpace(text) == "" {
		return types.ActiveScriptVersion{}, false
	}
	best, bestLen, ambiguous := -1, 0, false
	for i, version := range versions {
		for _, trigger := range version.Content.Triggers {
			phrase := normalizeForMatch(trigger)
			if phrase == "" || !strings.Contains(text, " "+phrase+" ") {
				continue
			}
			length := utf8.RuneCountInString(phrase)
			switch {
			case length > bestLen:
				best, bestLen, ambiguous = i, length, false
			case length == bestLen && versions[best].VersionID != version.VersionID:
				ambiguous = true
			}
		}
	}
	if best < 0 || ambiguous {
		return types.ActiveScriptVersion{}, false
	}
	return versions[best], true
}

func normalizeForMatch(value string) string {
	mapped := strings.Map(func(r rune) rune {
		if unicode.IsLetter(r) || unicode.IsMark(r) || unicode.IsDigit(r) {
			return unicode.ToLower(r)
		}
		return ' '
	}, norm.NFC.String(value))
	return strings.Join(strings.Fields(mapped), " ")
}
