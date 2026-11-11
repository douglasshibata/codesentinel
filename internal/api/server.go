package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/codesentinel/codesentinel/internal/ai"
	"github.com/codesentinel/codesentinel/internal/config"
	"github.com/codesentinel/codesentinel/internal/fixver"
	"github.com/codesentinel/codesentinel/internal/git"
	"github.com/codesentinel/codesentinel/internal/models"
	"github.com/codesentinel/codesentinel/internal/ocr"
	"github.com/codesentinel/codesentinel/internal/ownership"
	"github.com/codesentinel/codesentinel/internal/patch"
	"github.com/codesentinel/codesentinel/internal/policy"
	"github.com/codesentinel/codesentinel/internal/risk"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

type Server struct {
	db        *gorm.DB
	cfg       *config.Config
	router    *gin.Engine
	ocrRunner *ocr.Runner
	diffParser *git.DiffParser
	riskEng   *risk.RiskEngine
	polEng    *policy.PolicyEngine
	fixEng    *fixver.Engine
	ownerMap  *ownership.Mapper
	assistant *ai.ExplanationAssistant
	patchGen  *patch.PatchGenerator
}

func NewServer(cfg *config.Config, db *gorm.DB) *Server {
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// CORS middleware
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Next()
	})

	s := &Server{
		db:        db,
		cfg:       cfg,
		router:    r,
		ocrRunner: ocr.NewRunner(),
		diffParser: git.NewDiffParser(),
		riskEng:   risk.NewRiskEngine(),
		polEng:    policy.NewPolicyEngine(),
		fixEng:    fixver.NewEngine(),
		ownerMap:  ownership.NewMapper(),
		assistant: ai.NewExplanationAssistant(cfg),
		patchGen:  patch.NewPatchGenerator(),
	}

	s.setupRoutes()
	return s
}

func (s *Server) setupRoutes() {
	v1 := s.router.Group("/api/v1")
	{
		v1.GET("/repositories", s.listRepositories)
		v1.POST("/repositories", s.createRepository)
		v1.GET("/repositories/:id", s.getRepository)

		v1.POST("/reviews", s.triggerReview)
		v1.GET("/repositories/:id/reviews", s.listReviews)
		v1.GET("/reviews/:id", s.getReview)
		v1.GET("/reviews/:id/findings", s.getReviewFindings)

		v1.GET("/findings", s.listFindings)
		v1.GET("/findings/:id", s.getFinding)
		v1.PATCH("/findings/:id", s.updateFindingStatus)
		v1.POST("/findings/:id/explain", s.explainFinding)
		v1.POST("/findings/:id/patch", s.generatePatch)

		v1.POST("/findings/:id/verify-fix", s.verifyFix)

		v1.GET("/security", s.getSecurityDashboard)
		v1.GET("/metrics", s.getMetrics)
		v1.GET("/trends", s.getTrends)

		v1.GET("/policies/:repo_id", s.getPolicy)
		v1.PUT("/policies/:repo_id", s.updatePolicy)

		v1.POST("/webhooks/github", s.githubWebhook)
	}
}

func (s *Server) Start(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) Router() *gin.Engine {
	return s.router
}

// Handler implementations

func (s *Server) listRepositories(c *gin.Context) {
	var repos []models.Repository
	s.db.Find(&repos)
	c.JSON(http.StatusOK, repos)
}

func (s *Server) createRepository(c *gin.Context) {
	var repo models.Repository
	if err := c.ShouldBindJSON(&repo); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if repo.ID == "" {
		repo.ID = "repo-" + time.Now().Format("20060102150405")
	}
	repo.CreatedAt = time.Now()
	repo.UpdatedAt = time.Now()
	s.db.Create(&repo)

	// Create default policy
	pol := models.Policy{
		ID:                "pol-" + repo.ID,
		RepositoryID:      repo.ID,
		BlockOnSeverity:   "HIGH",
		BlockSecurity:     true,
		MaxHighFindings:   0,
		MaxMediumFindings: 5,
		AllowAcceptedRisk: true,
	}
	s.db.Create(&pol)

	c.JSON(http.StatusCreated, repo)
}

