// Package validate holds small, dependency-free input validators shared across
// HTTP handlers and the install CLI.
package validate

import (
	"regexp"
	"strings"
)

// slugRe matches a non-empty string of only ASCII letters, digits, underscore
// and hyphen.
var slugRe = regexp.MustCompile(`^[A-Za-z0-9_-]+$`)

// Slug reports whether s is safe to embed in a public proxy path
// (/{username}/{slug}/{route}): only letters, digits, underscore and hyphen,
// and non-empty.
func Slug(s string) bool {
	return slugRe.MatchString(s)
}

// Sanitize coerces an arbitrary string into a valid Slug by replacing each run
// of disallowed characters with a single hyphen and trimming leading/trailing
// hyphens. Used for auto-derived identifiers (e.g. a username built from a
// Google profile) where rejecting the input is not an option. Returns "" when
// nothing usable remains; callers must handle that.
func Sanitize(s string) string {
	var b strings.Builder
	prevHyphen := false
	for _, r := range s {
		switch {
		case (r >= 'A' && r <= 'Z') || (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_':
			b.WriteRune(r)
			prevHyphen = false
		default:
			if b.Len() > 0 && !prevHyphen {
				b.WriteByte('-')
				prevHyphen = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}
