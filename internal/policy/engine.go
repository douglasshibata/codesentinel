package policy

import (
	"fmt"
	"strings"

	"github.com/codesentinel/codesentinel/internal/models"
)

type PolicyEngine struct{}

func NewPolicyEngine() *PolicyEngine {
	return &PolicyEngine{}
}

type EvaluationResult struct {
	Passed      bool     `json:"passed"`
	Blocked     bool     `json:"blocked"`
	Reasons     []string `json:"reasons"`
	MaxAllowed  int      `json:"max_allowed"`
	TotalFound  int      `json:"total_found"`
}

// Evaluate checks repository findings against policy constraints deterministically
func (pe *PolicyEngine) Evaluate(policy *models.Policy, findings []models.Finding) EvaluationResult {
	res := EvaluationResult{
		Passed:  true,
		Blocked: false,
		Reasons: []string{},
	}

	if policy == nil {
		return res
	}

	criticalCount := 0
	highCount := 0
	mediumCount := 0
	securityCount := 0

	for _, f := range findings {
		if f.Status == "False Positive" || f.Status == "Accepted Risk" {
			continue
		}

		switch strings.ToUpper(f.Severity) {
		case "CRITICAL":
			criticalCount++
		case "HIGH":
			highCount++
		case "MEDIUM":
			mediumCount++
		}

		if isSecurityCategory(f.Category) {
			securityCount++
		}
	}

	res.TotalFound = len(findings)

	// Severity blocking policy
	if policy.BlockOnSeverity != "" {
		targetSev := strings.ToUpper(policy.BlockOnSeverity)
		if targetSev == "CRITICAL" && criticalCount > 0 {
			res.Passed = false
			res.Blocked = true
			res.Reasons = append(res.Reasons, fmt.Sprintf("Policy violation: Found %d CRITICAL findings (policy blocks on CRITICAL).", criticalCount))
		} else if targetSev == "HIGH" && (criticalCount > 0 || highCount > 0) {
			res.Passed = false
			res.Blocked = true
			res.Reasons = append(res.Reasons, fmt.Sprintf("Policy violation: Found %d CRITICAL/HIGH findings (policy blocks on HIGH+).", criticalCount+highCount))
		}
	}

	// Security issue blocking policy
	if policy.BlockSecurity && securityCount > 0 {
		res.Passed = false
		res.Blocked = true
		res.Reasons = append(res.Reasons, fmt.Sprintf("Policy violation: Security policy enabled and %d security vulnerabilities were detected.", securityCount))
	}

	// Max High findings limit
	if policy.MaxHighFindings >= 0 && highCount > policy.MaxHighFindings {
		res.Passed = false
		res.Blocked = true
		res.Reasons = append(res.Reasons, fmt.Sprintf("Policy violation: High severity count (%d) exceeds max allowed (%d).", highCount, policy.MaxHighFindings))
	}

	// Max Medium findings limit
	if policy.MaxMediumFindings >= 0 && mediumCount > policy.MaxMediumFindings {
		res.Passed = false
		res.Blocked = true
		res.Reasons = append(res.Reasons, fmt.Sprintf("Policy violation: Medium severity count (%d) exceeds max allowed (%d).", mediumCount, policy.MaxMediumFindings))
	}

	return res
}

func isSecurityCategory(category string) bool {
	secCats := []string{"sql injection", "xss", "command injection", "ssrf", "secrets", "authentication", "authorization"}
	catLower := strings.ToLower(category)
	for _, c := range secCats {
		if strings.Contains(catLower, c) {
			return true
		}
	}
	return false
}
