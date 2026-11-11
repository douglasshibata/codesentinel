package config

import (
	"os"
	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerPort string `yaml:"server_port" json:"server_port"`
	DBType     string `yaml:"db_type" json:"db_type"`     // "sqlite" or "postgres"
	DBPath     string `yaml:"db_path" json:"db_path"`     // For sqlite
	DSN        string `yaml:"dsn" json:"dsn"`             // For postgres
	PrivacyMode bool  `yaml:"privacy_mode" json:"privacy_mode"`
	LLMProvider string `yaml:"llm_provider" json:"llm_provider"` // "openai", "anthropic", "ollama", "mock"
	LLMAPIKey   string `yaml:"llm_api_key" json:"llm_api_key"`
	LLMModel    string `yaml:"llm_model" json:"llm_model"`
	LLMBaseURL  string `yaml:"llm_base_url" json:"llm_base_url"`
}

func DefaultConfig() *Config {
	return &Config{
		ServerPort:  "8080",
		DBType:      "sqlite",
		DBPath:      "codesentinel.db",
		DSN:         "",
		PrivacyMode: false,
		LLMProvider: "mock",
		LLMModel:    "gpt-4o",
	}
}

func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()
	if path == "" {
		return cfg, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return cfg, nil
	}
	err = yaml.Unmarshal(data, cfg)
	if err != nil {
		return nil, err
	}
	return cfg, nil
}
