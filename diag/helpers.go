//go:build windows

package diag

import (
	"strings"
)

func parseDeletePaths(cmdLine string) (path32, path64 string) {
	rest := strings.TrimPrefix(cmdLine, CmdDeleteFiles)
	rest = strings.TrimSpace(rest)
	parts := extractQuotedStrings(rest)
	if len(parts) > 0 {
		path64 = parts[0]
	}
	if len(parts) > 1 {
		path32 = parts[1]
	}
	return path32, path64
}

func extractQuotedStrings(s string) []string {
	var result []string
	for {
		start := strings.Index(s, `"`)
		if start < 0 {
			break
		}
		s = s[start+1:]
		end := strings.Index(s, `"`)
		if end < 0 {
			break
		}
		result = append(result, s[:end])
		s = s[end+1:]
	}
	return result
}

func stringsHasPrefix(s, prefix string) bool {
	return len(s) >= len(prefix) && s[:len(prefix)] == prefix
}

func stringsJoinArgs(args []string) string {
	if len(args) == 0 {
		return ""
	}
	result := args[0]
	for i := 1; i < len(args); i++ {
		if strings.Contains(args[i], " ") {
			result += ` "` + args[i] + `"`
		} else {
			result += " " + args[i]
		}
	}
	return result
}
