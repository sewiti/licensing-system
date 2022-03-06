package core

import "strings"

func ValidUsername(username string) bool {
	const (
		minLen = 3
		maxLen = 64
	)
	if len(username) < minLen {
		return false
	}
	if len(username) > maxLen {
		return false
	}
	// Allow only [A-Za-z0-9_-]+
	for _, r := range username {
		if r >= 'A' && r <= 'Z' {
			continue
		}
		if r >= 'a' && r <= 'z' {
			continue
		}
		if r >= '0' && r <= '9' {
			continue
		}
		if strings.ContainsRune("_-", r) {
			continue
		}
		return false
	}
	return true
}

func ValidNote(note string) bool {
	const maxLen = 500
	return len(note) <= maxLen
}
