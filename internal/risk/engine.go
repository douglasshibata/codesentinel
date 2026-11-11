package risk

import (
	"strings"

	"github.com/codesentinel/codesentinel/internal/models"
)

type RiskEngine struct{}

func NewRiskEngine() *RiskEngine {
	return &RiskEngine{}
}

// CalculatePRRisk computes a deterministic PR Risk Score (0-100) and Level (LOW, MODERATE, HIGH, CRITICAL)
func (e *RiskEngine) CalculatePRRisk(findings []models.Finding) (int, string) {
	if len(findings) == 0 {
		return 0, "LOW"
	}

	score := 0

	for _, f := range findings {
		// Severity weights
		switch strings.ToUpper(f.Severity) {
		case "CRITICAL":
			score += 35
		case "HIGH":
			score += 20
		case "MEDIUM":
			score += 10
		case "LOW":
			score += 3
		case "INFO":
			score += 1
		}

		// Security multiplier
		if isSecurityCategory(f.Category) {
			score += 10
		}

		// Critical path files (payment, auth, crypto, security, main)
		if isCriticalFile(f.File) {
			score += 15
		}
	}

	if score > 100 {
		score = 100
	}

	level := "LOW"
	switch {
	case score >= 71:
		level = "CRITICAL"
	case score >= 41:
		level = "HIGH"
	case score >= 21:
		level = "MODERATE"
	default:
		level = "LOW"
	}

	return score, level
}

func isSecurityCategory(category string) bool {
	secCats := []string{
		"SQL injection", "XSS", "Command injection", "SSRF", "Path traversal",
		"Authentication issues", "Authorization issues", "Secrets", "Unsafe deserialization",
		"Race conditions", "Thread-safety", "Cryptographic mistakes", "Data leakage",
	}
	catLower := strings.ToLower(category)
	for _, c := range secCats {
		if strings.Contains(catLower, strings.ToLower(c)) {
			return true
		}
	}
	return false
}

func isCriticalFile(filePath string) bool {
	pathLower := strings.ToLower(filePath)
	keywords := []string{"auth", "login", "payment", "crypto", "secret", "security", "token", "session", "admin", "db"}
	for _, kw := range keywords {
		if strings.Contains(pathLower, kw) {
			return true
		}
	}
	return false
}
