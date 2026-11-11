package fixver

import (
	"fmt"

	"github.com/codesentinel/codesentinel/internal/models"
)

type FixVerificationResult struct {
	PreviousCount int              `json:"previous_count"`
	CurrentCount  int              `json:"current_count"`
	FixedCount    int              `json:"fixed_count"`
	RemainingCount int             `json:"remaining_count"`
	ReintroducedCount int          `json:"reintroduced_count"`
	FixConfidence string           `json:"fix_confidence"` // "100%", "HIGH", "MEDIUM", "LOW"
	FixedFindings []models.Finding `json:"fixed_findings"`
	RemainingFindings []models.Finding `json:"remaining_findings"`
	ReintroducedFindings []models.Finding `json:"reintroduced_findings"`
}

type Engine struct{}

func NewEngine() *Engine {
	return &Engine{}
}

// CompareReviews compares findings from previous review vs current review by fingerprint
func (e *Engine) CompareReviews(previousFindings, currentFindings []models.Finding) FixVerificationResult {
	prevMap := make(map[string]models.Finding)
	for _, pf := range previousFindings {
		prevMap[pf.Fingerprint] = pf
	}

	currMap := make(map[string]models.Finding)
	for _, cf := range currentFindings {
		currMap[cf.Fingerprint] = cf
	}

	var fixed []models.Finding
	var remaining []models.Finding
	var reintroduced []models.Finding

	// Check fixed issues
	for fp, pf := range prevMap {
		if pf.Status == "Fixed" {
			// Check if reintroduced
			if _, exists := currMap[fp]; exists {
				cf := currMap[fp]
				cf.Status = "Open"
				cf.IsNew = true
				reintroduced = append(reintroduced, cf)
			}
			continue
		}

		if _, exists := currMap[fp]; !exists {
			// Finding no longer present in current review -> Fixed!
			pf.Status = "Fixed"
			fixed = append(fixed, pf)
		}
	}

	// Check current active issues
	for fp, cf := range currMap {
		if pf, exists := prevMap[fp]; exists {
			cf.FirstSeen = pf.FirstSeen
			cf.IsNew = false
			cf.Status = pf.Status
			remaining = append(remaining, cf)
		} else {
			cf.IsNew = true
			cf.Status = "Open"
			remaining = append(remaining, cf)
		}
	}

	totalFixed := len(fixed)
	totalPrev := len(previousFindings)
	confidence := "HIGH"
	if totalPrev > 0 {
		pct := (float64(totalFixed) / float64(totalPrev)) * 100.0
		if pct == 100.0 {
			confidence = "100% (COMPLETE)"
		} else if pct >= 50.0 {
			confidence = "HIGH"
		} else if pct > 0 {
			confidence = "MEDIUM"
		} else {
			confidence = "LOW"
		}
	} else {
		confidence = "N/A"
	}

	return FixVerificationResult{
		PreviousCount:     totalPrev,
		CurrentCount:      len(currentFindings),
		FixedCount:        totalFixed,
		RemainingCount:    len(remaining),
		ReintroducedCount: len(reintroduced),
		FixConfidence:     confidence,
		FixedFindings:     fixed,
		RemainingFindings: remaining,
		ReintroducedFindings: reintroduced,
	}
}

func (e *Engine) FormatSummary(res FixVerificationResult) string {
	return fmt.Sprintf("BEFORE: %d issues | AFTER: %d issues | Fixed: %d | Remaining: %d | Reintroduced: %d | Fix Confidence: %s",
		res.PreviousCount, res.CurrentCount, res.FixedCount, res.RemainingCount, res.ReintroducedCount, res.FixConfidence)
}
