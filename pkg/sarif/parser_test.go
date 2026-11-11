package sarif_test

import (
	"strings"
	"testing"

	"github.com/codesentinel/codesentinel/internal/models"
	"github.com/codesentinel/codesentinel/pkg/sarif"
)

func TestParseSARIFReaderAndNormalization(t *testing.T) {
	sarifJSON := `{
  "version": "2.1.0",
  "$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json",
  "runs": [
    {
      "tool": {
        "driver": {
          "name": "Semgrep",
          "rules": [
            {
              "id": "rules.go.sql-injection",
              "defaultConfiguration": {
                "level": "error"
              }
            }
          ]
        }
      },
      "results": [
        {
          "ruleId": "rules.go.sql-injection",
          "level": "error",
          "message": {
            "text": "SQL injection detected in DB query"
          },
          "locations": [
            {
              "physicalLocation": {
                "artifactLocation": {
                  "uri": "core/security/auth.go"
                },
                "region": {
                  "startLine": 45,
                  "endLine": 46,
                  "snippet": {
                    "text": "db.Raw(\"SELECT * FROM users WHERE id = \" + id)"
                  }
                }
              }
            }
          ]
        }
      ]
    }
  ]
}`

	report, err := sarif.ParseSARIFReader(strings.NewReader(sarifJSON))
	if err != nil {
		t.Fatalf("Failed to parse SARIF: %v", err)
	}

	findings := sarif.NormalizeSARIFToFindings(report, "rev-001", "repo-main", "PR-10", "c0ffee")
	if len(findings) != 1 {
		t.Fatalf("Expected 1 normalized finding, got %d", len(findings))
	}

	f := findings[0]
	if f.RuleID != "rules.go.sql-injection" {
		t.Fatalf("Expected rule ID 'rules.go.sql-injection', got %s", f.RuleID)
	}
	if f.Severity != models.SeverityCritical {
		t.Fatalf("Expected CRITICAL severity for SARIF error level, got %s", f.Severity)
	}
	if !f.IsSecuritySensitiveScope {
		t.Fatalf("Expected core/security/auth.go to be marked as security sensitive scope")
	}
	if f.Fingerprint == "" {
		t.Fatalf("Expected non-empty structural fingerprint")
	}
}
