package domain

import "strings"

// SanitizeString trims whitespace, strips control characters and caps the
// length of user-provided text (pseudo, names, alignments, messages...).
func SanitizeString(s string, max int) string {
	var b strings.Builder
	for _, r := range s {
		if r < 32 || r == 127 {
			continue
		}
		b.WriteRune(r)
	}
	out := strings.TrimSpace(b.String())
	runes := []rune(out)
	if len(runes) > max {
		runes = runes[:max]
	}
	return string(runes)
}

// ValidPseudo reports whether s is an acceptable player/DM pseudo.
// Case is preserved; allowed characters are letters, digits, '_' and '-'.
func ValidPseudo(s string) bool {
	if len([]rune(s)) < 2 || len([]rune(s)) > 24 {
		return false
	}
	for _, r := range s {
		switch {
		case r == '_' || r == '-':
		case r >= '0' && r <= '9':
		case r >= 'a' && r <= 'z':
		case r >= 'A' && r <= 'Z':
		case r >= 0xC0 && r <= 0xFF: // latin-1 accented letters
		default:
			return false
		}
	}
	return true
}

// ValidCharName reports whether s is an acceptable character display name.
func ValidCharName(s string) bool {
	n := len([]rune(s))
	return n >= 1 && n <= 40
}