package models

const (
	SeverityCritical = "CRITICAL"
	SeverityHigh     = "HIGH"
	SeverityMedium   = "MEDIUM"
	SeverityLow      = "LOW"
	SeverityInfo     = "INFO"

	StatusOpen         = "Open"
	StatusConfirmed    = "Confirmed"
	StatusFalsePositive = "False Positive"
	StatusFixed        = "Fixed"
	StatusIgnored      = "Ignored"
	StatusAcceptedRisk = "Accepted Risk"
)

// PolicyConfig represents configuration thresholds and gate policy rules
type PolicyConfig struct {
	FailFast               bool    `yaml:"fail_fast" json:"fail_fast"`
	BlockOnCritical        bool    `yaml:"block_on_critical" json:"block_on_critical"`
	MaxHighFindings        int     `yaml:"max_high_findings" json:"max_high_findings"`
	MaxMediumFindings      int     `yaml:"max_medium_findings" json:"max_medium_findings"`
	MaxRiskScoreThreshold  float64 `yaml:"max_risk_score_threshold" json:"max_risk_score_threshold"`
	AllowAcceptedRisk      bool    `yaml:"allow_accepted_risk" json:"allow_accepted_risk"`
}

// ScannersConfig represents scanner configurations for SARIF inputs
type ScannersConfig struct {
	SARIFInputs []string `yaml:"sarif_inputs" json:"sarif_inputs"`
}

// FullConfig mirrors the production corporate codesentinel.yaml schema
type FullConfig struct {
	Version  string `yaml:"version" json:"version"`
	Server   struct {
		Host            string `yaml:"host" json:"host"`
		Port            int    `yaml:"port" json:"port"`
		ReadTimeoutSec  int    `yaml:"read_timeout_sec" json:"read_timeout_sec"`
		WriteTimeoutSec int    `yaml:"write_timeout_sec" json:"write_timeout_sec"`
	} `yaml:"server"`
	Database struct {
		Type         string `yaml:"type" json:"type"`
		DSN          string `yaml:"dsn" json:"dsn"`
		MaxOpenConns int    `yaml:"max_open_conns" json:"max_open_conns"`
		MaxIdleConns int    `yaml:"max_idle_conns" json:"max_idle_conns"`
	} `yaml:"database"`
	Privacy struct {
		StrictRedaction        bool     `yaml:"strict_redaction" json:"strict_redaction"`
		EnableExternalLLM      bool     `yaml:"enable_external_llm" json:"enable_external_llm"`
		AllowedRemoteProviders []string `yaml:"allowed_remote_providers" json:"allowed_remote_providers"`
	} `yaml:"privacy"`
	LLMGateway struct {
		Provider    string  `yaml:"provider" json:"provider"`
		Model       string  `yaml:"model" json:"model"`
		Temperature float64 `yaml:"temperature" json:"temperature"`
		MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`
		TimeoutSec  int     `yaml:"timeout_sec" json:"timeout_sec"`
	} `yaml:"llm_gateway"`
	Policy   PolicyConfig   `yaml:"policy" json:"policy"`
	Scanners ScannersConfig `yaml:"scanners" json:"scanners"`
}
