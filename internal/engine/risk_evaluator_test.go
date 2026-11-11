package engine_test

import (
	"testing"

	"github.com/codesentinel/codesentinel/internal/engine"
	"github.com/codesentinel/codesentinel/internal/models"
)

func TestRiskEvaluatorCalculation(t *testing.T) {
	evaluator := engine.NewRiskEvaluator()

	// Finding in security scope: core/security/auth.go (1.5x)
	// Critical finding: 40.0 * 1.5 = 60.0
	findings := []models.Finding{
		{
			File:                 "core/security/auth.go",
			Severity:             models.SeverityCritical,
			ConfidenceMultiplier: 1.0,
		},
		{
			File:                 "tests/auth_test.go", // 0.2x multiplier
			Severity:             models.SeverityHigh, // 20.0 * 0.2 = 4.0
			ConfidenceMultiplier: 1.0,
		},
	}

	score, level := evaluator.CalculatePRRisk(findings)
	expectedScore := 64.0
	if score != expectedScore {
		t.Fatalf("Expected score %f, got %f", expectedScore, score)
	}
	if level != "CRITICAL" {
		t.Fatalf("Expected level CRITICAL, got %s", level)
	}
}

func TestGetScopeMultiplier(t *testing.T) {
	if m := engine.GetScopeMultiplier("core/security/vault.go"); m != 1.5 {
		t.Fatalf("Expected 1.5 for security scope, got %f", m)
	}
	if m := engine.GetScopeMultiplier("infra/main.tf"); m != 1.5 {
		t.Fatalf("Expected 1.5 for terraform file, got %f", m)
	}
	if m := engine.GetScopeMultiplier("Dockerfile"); m != 1.5 {
		t.Fatalf("Expected 1.5 for Dockerfile, got %f", m)
	}
	if m := engine.GetScopeMultiplier("tests/unit/test.go"); m != 0.2 {
		t.Fatalf("Expected 0.2 for tests, got %f", m)
	}
	if m := engine.GetScopeMultiplier("docs/guide.md"); m != 0.2 {
		t.Fatalf("Expected 0.2 for docs, got %f", m)
	}
	if m := engine.GetScopeMultiplier("internal/service/user.go"); m != 1.0 {
		t.Fatalf("Expected 1.0 for internal service, got %f", m)
	}
}
