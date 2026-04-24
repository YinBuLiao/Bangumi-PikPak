package sanitize

import (
	"strings"
	"unicode"
)

func Name(s string) string {
	var b strings.Builder
	lastUnderscore := false
	for _, r := range s {
		invalid := r < 32 || strings.ContainsRune(`<>:"/\|?*`, r) || unicode.IsControl(r)
		if invalid {
			if !lastUnderscore {
				b.WriteRune('_')
				lastUnderscore = true
			}
			continue
		}
		b.WriteRune(r)
		lastUnderscore = r == '_'
	}
	out := strings.TrimSpace(b.String())
	out = strings.TrimRight(out, " ._")
	out = strings.TrimLeft(out, " ._")
	if out == "" {
		return "untitled"
	}
	return out
}
