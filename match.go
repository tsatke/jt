package jt

import (
	"strings"
)

func ClassNameMatches(path string, search string) bool {
	return strings.Contains(strings.ToLower(path), strings.ToLower(search))
}
