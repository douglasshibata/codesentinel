package gate

import (
	"fmt"
	"github.com/codesentinel/codesentinel/internal/models"
)

type GateResult struct {
	Allowed        bool
	RiskScore      float64
	BlockingReason string
	CriticalCount  int
	HighCount      int
}

func EvaluatePullRequest(findings []models.Finding, policy models.PolicyConfig) GateResult {
	result := GateResult{Allowed: true}
	var totalScore float64

	for _, f := range findings {
		if f.Status == models.StatusAcceptedRisk {
			continue
		}

		var weight float64
		switch f.Severity {
		case models.SeverityCritical:
			result.CriticalCount++
			weight = 40.0
		case models.SeverityHigh:
			result.HighCount++
			weight = 20.0
		case models.SeverityMedium:
			weight = 7.0
		case models.SeverityLow, models.SeverityInfo:
			weight = 1.5
		}

		scopeMultiplier := 1.0
		if f.IsSecuritySensitiveScope {
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
	result.RiskScore = totalScore

	// Validação de Limiares da Política
	if result.CriticalCount > 0 && policy.BlockOnCritical {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: %d achados CRÍTICOS encontrados.", result.CriticalCount)
		return result
	}

	if result.HighCount > policy.MaxHighFindings {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: Achados HIGH (%d) excedem o limite tolerado de %d.", result.HighCount, policy.MaxHighFindings)
		return result
	}

	if result.RiskScore >= policy.MaxRiskScoreThreshold {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: Pontuação de risco (%.1f) excede o teto permitido de %.1f.", result.RiskScore, policy.MaxRiskScoreThreshold)
		return result
	}

	return result
}