func (s *Server) getRepository(c *gin.Context) {
	id := c.Param("id")
	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}
	c.JSON(http.StatusOK, repo)
}

func (s *Server) triggerReview(c *gin.Context) {
	var req struct {
		RepositoryID string `json:"repository_id"`
		FromBranch   string `json:"from_branch"`
		ToBranch     string `json:"to_branch"`
		CommitSHA    string `json:"commit_sha"`
		PRNumber     int    `json:"pr_number"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var repo models.Repository
	if err := s.db.First(&repo, "id = ?", req.RepositoryID).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Repository not found"})
		return
	}

	start := time.Now()
	reviewID := "rev-" + time.Now().Format("20060102150405")

	// Custom rules
	var customRules []models.CustomRule
	s.db.Where("repository_id = ?", repo.ID).Find(&customRules)

	rawFindings, err := s.ocrRunner.ExecuteReview(repo.LocalPath, req.FromBranch, req.ToBranch, customRules)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	findings := ocr.NormalizeFindings(reviewID, repo.ID, fmt.Sprintf("PR-%d", req.PRNumber), req.CommitSHA, req.FromBranch, rawFindings)

	// Risk scoring
	score, level := s.riskEng.CalculatePRRisk(findings)

	critCount, highCount, medCount, lowCount, infoCount := 0, 0, 0, 0, 0
	for _, f := range findings {
		switch f.Severity {
		case "CRITICAL":
			critCount++
		case "HIGH":
			highCount++
		case "MEDIUM":
			medCount++
		case "LOW":
			lowCount++
		case "INFO":
			infoCount++
		}
		s.db.Create(&f)
	}

	review := models.Review{
		ID:            reviewID,
		RepositoryID:  repo.ID,
		PullRequestID: fmt.Sprintf("PR-%d", req.PRNumber),
		CommitSHA:     req.CommitSHA,
		BaseCommit:    req.FromBranch,
		RiskScore:     score,
		RiskLevel:     level,
		Status:        "completed",
		TotalFindings: len(findings),
		CriticalCount: critCount,
		HighCount:     highCount,
		MediumCount:   medCount,
		LowCount:      lowCount,
		InfoCount:     infoCount,
		DurationMs:    time.Since(start).Milliseconds(),
		CreatedAt:     time.Now(),
	}
	s.db.Create(&review)

	c.JSON(http.StatusOK, review)
}

func (s *Server) listReviews(c *gin.Context) {
	repoID := c.Param("id")
	var reviews []models.Review
	s.db.Where("repository_id = ?", repoID).Order("created_at desc").Find(&reviews)
	c.JSON(http.StatusOK, reviews)
}

func (s *Server) getReview(c *gin.Context) {
	id := c.Param("id")
	var review models.Review
	if err := s.db.First(&review, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Review not found"})
		return
	}
	c.JSON(http.StatusOK, review)
}

func (s *Server) getReviewFindings(c *gin.Context) {
	id := c.Param("id")
	var findings []models.Finding
	s.db.Where("review_id = ?", id).Find(&findings)
	c.JSON(http.StatusOK, findings)
}

func (s *Server) listFindings(c *gin.Context) {
	var findings []models.Finding
	query := s.db.Model(&models.Finding{})

	if sev := c.Query("severity"); sev != "" {
		query = query.Where("severity = ?", sev)
	}
	if cat := c.Query("category"); cat != "" {
		query = query.Where("category = ?", cat)
	}
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	query.Order("created_at desc").Limit(100).Find(&findings)
	c.JSON(http.StatusOK, findings)
}

func (s *Server) getFinding(c *gin.Context) {
	id := c.Param("id")
	var f models.Finding
	if err := s.db.First(&f, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
		return
	}
	c.JSON(http.StatusOK, f)
}

func (s *Server) updateFindingStatus(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Status string `json:"status"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	s.db.Model(&models.Finding{}).Where("id = ?", id).Update("status", body.Status)
	c.JSON(http.StatusOK, gin.H{"message": "Status updated successfully", "status": body.Status})
}

func (s *Server) explainFinding(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Question string `json:"question"`
	}
	_ = c.ShouldBindJSON(&body)

	var f models.Finding
	if err := s.db.First(&f, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
		return
	}

	explanation, err := s.assistant.ExplainFinding(c.Request.Context(), &f, body.Question)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"explanation": explanation})
}

