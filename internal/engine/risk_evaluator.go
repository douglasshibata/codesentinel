package engine

import (
	"strings"

	"github.com/codesentinel/codesentinel/internal/models"
)

// RiskEvaluator calculates deterministic Pull Request risk score (0-100)
type RiskEvaluator struct{}

func NewRiskEvaluator() *RiskEvaluator {
	return &RiskEvaluator{}
}

// CalculatePRRisk computes the PR Risk Score according to the specification:
// Severity Weights: CRITICAL: 40.0, HIGH: 20.0, MEDIUM: 7.0, LOW/INFO: 1.5
// Scope Multipliers: core/security/**, auth/**, *.tf, Dockerfile: 1.5x; internal/**, services/**: 1.0x; tests/**, docs/**: 0.2x
func (e *RiskEvaluator) CalculatePRRisk(findings []models.Finding) (float64, string) {
	if len(findings) == 0 {
		return 0.0, "LOW"
	}

	var totalScore float64

	for _, f := range findings {
		if f.Status == models.StatusAcceptedRisk {
			continue
		}

		var weight float64
		switch strings.ToUpper(f.Severity) {
		case models.SeverityCritical:
			weight = 40.0
		case models.SeverityHigh:
			weight = 20.0
		case models.SeverityMedium:
			weight = 7.0
		case models.SeverityLow, models.SeverityInfo:
			weight = 1.5
		default:
			weight = 1.5
		}

		scopeMultiplier := GetScopeMultiplier(f.File)
		if f.IsSecuritySensitiveScope && scopeMultiplier < 1.5 {
			scopeMultiplier = 1.5
		}

		confMultiplier := f.ConfidenceMultiplier
		if confMultiplier <= 0 {
			confMultiplier = 1.0
		}

		totalScore += weight * scopeMultiplier * confMultiplier
	}

	if totalScore > 100.0 {
		totalScore = 100.0
	}

	level := "LOW"
	switch {
	case totalScore >= 60.0:
		level = "CRITICAL"
	case totalScore >= 40.0:
		level = "HIGH"
	case totalScore >= 20.0:
		level = "MODERATE"
	default:
		level = "LOW"
	}

	return totalScore, level
}

// GetScopeMultiplier returns W_scope based on file path patterns
func GetScopeMultiplier(filePath string) float64 {
	p := strings.ToLower(filePath)
	if strings.HasPrefix(p, "tests/") || strings.HasPrefix(p, "test/") || strings.Contains(p, "/test/") || strings.HasPrefix(p, "docs/") || strings.HasSuffix(p, ".md") {
		return 0.2
	}
	if strings.Contains(p, "security") || strings.Contains(p, "auth") || strings.HasSuffix(p, ".tf") || strings.Contains(p, "dockerfile") {
		return 1.5
	}
	if strings.HasPrefix(p, "internal/") || strings.HasPrefix(p, "services/") || strings.Contains(p, "service") {
		return 1.0
	}
	return 1.0
}
