package util

import (
	"regexp"
	"strings"
)

var nonPrintSymRegex = regexp.MustCompile(`\p{C}`)
var whiteSpaceRegex = regexp.MustCompile(`\p{Zs}+`)

func Sanitize(raw string) (txt string) {
	txt = nonPrintSymRegex.ReplaceAllString(raw, " ") // clean from the non-printable symbols 1st
	txt = whiteSpaceRegex.ReplaceAllString(txt, " ")
	txt = strings.TrimSpace(txt)
	txt = strings.ToLower(txt)
	return
}
