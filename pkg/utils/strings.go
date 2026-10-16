package utils

import "strings"

// FirstNonEmpty returns the first non-empty trimmed string from candidates, or empty string.
func FirstNonEmpty(candidates ...string) string {
	for _, c := range candidates {
		if trimmed := strings.TrimSpace(c); trimmed != "" {
			return trimmed
		}
	}
	return ""
}

// Truncate ensures a string does not exceed maxLen characters.
// If truncated, ellipsis is appended unless maxLen is too small.
func Truncate(s string, maxLen int) string {
	if maxLen <= 0 {
		return ""
	}
	runes := []rune(s)
	if len(runes) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return string(runes[:maxLen])
	}
	return string(runes[:maxLen-3]) + "..."
}
