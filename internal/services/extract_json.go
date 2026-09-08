package services

import "strings"

// ExtractJSON extracts the outermost JSON object from an LLM response text.
// Mirrors the Python services/extract_json.py logic.
func ExtractJSON(text string) string {
	cleaned := strings.TrimSpace(text)

	// Strip code fences
	for _, prefix := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(cleaned, prefix) {
			cleaned = cleaned[len(prefix):]
		}
	}
	if strings.HasSuffix(cleaned, "```") {
		cleaned = cleaned[:len(cleaned)-3]
	}
	cleaned = strings.TrimSpace(cleaned)

	start := strings.Index(cleaned, "{")
	if start == -1 {
		return ""
	}

	depth := 0
	for i := start; i < len(cleaned); i++ {
		switch cleaned[i] {
		case '{':
			depth++
		case '}':
			depth--
			if depth == 0 {
				return cleaned[start : i+1]
			}
		}
	}
	return ""
}
