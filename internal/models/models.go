package models

import (
	"time"
)

// Organization represents a multi-tenant boundary.
type Organization struct {
	ID        string    `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug" gorm:"uniqueIndex"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// User represents a system user.
type User struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organization_id" gorm:"index"`
	Email          string    `json:"email" gorm:"uniqueIndex"`
	Name           string    `json:"name"`
	Role           string    `json:"role"` // "admin", "developer", "viewer"
	CreatedAt      time.Time `json:"created_at"`
}

// Repository represents a connected Git repo (Local, GitHub, GitLab, Bitbucket).
type Repository struct {
	ID             string    `json:"id" gorm:"primaryKey"`
	OrganizationID string    `json:"organization_id" gorm:"index"`
	Name           string    `json:"name"`
	Type           string    `json:"type"` // "local", "github", "gitlab", "bitbucket"
	URL            string    `json:"url"`
	LocalPath      string    `json:"local_path"`
	DefaultBranch  string    `json:"default_branch"`
	Profile        string    `json:"profile"` // "general", "security", "strict", etc.
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
}

// PullRequest tracks an open PR or MR.
type PullRequest struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	RepositoryID string    `json:"repository_id" gorm:"index"`
	Number       int       `json:"number"`
	Title        string    `json:"title"`
	Author       string    `json:"author"`
	SourceBranch string    `json:"source_branch"`
	TargetBranch string    `json:"target_branch"`
	Status       string    `json:"status"` // "open", "merged", "closed"
	RiskScore    int       `json:"risk_score"`
	RiskLevel    string    `json:"risk_level"` // "LOW", "MODERATE", "HIGH", "CRITICAL"
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

// Commit represents a Git commit reviewed.
type Commit struct {
	SHA          string    `json:"sha" gorm:"primaryKey"`
	RepositoryID string    `json:"repository_id" gorm:"index"`
	Author       string    `json:"author"`
	Message      string    `json:"message"`
	CommittedAt  time.Time `json:"committed_at"`
}

// Review represents an automated review execution.
type Review struct {
	ID            string    `json:"id" gorm:"primaryKey"`
	RepositoryID  string    `json:"repository_id" gorm:"index"`
	PullRequestID string    `json:"pull_request_id" gorm:"index"`
	CommitSHA     string    `json:"commit_sha"`
	BaseCommit    string    `json:"base_commit"`
	RiskScore     int       `json:"risk_score"`
	RiskLevel     string    `json:"risk_level"`
	Status        string    `json:"status"` // "completed", "failed"
	Error         string    `json:"error,omitempty"`
	TotalFindings int       `json:"total_findings"`
	CriticalCount int       `json:"critical_count"`
	HighCount     int       `json:"high_count"`
	MediumCount   int       `json:"medium_count"`
	LowCount      int       `json:"low_count"`
	InfoCount     int       `json:"info_count"`
	DurationMs    int64     `json:"duration_ms"`
	CreatedAt     time.Time `json:"created_at"`
}

// Policy defines automated quality gate rules.
type Policy struct {
	ID                string `json:"id" gorm:"primaryKey"`
	RepositoryID      string `json:"repository_id" gorm:"uniqueIndex"`
	BlockOnSeverity   string `json:"block_on_severity"` // "CRITICAL", "HIGH", etc.
	BlockSecurity     bool   `json:"block_security"`
	MaxHighFindings   int    `json:"max_high_findings"`
	MaxMediumFindings int    `json:"max_medium_findings"`
	AllowAcceptedRisk bool   `json:"allow_accepted_risk"`
}

// CustomRule represents a repository-specific static check rule.
type CustomRule struct {
	ID           string    `json:"id" gorm:"primaryKey"`
	RepositoryID string    `json:"repository_id" gorm:"index"`
	RuleID       string    `json:"rule_id"`
	Name         string    `json:"name"`
	Pattern      string    `json:"pattern"` // Regex or keyword
	Category     string    `json:"category"`
	Severity     string    `json:"severity"`
	Message      string    `json:"message"`
	CreatedAt    time.Time `json:"created_at"`
}

// Patch represents an AI-generated diff proposal.
type Patch struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	FindingID   string    `json:"finding_id" gorm:"index"`
	Original    string    `json:"original"`
	Proposed    string    `json:"proposed"`
	Diff        string    `json:"diff"`
	Explanation string    `json:"explanation"`
	CreatedAt   time.Time `json:"created_at"`
}

// Metric represents system-wide or repo-specific aggregated health stats.
type Metric struct {
	ID                   string    `json:"id" gorm:"primaryKey"`
	RepositoryID         string    `json:"repository_id" gorm:"index"`
	HealthScore          int       `json:"health_score"`
	SecurityScore        int       `json:"security_score"`
	ReliabilityScore     int       `json:"reliability_score"`
	MaintainabilityScore int       `json:"maintainability_score"`
	TotalOpenFindings    int       `json:"total_open_findings"`
	MeanTimeToFixHours   float64   `json:"mean_time_to_fix_hours"`
	FalsePositiveRate    float64   `json:"false_positive_rate"`
	RecordedAt           time.Time `json:"recorded_at"`
}
