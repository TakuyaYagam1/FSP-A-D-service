package domain

import (
	"strings"
	"unicode"
)

func NormalizeUsername(value string) (string, error) {
	username := strings.ToLower(strings.TrimSpace(value))
	if len(username) < 3 || len(username) > 32 {
		return "", ErrInvalidInput
	}

	for _, char := range username {
		if !isValidUsernameChar(char) {
			return "", ErrInvalidInput
		}
	}
	return username, nil
}

func isValidUsernameChar(char rune) bool {
	return (char >= 'a' && char <= 'z') ||
		(char >= '0' && char <= '9') ||
		char == '_' || char == '-' || char == '.'
}

func ValidatePassword(value string) error {
	if len(value) < 10 || len(value) > 72 {
		return ErrInvalidInput
	}
	return validatePasswordCharacters(value)
}

func ValidateLoginPassword(value string) error {
	if len(value) < 1 || len(value) > 72 {
		return ErrInvalidInput
	}
	return validatePasswordCharacters(value)
}

func validatePasswordCharacters(value string) error {
	for _, char := range value {
		if unicode.IsControl(char) {
			return ErrInvalidInput
		}
	}
	return nil
}

func ValidateRouteText(value string) (string, error) {
	text := strings.TrimSpace(value)
	if text == "" || len(text) > 120 {
		return "", ErrInvalidInput
	}
	for _, char := range text {
		if unicode.IsControl(char) {
			return "", ErrInvalidInput
		}
	}
	return text, nil
}
