package profiles

import (
	"fmt"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	MinDisplayNameLength = 2
	MaxDisplayNameLength = 32
	MaxBiographyLength   = 280
)

func NormalizeName(input string) (displayName string, normalizedName string, err error) {
	value := norm.NFKC.String(input)
	fields := strings.FieldsFunc(value, unicode.IsSpace)
	value = strings.Join(fields, " ")

	length := len([]rune(value))
	if length < MinDisplayNameLength || length > MaxDisplayNameLength {
		return "", "", fmt.Errorf("display name must contain between %d and %d characters", MinDisplayNameLength, MaxDisplayNameLength)
	}

	for _, r := range value {
		if isBidiControl(r) || unicode.IsControl(r) || !isAllowedNameRune(r) {
			return "", "", fmt.Errorf("display name contains unsupported characters")
		}
	}

	return value, strings.ToLower(value), nil
}

func isAllowedNameRune(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r) || r == ' ' || r == '\'' || r == '’' || r == '-' || r == '.' || r == '_'
}

func isBidiControl(r rune) bool {
	switch r {
	case '\u061c', '\u200e', '\u200f', '\u202a', '\u202b', '\u202c', '\u202d', '\u202e', '\u2066', '\u2067', '\u2068', '\u2069':
		return true
	default:
		return false
	}
}
