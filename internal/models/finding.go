package models

import "time"

// Finding represents a single code quality or security issue (SARIF & relacional compliant).
type Finding struct {
	ID                       string    `json:"id" gorm:"primaryKey"`
	ReviewID                 string    `json:"review_id" gorm:"index"`
	RepositoryID             string    `json:"repository_id" gorm:"index"`
	PullRequestID            string    `json:"pull_request_id" gorm:"index"`
	CommitSHA                string    `json:"commit_sha"`
	Branch                   string    `json:"branch"`
	File                     string    `json:"file"`
	Line                     int       `json:"line"`
	EndLine                  int       `json:"end_line"`
	RuleID                   string    `json:"rule_id"`
	Category                 string    `json:"category"` // "SQL injection", "XSS", "Secrets", etc.
	Severity                 string    `json:"severity"` // "CRITICAL", "HIGH", "MEDIUM", "LOW", "INFO"
	Confidence               string    `json:"confidence"` // "HIGH", "MEDIUM", "LOW"
	ConfidenceMultiplier     float64   `json:"confidence_multiplier"`
	IsSecuritySensitiveScope bool      `json:"is_security_sensitive_scope"`
	Title                    string    `json:"title"`
	Description              string    `json:"description"`
	Impact                   string    `json:"impact"`
	SuggestedFix             string    `json:"suggested_fix"`
	Fingerprint              string    `json:"fingerprint" gorm:"index"`
	FirstSeen                time.Time `json:"first_seen"`
	LastSeen                 time.Time `json:"last_seen"`
	Status                   string    `json:"status"` // "Open", "Confirmed", "False Positive", "Fixed", "Ignored", "Accepted Risk"
	IsNew                    bool      `json:"is_new"`
	Author                   string    `json:"author"`
	Reviewer                 string    `json:"reviewer"`
	CodeSnippet              string    `json:"code_snippet"`
	CreatedAt                time.Time `json:"created_at"`
}
