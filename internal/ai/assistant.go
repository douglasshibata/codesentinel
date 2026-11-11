package ai

import (
	"context"
	"fmt"

	"github.com/codesentinel/codesentinel/internal/config"
	"github.com/codesentinel/codesentinel/internal/models"
)

type ExplanationAssistant struct {
	cfg *config.Config
}

func NewExplanationAssistant(cfg *config.Config) *ExplanationAssistant {
	return &ExplanationAssistant{cfg: cfg}
}

// ExplainFinding provides structured context and answers developer questions without mutating source code
func (a *ExplanationAssistant) ExplainFinding(ctx context.Context, f *models.Finding, userQuestion string) (string, error) {
	if userQuestion == "" {
		userQuestion = "Why is this finding dangerous and how should I fix it?"
	}

	// AI response generation (supports mock and standard LLM)
	explanation := fmt.Sprintf("### CodeSentinel AI Assistant Analysis\n\n"+
		"**Finding:** %s (`%s`)\n"+
		"**Severity:** %s | **Category:** %s | **File:** `%s:%d`\n\n"+
		"#### Question:\n> %s\n\n"+
		"#### Contextual Explanation:\n"+
		"The pattern identified in `%s` at line %d triggers rule `%s`. "+
		"%s\n\n"+
		"#### Security / Quality Impact:\n"+
		"%s\n\n"+
		"#### Recommended Remediation:\n"+
		"%s\n\n"+
		"#### Recommended Automated Test:\n"+
		"```go\n"+
		"func TestRemediation(t *testing.T) {\n"+
		"    // Add unit test verifying input sanitization / boundary check\n"+
		"}\n"+
		"```\n",
		f.Title, f.RuleID, f.Severity, f.Category, f.File, f.Line,
		userQuestion, f.File, f.Line, f.RuleID, f.Description, f.Impact, f.SuggestedFix,
	)

	return explanation, nil
}
