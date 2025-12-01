package tools

import (
	"encoding/json"
	"regexp"
	"strings"

	"github.com/DanielNikkari/maahinen/internal/llm"
)

func ParseToolCallFromContent(content string) (*llm.ToolCall, bool) {
	content = strings.TrimSpace(content)

	// Try to extract name
	namePattern := regexp.MustCompile(`"name"\s*:\s*"([^"]+)"`)
	nameMatch := namePattern.FindStringSubmatch(content)
	if nameMatch == nil {
		return nil, false
	}
	name := nameMatch[1]

	// Try to find path argument
	pathPattern := regexp.MustCompile(`"path"\s*:\s*"([^"]+)"`)
	pathMatch := pathPattern.FindStringSubmatch(content)

	// Try to find command argument - (?s) makes . match newlines
	cmdPattern := regexp.MustCompile(`(?s)"command"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	cmdMatch := cmdPattern.FindStringSubmatch(content)

	// Try to find content argument - (?s) makes . match newlines
	contentPattern := regexp.MustCompile(`(?s)"content"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	contentMatch := contentPattern.FindStringSubmatch(content)

	// Try old_string and new_string for edit tool
	oldStrPattern := regexp.MustCompile(`(?s)"old_string"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	oldStrMatch := oldStrPattern.FindStringSubmatch(content)

	newStrPattern := regexp.MustCompile(`(?s)"new_string"\s*:\s*"((?:[^"\\]|\\.)*)"`)
	newStrMatch := newStrPattern.FindStringSubmatch(content)

	args := make(map[string]any)

	if pathMatch != nil {
		args["path"] = pathMatch[1]
	}
	if cmdMatch != nil {
		args["command"] = unescapeJSON(cmdMatch[1])
	}
	if contentMatch != nil {
		args["content"] = unescapeJSON(contentMatch[1])
	}
	if oldStrMatch != nil {
		args["old_string"] = unescapeJSON(oldStrMatch[1])
	}
	if newStrMatch != nil {
		args["new_string"] = unescapeJSON(newStrMatch[1])
	}

	// Try standard JSON parsing as fallback
	if len(args) == 0 {
		jsonContent := extractJSON(content)
		if jsonContent != "" {
			var tc struct {
				Name       string         `json:"name"`
				Arguments  map[string]any `json:"arguments"`
				Parameters map[string]any `json:"parameters"`
			}
			if err := json.Unmarshal([]byte(jsonContent), &tc); err == nil {
				if tc.Arguments != nil {
					args = tc.Arguments
				} else if tc.Parameters != nil {
					args = tc.Parameters
				}
			}
		}
	}

	if name == "" || len(args) == 0 {
		return nil, false
	}

	return &llm.ToolCall{
		Function: llm.ToolFunction{
			Name:      name,
			Arguments: args,
		},
	}, true
}

func extractJSON(content string) string {
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")
	content = strings.TrimSpace(content)

	content = fixBacktickStrings(content)

	start := strings.Index(content, "{")
	if start == -1 {
		return ""
	}

	depth := 0
	inString := false
	escaped := false

	for i := start; i < len(content); i++ {
		c := content[i]

		if escaped {
			escaped = false
			continue
		}

		if c == '\\' && inString {
			escaped = true
			continue
		}

		if c == '"' {
			inString = !inString
			continue
		}

		if inString {
			continue
		}

		if c == '{' {
			depth++
		} else if c == '}' {
			depth--
			if depth == 0 {
				return content[start : i+1]
			}
		}
	}

	return ""
}

func fixBacktickStrings(content string) string {
	result := strings.Builder{}
	i := 0

	for i < len(content) {
		// Check if we're at a backtick that's part of a JSON value
		if content[i] == '`' {
			// Find the closing backtick
			end := strings.Index(content[i+1:], "`")
			if end == -1 {
				result.WriteByte(content[i])
				i++
				continue
			}

			// Extract the content between backticks
			inner := content[i+1 : i+1+end]

			// Convert to JSON string: escape quotes and newlines
			jsonStr := escapeForJSON(inner)
			result.WriteString(`"` + jsonStr + `"`)

			i = i + 1 + end + 1 // Skip past closing backtick
			continue
		}

		result.WriteByte(content[i])
		i++
	}

	return result.String()
}

func escapeForJSON(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, "\n", `\n`)
	s = strings.ReplaceAll(s, "\r", `\r`)
	s = strings.ReplaceAll(s, "\t", `\t`)
	return s
}

func unescapeJSON(s string) string {
	s = strings.ReplaceAll(s, `\n`, "\n")
	s = strings.ReplaceAll(s, `\t`, "\t")
	s = strings.ReplaceAll(s, `\"`, `"`)
	s = strings.ReplaceAll(s, `\\`, `\`)
	return s
}
