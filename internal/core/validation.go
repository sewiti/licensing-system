package core

import (
	"net/mail"
	"strings"
)

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
		switch {
		case strings.ContainsRune("_-", r),
			r >= 'a' && r <= 'z',
			r >= 'A' && r <= 'Z',
			r >= '0' && r <= '9':
		default:
			return false
		}
	}
	return true
}

func ValidLicenseNote(note string) bool {
	const maxLen = 500
	return len(note) <= maxLen
}

func ValidLicenseName(name string) bool {
	const maxLen = 64
	return len(name) <= maxLen
}

func ValidLicenseTags(tags []string) bool {
	const (
		maxTags = 20

		minTagLen = 1
		maxTagLen = 64
	)
	if len(tags) > maxTags {
		return false
	}
	for _, t := range tags {
		if len(t) < minTagLen {
			return false
		}
		if len(t) > maxTagLen {
			return false
		}
	}
	return true
}

func ValidEmail(email string) bool {
	const maxLen = 128
	if len(email) > maxLen {
		return false
	}
	addr, err := mail.ParseAddress(email)
	if err != nil {
		return false
	}
	return addr.Address == email
}

func ValidPhoneNumber(phoneNumber string) bool {
	const (
		maxLen = 24

		minDigits = 7
		maxDigits = 15
	)
	if len(phoneNumber) > maxLen {
		return false
	}
	openBracket := false
	digits := 0
	for i, r := range phoneNumber {
		switch {
		case i == 0 && r == '+':
		case r >= '0' && r <= '9':
			digits++
		case strings.ContainsRune("- ", r):
		case r == '(':
			if openBracket {
				return false
			}
			openBracket = true
		case r == ')':
			if !openBracket {
				return false
			}
			openBracket = false
		default:
			return false
		}
	}
	switch {
	case openBracket,
		digits < minDigits,
		digits > maxDigits:
		return false
	default:
		return true
	}
}

func ValidProductName(name string) bool {
	const maxLen = 64
	return len(name) <= maxLen
}
