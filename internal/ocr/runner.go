package ocr

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/codesentinel/codesentinel/internal/models"
)

// RawOCRFinding represents raw JSON output format from Open Code Review
type RawOCRFinding struct {
	File        string `json:"file"`
	Line        int    `json:"line"`
	EndLine     int    `json:"end_line"`
	RuleID      string `json:"rule_id"`
	Category    string `json:"category"`
	Severity    string `json:"severity"`
	Confidence  string `json:"confidence"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Impact      string `json:"impact"`
	SuggestedFix string `json:"suggested_fix"`
	CodeSnippet string `json:"code_snippet"`
}

type Runner struct{}

func NewRunner() *Runner {
	return &Runner{}
}

// ExecuteReview runs Open Code Review CLI if available, or falls back to integrated deterministic engine
func (r *Runner) ExecuteReview(repoPath, fromBranch, toBranch string, customRules []models.CustomRule) ([]RawOCRFinding, error) {
	// Attempt to run `open-code-review` CLI tool
	cmd := exec.Command("open-code-review", "scan", "--path", repoPath, "--format", "json")
	out, err := cmd.Output()
	if err == nil && len(out) > 0 {
		var rawFindings []RawOCRFinding
		if jsonErr := json.Unmarshal(out, &rawFindings); jsonErr == nil {
			return rawFindings, nil
		}
	}

	// Fallback: Integrated deterministic static analysis engine + rule matcher
	return r.runIntegratedEngine(repoPath, customRules)
}

func (r *Runner) runIntegratedEngine(repoPath string, customRules []models.CustomRule) ([]RawOCRFinding, error) {
	var findings []RawOCRFinding

	// Standard pattern rules
	builtinRules := []struct {
		Pattern     string
		RuleID      string
		Category    string
		Severity    string
		Title       string
		Description string
		Impact      string
		Fix         string
	}{
		{
			Pattern:     `(?i)(password|secret|api_key|token|private_key)\s*=\s*["'][^"']{6,}["']`,
			RuleID:      "SEC-001",
			Category:    "Secrets",
			Severity:    "CRITICAL",
			Title:       "Hardcoded Credentials or Secret Token Detected",
			Description: "A sensitive credential or API secret token was found hardcoded in the source code.",
			Impact:      "Attachers could extract credentials from source control and compromise services.",
			Fix:         "Store sensitive secrets in environment variables or a secret vault.",
		},
		{
			Pattern:     `(?i)(SELECT|INSERT|UPDATE|DELETE).*\+.*`,
			RuleID:      "SEC-002",
			Category:    "SQL injection",
			Severity:    "CRITICAL",
			Title:       "Potential SQL Injection via String Concatenation",
			Description: "Dynamic SQL query constructed by string concatenation instead of parameterized placeholders.",
			Impact:      "Malicious inputs can alter query logic and compromise database integrity.",
			Fix:         "Use parameterized queries / prepared statements.",
		},
		{
			Pattern:     `(?i)exec\.Command\([^,)]*\+`,
			RuleID:      "SEC-003",
			Category:    "Command injection",
			Severity:    "CRITICAL",
			Title:       "Potential Command Injection",
			Description: "Executing OS commands constructed from concatenated external inputs.",
			Impact:      "Unsanitized input can allow arbitrary system command execution.",
			Fix:         "Pass command arguments as separate sanitized array elements.",
		},
		{
			Pattern:     `(?i)(eval\(|dangerouslySetInnerHTML)`,
			RuleID:      "SEC-004",
			Category:    "XSS",
			Severity:    "HIGH",
			Title:       "Unsafe Code Execution or XSS Vector",
			Description: "Direct usage of eval() or unsanitized innerHTML rendering.",
			Impact:      "Cross-site scripting (XSS) or arbitrary code execution in client browser.",
			Fix:         "Sanitize dynamic content using safe DOM rendering methods.",
		},
		{
			Pattern:     `(?i)http\.Get\(.*r\.URL\.Query`,
			RuleID:      "SEC-005",
			Category:    "SSRF",
			Severity:    "HIGH",
			Title:       "Potential Server-Side Request Forgery (SSRF)",
			Description: "Making outbound HTTP requests directly to URLs provided in request parameters.",
			Impact:      "Allows attackers to scan internal infrastructure and bypass firewall boundaries.",
			Fix:         "Validate and sanitize destination URLs against an explicit allowlist.",
		},
	}

	// Apply builtin rules and custom rules across repository files
	gitCmd := exec.Command("git", "ls-files")
	gitCmd.Dir = repoPath
	gitOut, err := gitCmd.Output()
	if err != nil {
		return findings, nil
	}

	files := strings.Split(string(gitOut), "\n")
	for _, file := range files {
		file = strings.TrimSpace(file)
		if file == "" || strings.HasPrefix(file, ".") {
			continue
		}

		_ = filepath.Join(repoPath, file)
		cmd := exec.Command("git", "show", fmt.Sprintf("HEAD:%s", file))
		cmd.Dir = repoPath
		contentBytes, err := cmd.Output()
		if err != nil {
			continue
		}

		lines := strings.Split(string(contentBytes), "\n")
		for lineIdx, lineStr := range lines {
			lineNum := lineIdx + 1

			// Builtin rules matching
			for _, rule := range builtinRules {
				re, err := regexp.Compile(rule.Pattern)
				if err != nil {
					continue
				}
				if re.MatchString(lineStr) {
					snippet := strings.TrimSpace(lineStr)
					findings = append(findings, RawOCRFinding{
						File:        file,
						Line:        lineNum,
						EndLine:     lineNum,
						RuleID:      rule.RuleID,
						Category:    rule.Category,
						Severity:    rule.Severity,
						Confidence:  "HIGH",
						Title:       rule.Title,
						Description: rule.Description,
						Impact:      rule.Impact,
						SuggestedFix: rule.Fix,
						CodeSnippet: snippet,
					})
				}
			}

			// Custom rules matching
			for _, cRule := range customRules {
				re, err := regexp.Compile(cRule.Pattern)
				if err != nil {
					continue
				}
				if re.MatchString(lineStr) {
					snippet := strings.TrimSpace(lineStr)
					findings = append(findings, RawOCRFinding{
						File:        file,
						Line:        lineNum,
						EndLine:     lineNum,
						RuleID:      cRule.RuleID,
						Category:    cRule.Category,
						Severity:    cRule.Severity,
						Confidence:  "HIGH",
						Title:       cRule.Name,
						Description: cRule.Message,
						Impact:      "Policy violation based on custom repository rules.",
						SuggestedFix: "Update code to adhere to repository policy.",
						CodeSnippet: snippet,
					})
				}
			}
		}
	}

	return findings, nil
}

// GenerateFingerprint creates a stable hash based on file + ruleID + normalized code snippet context
func GenerateFingerprint(file, ruleID, codeSnippet string) string {
	normalizedSnippet := strings.ToLower(strings.Join(strings.Fields(codeSnippet), " "))
	data := fmt.Sprintf("%s|%s|%s", file, ruleID, normalizedSnippet)
	hash := sha256.Sum256([]byte(data))
	return hex.EncodeToString(hash[:16])
}

// NormalizeFindings maps raw findings into structured CodeSentinel Finding domain models
func NormalizeFindings(reviewID, repoID, prID, commitSHA, branch string, raw []RawOCRFinding) []models.Finding {
	now := time.Now()
	var findings []models.Finding

	for _, r := range raw {
		fp := GenerateFingerprint(r.File, r.RuleID, r.CodeSnippet)
		f := models.Finding{
			ID:            fmt.Sprintf("fnd-%s-%d", fp[:8], now.UnixNano()%10000),
			ReviewID:      reviewID,
			RepositoryID:  repoID,
			PullRequestID: prID,
			CommitSHA:     commitSHA,
			Branch:        branch,
			File:          r.File,
			Line:          r.Line,
			EndLine:       r.EndLine,
			RuleID:        r.RuleID,
			Category:      r.Category,
			Severity:      strings.ToUpper(r.Severity),
			Confidence:    strings.ToUpper(r.Confidence),
			Title:         r.Title,
			Description:   r.Description,
			Impact:        r.Impact,
			SuggestedFix:  r.SuggestedFix,
			Fingerprint:   fp,
			FirstSeen:     now,
			LastSeen:      now,
			Status:        "Open",
			IsNew:         true,
			CodeSnippet:   r.CodeSnippet,
			CreatedAt:     now,
		}
		findings = append(findings, f)
	}

	return findings
}
