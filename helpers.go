package configreader

import (
	"strings"
)

// -----------------------------------------------------------------------------

// AddressOfString returns the address of the provided string
func AddressOfString(s string) *string {
	return &s
}

// Escape escapes a string to avoid embedded % characters to be recognized as environment variable tags.
func Escape(s string) string {
	return strings.Replace(s, "%", "%%", -1)
}
