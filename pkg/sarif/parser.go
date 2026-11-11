package sarif

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/codesentinel/codesentinel/internal/ast"
	"github.com/codesentinel/codesentinel/internal/models"
)

// OASIS SARIF v2.1.0 data models
type SARIFReport struct {
	Version string     `json:"version"`
	Schema  string     `json:"$schema"`
	Runs    []SARIFRun `json:"runs"`
}

type SARIFRun struct {
	Tool    SARIFTool     `json:"tool"`
	Results []SARIFResult `json:"results"`
}

type SARIFTool struct {
	Driver SARIFDriver `json:"driver"`
}

type SARIFDriver struct {
	Name           string      `json:"name"`
	Version        string      `json:"version"`
	InformationURI string      `json:"informationUri"`
	Rules          []SARIFRule `json:"rules"`
}

type SARIFRule struct {
	ID               string               `json:"id"`
	Name             string               `json:"name"`
	ShortDescription SARIFMultiformatText `json:"shortDescription"`
	FullDescription  SARIFMultiformatText `json:"fullDescription"`
	DefaultConfiguration SARIFRuleConfig `json:"defaultConfiguration"`
}

type SARIFRuleConfig struct {
	Level string `json:"level"` // "error", "warning", "note", "none"
}

type SARIFMultiformatText struct {
	Text string `json:"text"`
}

type SARIFResult struct {
	RuleID    string               `json:"ruleId"`
	Level     string               `json:"level"` // "error", "warning", "note"
	Message   SARIFMultiformatText `json:"message"`
	Locations []SARIFLocation      `json:"locations"`
}

type SARIFLocation struct {
	PhysicalLocation SARIFPhysicalLocation `json:"physicalLocation"`
}

type SARIFPhysicalLocation struct {
	ArtifactLocation SARIFArtifactLocation `json:"artifactLocation"`
	Region           SARIFRegion           `json:"region"`
}

type SARIFArtifactLocation struct {
	URI string `json:"uri"`
}

type SARIFRegion struct {
	StartLine   int          `json:"startLine"`
	EndLine     int          `json:"endLine"`
	Snippet     SARIFSnippet `json:"snippet"`
}

type SARIFSnippet struct {
	Text string `json:"text"`
}

// ParseSARIFFile parses an OASIS SARIF v2.1.0 file from disk
func ParseSARIFFile(filePath string) (*SARIFReport, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SARIF file: %w", err)
	}
	defer file.Close()

	return ParseSARIFReader(file)
}

// ParseSARIFReader parses an OASIS SARIF v2.1.0 report from an io.Reader
func ParseSARIFReader(r io.Reader) (*SARIFReport, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read SARIF data: %w", err)
	}

	var report SARIFReport
	if err := json.Unmarshal(data, &report); err != nil {
		return nil, fmt.Errorf("failed to parse SARIF JSON: %w", err)
	}

	return &report, nil
}

// NormalizeSARIFToFindings converts SARIF results to canonical CodeSentinel Finding models
func NormalizeSARIFToFindings(report *SARIFReport, reviewID, repoID, prID, commitSHA string) []models.Finding {
	now := time.Now()
	var findings []models.Finding

	if report == nil {
		return findings
	}

	for _, run := range report.Runs {
		// Map rule IDs to default severity if specified
		ruleSeverityMap := make(map[string]string)
		for _, r := range run.Tool.Driver.Rules {
			ruleSeverityMap[r.ID] = mapLevelToSeverity(r.DefaultConfiguration.Level)
		}

		for _, res := range run.Results {
			uri := ""
			startLine := 1
			endLine := 1
			snippet := ""

			if len(res.Locations) > 0 {
				loc := res.Locations[0].PhysicalLocation
				uri = loc.ArtifactLocation.URI
				startLine = loc.Region.StartLine
				endLine = loc.Region.EndLine
				if endLine < startLine {
					endLine = startLine
				}
				snippet = loc.Region.Snippet.Text
			}

			severity := mapLevelToSeverity(res.Level)
			if severity == models.SeverityLow && ruleSeverityMap[res.RuleID] != "" {
				severity = ruleSeverityMap[res.RuleID]
			}

			isSecScope := isSecurityPath(uri)
			fp := ast.GenerateStructuralFingerprint(res.RuleID, uri, "scope", snippet)

			f := models.Finding{
				ID:                       fmt.Sprintf("sarif-%s-%d", fp[:8], now.UnixNano()%10000),
				ReviewID:                 reviewID,
				RepositoryID:             repoID,
				PullRequestID:            prID,
				CommitSHA:                commitSHA,
				File:                     uri,
				Line:                     startLine,
				EndLine:                  endLine,
				RuleID:                   res.RuleID,
				Category:                 "Static Analysis",
				Severity:                 severity,
				Confidence:               "HIGH",
				ConfidenceMultiplier:     1.0,
				IsSecuritySensitiveScope: isSecScope,
				Title:                    res.Message.Text,
				Description:              res.Message.Text,
				Impact:                   "Static vulnerability / code quality finding reported by " + run.Tool.Driver.Name,
				SuggestedFix:             "Review and patch code corresponding to SARIF rule " + res.RuleID,
				Fingerprint:              fp,
				FirstSeen:                now,
				LastSeen:                 now,
				Status:                   models.StatusOpen,
				IsNew:                    true,
				CodeSnippet:              snippet,
				CreatedAt:                now,
			}

			findings = append(findings, f)
		}
	}

	return findings
}

func mapLevelToSeverity(level string) string {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "error":
		return models.SeverityCritical
	case "warning":
		return models.SeverityHigh
	case "note":
		return models.SeverityMedium
	case "none":
		return models.SeverityLow
	default:
		return models.SeverityMedium
	}
}

func isSecurityPath(path string) bool {
	p := strings.ToLower(path)
	if strings.Contains(p, "security") || strings.Contains(p, "auth") || strings.HasSuffix(p, ".tf") || strings.Contains(p, "dockerfile") {
		return true
	}
	return false
}
