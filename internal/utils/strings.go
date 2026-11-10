package utils

import "strings"

func MixCase(s string) string {
	if len(s) == 0 {
		return s
	}
	first := string([]rune(s)[0])
	rest := string([]rune(s)[1:])

	if strings.ToUpper(first) == first {
		return strings.ToLower(first) + rest
	}
	return strings.ToUpper(first) + rest
}