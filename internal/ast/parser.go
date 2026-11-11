package ast

import (
	"path/filepath"
	"regexp"
	"strings"
)

var (
	funcRegex = regexp.MustCompile(`(?m)^(?:\s*(?:func|def|function|class)\s+([A-Za-z0-9_]+))`)
)

// ExtractEnclosingScope extracts the containing function, method, or class name around a given line in source.
func ExtractEnclosingScope(content string, targetLine int) string {
	lines := strings.Split(content, "\n")
	if targetLine <= 0 || targetLine > len(lines) {
		return "global"
	}

	currentScope := "global"
	for idx := 0; idx < targetLine; idx++ {
		match := funcRegex.FindStringSubmatch(lines[idx])
		if len(match) > 1 {
			currentScope = match[1]
		}
	}
	return currentScope
}

// NormalizeRelativePath normalizes file paths across different operating systems.
func NormalizeRelativePath(baseDir, targetPath string) string {
	rel, err := filepath.Rel(baseDir, targetPath)
	if err != nil {
		return filepath.ToSlash(targetPath)
	}
	return filepath.ToSlash(rel)
}
