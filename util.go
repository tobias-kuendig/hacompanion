package main

import (
	"strings"
	"unicode"
)

// ToSnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
// @see https://gist.github.com/elwinar/14e1e897fdbe4d3432e1#gistcomment-2246837
func ToSnakeCase(in string) string {
	runes := []rune(strings.ReplaceAll(in, " ", "_"))

	var out []rune
	for i := 0; i < len(runes); i++ {
		if i > 0 && (unicode.IsUpper(runes[i]) || unicode.IsNumber(runes[i])) && ((i+1 < len(runes) && unicode.IsLower(runes[i+1])) || unicode.IsLower(runes[i-1])) {
			out = append(out, '_')
		}
		out = append(out, unicode.ToLower(runes[i]))
	}

	return string(out)
}
