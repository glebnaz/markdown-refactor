package files

import "strings"

// SplitFrontmatter separates YAML frontmatter from the body.
// Returns (frontmatter including delimiters, body).
// If no frontmatter found, returns ("", original content).
func SplitFrontmatter(content string) (string, string) {
	if !strings.HasPrefix(content, "---\n") && !strings.HasPrefix(content, "---\r\n") {
		return "", content
	}

	// Find closing delimiter after the opening one.
	rest := content[4:] // skip "---\n"
	idx := strings.Index(rest, "\n---\n")
	if idx == -1 {
		// Try Windows line endings.
		idx = strings.Index(rest, "\r\n---\r\n")
		if idx == -1 {
			return "", content
		}
		end := 4 + idx + len("\r\n---\r\n")
		return content[:end], content[end:]
	}

	end := 4 + idx + len("\n---\n")
	return content[:end], content[end:]
}
