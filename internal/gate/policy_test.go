package gate_test

import (
	"testing"

	"github.com/codesentinel/codesentinel/internal/gate"
	"github.com/codesentinel/codesentinel/internal/models"
)

func TestEvaluatePullRequest_Pass(t *testing.T) {
	policy := models.PolicyConfig{
		BlockOnCritical:       true,
		MaxHighFindings:       1,
		MaxRiskScoreThreshold: 45.0,
	}

	findings := []models.Finding{
		{
			Severity:             models.SeverityMedium,
			ConfidenceMultiplier: 1.0,
		},
		{
			Severity:             models.SeverityLow,
			ConfidenceMultiplier: 1.0,
		},
	}

	res := gate.EvaluatePullRequest(findings, policy)
	if !res.Allowed {
		t.Fatalf("Expected PR to pass gate policy, got blocked: %s", res.BlockingReason)
	}
	if res.RiskScore != 8.5 { // 7.0 + 1.5
		t.Fatalf("Expected risk score 8.5, got %f", res.RiskScore)
	}
}

func TestEvaluatePullRequest_BlockOnCritical(t *testing.T) {
	policy := models.PolicyConfig{
		BlockOnCritical:       true,
		MaxHighFindings:       0,
		MaxRiskScoreThreshold: 50.0,
	}

	findings := []models.Finding{
		{
			Severity:                 models.SeverityCritical,
			IsSecuritySensitiveScope: true, // 40 * 1.5 = 60.0
			ConfidenceMultiplier:     1.0,
		},
	}

	res := gate.EvaluatePullRequest(findings, policy)
	if res.Allowed {
		t.Fatalf("Expected PR to be blocked on critical finding")
	}
	if res.CriticalCount != 1 {
		t.Fatalf("Expected CriticalCount 1, got %d", res.CriticalCount)
	}
}

func TestEvaluatePullRequest_AcceptedRiskExempt(t *testing.T) {
	policy := models.PolicyConfig{
		BlockOnCritical:       true,
		MaxHighFindings:       0,
		MaxRiskScoreThreshold: 30.0,
	}

	findings := []models.Finding{
		{
			Severity: models.SeverityCritical,
			Status:   models.StatusAcceptedRisk,
		},
	}

	res := gate.EvaluatePullRequest(findings, policy)
	if !res.Allowed {
		t.Fatalf("Expected accepted risk finding to not block gate")
	}
	if res.RiskScore != 0.0 {
		t.Fatalf("Expected risk score 0 for accepted risk findings, got %f", res.RiskScore)
	}
}
