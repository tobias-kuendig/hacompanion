package util

import (
	"math"
	"math/rand"
	"os"
	"strings"
	"unicode"
)

// ToSnakeCase convert the given string to snake case following the Golang format:
// acronyms are converted to lower-case and preceded by an underscore.
// @see https://gist.github.com/elwinar/14e1e897fdbe4d3432e1#gistcomment-2246837
func ToSnakeCase(in string) string {
	in = strings.TrimSpace(in)
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

func RandomString(n int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")

	s := make([]rune, n)
	for i := range s {
		s[i] = letters[rand.Intn(len(letters))]
	}
	return string(s)
}

func StringToOnOff(in string) string {
	in = strings.TrimSpace(in)
	if in == "1" {
		return "on"
	}
	if in == "0" {
		return "off"
	}
	return ""
}

func FileExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func RoundToTwoDecimals(in float64) float64 {
	return math.Floor(in*100) / 100
}
