package tag_parsers

import (
	"regexp"
	"strconv"
	"strings"
)

// Precompile once at package scope
var (
	reDQString  = regexp.MustCompile(`"(?:\\.|[^"\\])*"`) // "…", allows \" and escaped chars
	reRawString = regexp.MustCompile("`[^`]*`")           // `…` (no escapes inside)
	reIndent    = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_]*)`)
)

// firstIdent returns the first Go identifier from a line.
func firstIdent(s string) string {
	s = strings.TrimSpace(s)
	for s != "" && (s[0] == '*' || s[0] == '&') {
		s = s[1:]
	}
	// identifier: letters, numbers, underscores (start with letter or _)
	m := reIndent.FindStringSubmatch(s)
	if len(m) > 1 {
		return m[1]
	}
	return ""
}

// firstStringLiteral finds the first Go string literal in s (either "..." or `...`)
// and returns its unquoted value.
func firstStringLiteral(s string) (string, bool) {
	di := reDQString.FindStringIndex(s)
	ri := reRawString.FindStringIndex(s)

	var start, end int
	switch {
	case di == nil && ri == nil:
		return "", false
	case di != nil && (ri == nil || di[0] < ri[0]):
		start, end = di[0], di[1]
	default:
		start, end = ri[0], ri[1]
	}

	lit := s[start:end]
	v, err := strconv.Unquote(lit)
	if err != nil {
		return "", false
	}
	return v, true
}
