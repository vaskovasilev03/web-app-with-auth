package validator

import (
	"regexp"
	"strings"
	"unicode"
)

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

func IsValidEmail(email string) bool {
	email = strings.TrimSpace(email)
	if len(email) < 1 || len(email) > 100 {
		return false
	}

	if !emailRegex.MatchString(email) {
		return false
	}

	if strings.Contains(email, "..") {
		return false
	}

	localPart := strings.Split(email, "@")[0]
	if strings.HasPrefix(localPart, ".") || strings.HasSuffix(localPart, ".") {
		return false
	}

	return true
}

func IsValidPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool

	specialChars := "!@#$%^&*()-_=+[]{}|;:'\",.<>?/`~"

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case strings.ContainsRune(specialChars, char):
			hasSpecial = true
		}
	}

	return hasUpper && hasLower && hasDigit && hasSpecial
}

func IsValidName(name string) bool {
	name = strings.TrimSpace(name)
	if len(name) < 1 || len(name) > 50 {
		return false
	}
	for _, char := range name {
		if !unicode.IsLetter(char) && char != ' ' {
			return false
		}
	}
	return true
}
