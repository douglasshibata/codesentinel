package tests

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/codesentinel/codesentinel/internal/api"
	"github.com/codesentinel/codesentinel/internal/config"
	"github.com/codesentinel/codesentinel/internal/db"
	"github.com/codesentinel/codesentinel/internal/fixver"
	"github.com/codesentinel/codesentinel/internal/git"
	"github.com/codesentinel/codesentinel/internal/models"
	"github.com/codesentinel/codesentinel/internal/ocr"
	"github.com/codesentinel/codesentinel/internal/policy"
	"github.com/codesentinel/codesentinel/internal/risk"
)

func TestFingerprintConsistency(t *testing.T) {
	fp1 := ocr.GenerateFingerprint("service.go", "SEC-001", "apiKey := \"secret123\"")
	fp2 := ocr.GenerateFingerprint("service.go", "SEC-001", "apiKey   :=   \"secret123\"")
	if fp1 != fp2 {
		t.Fatalf("Expected fingerprints to match after whitespace normalization, got %s vs %s", fp1, fp2)
	}
}

func TestRiskScoringEngine(t *testing.T) {
	eng := risk.NewRiskEngine()
	findings := []models.Finding{
		{Severity: "CRITICAL", Category: "SQL injection", File: "payments/service.go"},
		{Severity: "HIGH", Category: "Secrets", File: "auth/jwt.go"},
	}

	score, level := eng.CalculatePRRisk(findings)
	if score <= 50 {
		t.Fatalf("Expected high risk score for CRITICAL+HIGH findings, got %d", score)
	}
	if level != "CRITICAL" && level != "HIGH" {
		t.Fatalf("Expected CRITICAL or HIGH risk level, got %s", level)
	}
}

func TestPolicyEngineEvaluation(t *testing.T) {
	eng := policy.NewPolicyEngine()
	pol := &models.Policy{
		BlockOnSeverity: "HIGH",
		BlockSecurity:   true,
		MaxHighFindings: 0,
	}

	findings := []models.Finding{
		{Severity: "HIGH", Category: "SQL injection", Status: "Open"},
	}

	res := eng.Evaluate(pol, findings)
	if res.Passed {
		t.Fatalf("Expected policy evaluation to fail on HIGH severity finding")
	}
	if !res.Blocked {
		t.Fatalf("Expected policy evaluation to block merge")
	}
}

func TestFixVerificationEngine(t *testing.T) {
	eng := fixver.NewEngine()

	fp1 := ocr.GenerateFingerprint("main.go", "SEC-001", "key := \"123\"")
	fp2 := ocr.GenerateFingerprint("main.go", "SEC-002", "query := \"SELECT\"")

	prev := []models.Finding{
		{Fingerprint: fp1, Status: "Open"},
		{Fingerprint: fp2, Status: "Open"},
	}

	curr := []models.Finding{
		{Fingerprint: fp2, Status: "Open"},
	}

	res := eng.CompareReviews(prev, curr)
	if res.FixedCount != 1 {
		t.Fatalf("Expected 1 fixed finding, got %d", res.FixedCount)
	}
	if res.RemainingCount != 1 {
		t.Fatalf("Expected 1 remaining finding, got %d", res.RemainingCount)
	}
}

func TestDiffParser(t *testing.T) {
	parser := git.NewDiffParser()
	diff := `diff --git a/main.go b/main.go
new file mode 100644
@@ -0,0 +1,5 @@
+package main
+func main() {}
`
	files, err := parser.ParseRawDiff(diff)
	if err != nil {
		t.Fatalf("Diff parsing failed: %v", err)
	}
	if len(files) != 1 {
		t.Fatalf("Expected 1 parsed changed file, got %d", len(files))
	}
	if files[0].Status != "added" {
		t.Fatalf("Expected status 'added', got %s", files[0].Status)
	}
}

func TestAPIServerRoutes(t *testing.T) {
	tmpDBPath := filepath.Join(t.TempDir(), "test_codesentinel.db")
	cfg := &config.Config{
		ServerPort: "8080",
		DBType:     "sqlite",
		DBPath:     tmpDBPath,
	}

	database, err := db.InitDB(cfg)
	if err != nil {
		t.Fatalf("Failed to init DB: %v", err)
	}
	defer func() {
		_ = os.Remove(tmpDBPath)
	}()

	srv := api.NewServer(cfg, database)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/v1/metrics", nil)
	srv.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Expected status 200 on /api/v1/metrics, got %d", w.Code)
	}
}
