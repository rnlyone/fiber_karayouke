package artisan

import (
	"strings"
	"unicode"
)

func toPascalCase(input string) string {
	if input == "" {
		return ""
	}
	parts := strings.FieldsFunc(input, func(r rune) bool {
		switch r {
		case '_', '-', ' ', '/':
			return true
		}
		return false
	})

	var builder strings.Builder
	for _, part := range parts {
		if part == "" {
			continue
		}
		builder.WriteString(capitalize(part))
	}
	return builder.String()
}

func toSnakeCase(input string) string {
	if input == "" {
		return ""
	}
	var result []rune
	for i, r := range input {
		if unicode.IsUpper(r) {
			if i > 0 {
				result = append(result, '_')
			}
			result = append(result, unicode.ToLower(r))
		} else {
			result = append(result, unicode.ToLower(r))
		}
	}
	return string(result)
}

func capitalize(s string) string {
	if s == "" {
		return ""
	}
	runes := []rune(strings.ToLower(s))
	runes[0] = unicode.ToUpper(runes[0])
	return string(runes)
}