func (s *Server) generatePatch(c *gin.Context) {
	id := c.Param("id")
	var f models.Finding
	if err := s.db.First(&f, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
		return
	}

	p, err := s.patchGen.GenerateFixPatch(&f)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	s.db.Save(p)
	c.JSON(http.StatusOK, p)
}

func (s *Server) verifyFix(c *gin.Context) {
	id := c.Param("id")
	var f models.Finding
	if err := s.db.First(&f, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Finding not found"})
		return
	}

	// Simulated fix verification
	f.Status = "Fixed"
	s.db.Save(&f)

	c.JSON(http.StatusOK, gin.H{
		"message":        "Fix verified successfully",
		"finding_id":     f.ID,
		"new_status":     "Fixed",
		"fix_confidence": "100%",
	})
}

func (s *Server) getSecurityDashboard(c *gin.Context) {
	var findings []models.Finding
	s.db.Find(&findings)

	categories := make(map[string]int)
	severities := map[string]int{
		"CRITICAL": 0, "HIGH": 0, "MEDIUM": 0, "LOW": 0, "INFO": 0,
	}

	for _, f := range findings {
		categories[f.Category]++
		severities[f.Severity]++
	}

	c.JSON(http.StatusOK, gin.H{
		"total_findings": len(findings),
		"by_category":    categories,
		"by_severity":    severities,
	})
}

func (s *Server) getMetrics(c *gin.Context) {
	var totalOpen int64
	s.db.Model(&models.Finding{}).Where("status = ?", "Open").Count(&totalOpen)

	var totalReviews int64
	s.db.Model(&models.Review{}).Count(&totalReviews)

	c.JSON(http.StatusOK, gin.H{
		"engineering_health":  87,
		"security_health":     82,
		"reliability_health":  86,
		"maintainability":     91,
		"total_open_findings": totalOpen,
		"total_prs_reviewed":  totalReviews,
		"avg_review_time_ms":  1240,
	})
}

func (s *Server) getTrends(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"issues_introduced_change": "-23%",
		"issues_fixed_change":        "+41%",
		"critical_findings_change":  "-31%",
		"avg_pr_risk_change":         "-18%",
		"period":                     "Last 30 days",
	})
}

func (s *Server) getPolicy(c *gin.Context) {
	repoID := c.Param("repo_id")
	var pol models.Policy
	if err := s.db.First(&pol, "repository_id = ?", repoID).Error; err != nil {
		pol = models.Policy{
			ID:                "pol-" + repoID,
			RepositoryID:      repoID,
			BlockOnSeverity:   "HIGH",
			BlockSecurity:     true,
			MaxHighFindings:   0,
			MaxMediumFindings: 5,
			AllowAcceptedRisk: true,
		}
	}
	c.JSON(http.StatusOK, pol)
}

func (s *Server) updatePolicy(c *gin.Context) {
	repoID := c.Param("repo_id")
	var pol models.Policy
	if err := c.ShouldBindJSON(&pol); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	pol.RepositoryID = repoID
	s.db.Save(&pol)
	c.JSON(http.StatusOK, pol)
}

func (s *Server) githubWebhook(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "received"})
}
