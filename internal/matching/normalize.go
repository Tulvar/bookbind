package matching

import (
	"regexp"
	"strings"
	"unicode"
)

var spacesRE = regexp.MustCompile(`\s+`)

func Normalize(value string) string {
	value = strings.ToLower(value)
	value = strings.ReplaceAll(value, "ё", "е")

	var builder strings.Builder
	for _, r := range value {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			builder.WriteRune(r)
			continue
		}
		builder.WriteRune(' ')
	}

	return strings.TrimSpace(spacesRE.ReplaceAllString(builder.String(), " "))
}

func SamePerson(a, b string) bool {
	a = Normalize(a)
	b = Normalize(b)
	if a == "" || b == "" {
		return false
	}
	if a == b {
		return true
	}

	aParts := strings.Fields(a)
	bParts := strings.Fields(b)
	if len(aParts) == 2 && len(bParts) == 2 {
		return aParts[0] == bParts[1] && aParts[1] == bParts[0]
	}
	return false
}
