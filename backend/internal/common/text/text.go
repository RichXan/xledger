package text

import "strings"

// StripEmojiPrefix removes the emoji prefix from a category name.
// Categories often have an emoji followed by a space as a prefix.
// If the first word contains non-ASCII (emoji), it strips the prefix.
// Otherwise, returns the name unchanged.
func StripEmojiPrefix(name string) string {
	if idx := strings.Index(name, " "); idx > 0 {
		firstPart := name[:idx]
		hasNonASCII := false
		for _, r := range firstPart {
			if r > 127 {
				hasNonASCII = true
				break
			}
		}
		if hasNonASCII {
			return strings.TrimSpace(name[idx:])
		}
	}
	return name
}
