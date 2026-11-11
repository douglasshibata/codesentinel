package patch

import (
	"fmt"
	"strings"

	"github.com/codesentinel/codesentinel/internal/models"
)

type PatchGenerator struct{}

func NewPatchGenerator() *PatchGenerator {
	return &PatchGenerator{}
}

// GenerateFixPatch creates a BEFORE / AFTER code diff proposal for a given finding
func (g *PatchGenerator) GenerateFixPatch(f *models.Finding) (*models.Patch, error) {
	oldCode := f.CodeSnippet
	if oldCode == "" {
		oldCode = fmt.Sprintf("// Code at %s:%d", f.File, f.Line)
	}

	newCode := oldCode
	explanation := "Proposed automated fix patch"

	switch f.Category {
	case "Secrets":
		newCode = "apiKey := os.Getenv(\"API_KEY\") // Credentials loaded from environment variable"
		explanation = "Replaced hardcoded credential with os.Getenv lookup."
	case "SQL injection":
		newCode = "db.Query(\"SELECT * FROM users WHERE id = $1\", userID)"
		explanation = "Replaced string concatenation with parameterized SQL placeholder."
	case "Command injection":
		newCode = "exec.Command(\"ls\", \"-l\", sanitizedPath)"
		explanation = "Passed sanitized arguments as slice instead of string concatenation."
	case "XSS":
		newCode = "safeText := html.EscapeString(userInput)"
		explanation = "Applied HTML entity escaping to untrusted input."
	default:
		newCode = fmt.Sprintf("// Sanitized and validated: %s", strings.TrimSpace(f.SuggestedFix))
		explanation = f.SuggestedFix
	}

	diff := fmt.Sprintf("--- a/%s\n+++ b/%s\n@@ -%d,1 +%d,1 @@\n- %s\n+ %s\n",
		f.File, f.File, f.Line, f.Line, oldCode, newCode)

	patch := &models.Patch{
		ID:          fmt.Sprintf("patch-%s", f.ID),
		FindingID:   f.ID,
		Original:    oldCode,
		Proposed:    newCode,
		Diff:        diff,
		Explanation: explanation,
	}

	return patch, nil
}
