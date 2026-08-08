// Package textutil holds string operations that respect UTF-8 rune boundaries.
//
// Cutting a Go string at a fixed byte offset splits whatever rune straddles that
// offset, and the result is not valid UTF-8. Nothing reports it: encoding/json
// substitutes U+FFFD for the invalid bytes rather than returning an error, so a
// corrupted string reaches the dashboard with no error raised anywhere along the
// path. It is silent by construction, which is why it survived across the fleet.
package textutil

import (
	"unicode"
	"unicode/utf8"
)

// TruncateAtRuneBoundary returns s shortened to at most maxBytes bytes, cutting
// only where a rune ends. A rune straddling the cut is dropped whole rather than
// split, so the result is valid UTF-8 whenever s is.
//
// It appends nothing. Callers keep their own ellipsis arithmetic, because the
// fleet has not settled whether an ellipsis belongs inside the byte budget or
// outside it, and answering that here would change every caller's output length.
func TruncateAtRuneBoundary(s string, maxBytes int) string {
	if maxBytes <= 0 {
		return ""
	}
	if len(s) <= maxBytes {
		return s
	}
	cut := maxBytes
	// s[cut] is the first byte that would be dropped. If it does not start a
	// rune, the cut lands inside one, so walk back until it does.
	for cut > 0 && !utf8.RuneStart(s[cut]) {
		cut--
	}
	return s[:cut]
}

// UpperFirstRune returns s with its first rune upper-cased and the rest left
// alone.
//
// The obvious spelling of this, strings.ToUpper(s[:1]) + s[1:], corrupts any
// string whose first rune is multi-byte — and corrupts it worse than a plain
// byte cut does. strings.ToUpper replaces the stray leading byte with a
// three-byte U+FFFD and the continuation bytes are left stranded behind it, so
// "🎯quest" comes back as invalid UTF-8 rather than merely mis-capitalised.
func UpperFirstRune(s string) string {
	if s == "" {
		return s
	}
	first, width := utf8.DecodeRuneInString(s)
	if first == utf8.RuneError && width <= 1 {
		// s already starts with an invalid byte. Upper-casing it would replace
		// that byte with U+FFFD and change the string's length; leave it be.
		return s
	}
	return string(unicode.ToUpper(first)) + s[width:]
}
