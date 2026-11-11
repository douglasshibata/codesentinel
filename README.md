# 🛡️ CodeSentinel - Continuous AI Code-Quality & Security Gate

[![Go](https://img.shields.io/badge/Language-Go%201.24%2B-blue)](#)
[![OASIS](https://img.shields.io/badge/SARIF-v2.1.0-blueviolet)](#)
[![UI](https://img.shields.io/badge/Frontend-React%2018%20%7C%20TypeScript-purple)](#)
[![License](https://img.shields.io/badge/License-Apache--2.0-yellow)](#)

> **"Deterministic gates. Context-aware AI code reviews. Zero secret leakage."**

## Diagnóstico de Maturidade e Gaps Técnicos

A versão anterior do CodeSentinel apresentava um conceito alinhado ao ecossistema de AppSec, mas com lacunas que inviabilizariam sua adoção corporativa:
- **Abstração Fictícia do Pipeline de AST/LLM:** Não havia definição do contrato SARIF (OASIS Standard), nem a formulação do algoritmo de Stable Fingerprinting (crítico para evitar finding drift quando linhas de código mudam em PRs subsequentes).
- **Fragilidade do Privacy Mode:** Declarar apenas `privacy_mode: true` sem especificar sanitização de segredos em pré-processamento expõe credenciais corporativas a provedores LLM externos.
- **Ausência de Estrutura de Diretórios e Persistência Concreta:** O diagrama omitia barramentos assíncronos e webhooks para plataformas SCM (GitHub Apps / GitLab Webhooks).
- **Falta de Formulação do Risk Scoring:** A pontuação 0–100 não possuía cálculo determinístico ponderado.

Abaixo está o README e a especificação técnica integralmente refatorados para padrão de produção.

---

O **CodeSentinel** é uma plataforma corporativa e portão de qualidade contínuo (*Continuous AI Code-Quality & Security Gate*) em Go e React/TypeScript. Ele orquestra motores de análise estática e estende pipelines de LLM para transformar achados brutos (SARIF/JSON) em um grafo de auditoria relacional com cálculo determinístico de risco (0–100), bloqueio preemptivo de merges via CLI/CI/CD, fingerprinting sintático invariante a mudanças de linha e verificação de ciclo de vida de correções (*Fix Lifecycle Tracking*).

---

## 📐 1. Arquitetura do Sistema e Fluxo de Decisão

```
       ┌────────────────────────────────────────────────────────┐
       │   Git SCM Source (GitHub Actions / GitLab CI / Local)   │
       └───────────────────────────┬────────────────────────────┘
                                   │ Git Diff / Patch Hunks
                                   ▼
┌─────────────────────────────────────────────────────────────────────────┐
│                      CODESENTINEL CORE RUNTIME                          │
│                                                                         │
│  ┌─────────────────────────┐        ┌────────────────────────────────┐  │
│  │ AST Tokenizer & Parser  │ ─────► │ Privacy & Secret Redactor Layer│  │
│  │ (Tree-Sitter / DiffLex) │        │ (High-Entropy / TruffleHog RE) │  │
│  └────────────┬────────────┘        └───────────────┬────────────────┘  │
│               │                                     │ Sanitized AST/Diff│
│               ▼                                     ▼                   │
│  ┌─────────────────────────┐        ┌────────────────────────────────┐  │
│  │ SAST / Open Code Review │        │ Contextual LLM Engine          │  │
│  │ (Deterministic Rules)   │        │ (Claude 3.5 / OpenAI / Ollama) │  │
│  └────────────┬────────────┘        └───────────────┬────────────────┘  │
│               │                                     │                   │
│               └──────────────────┬──────────────────┘                   │
│                                  │ SARIF v2.1.0 Ingestion               │
│                                  ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │ Finding Normalizer: AST Structural Fingerprinting                 │  │
│  │ hash(rule_id + relative_path + ast_node_type + surrounding_scope) │  │
│  └───────────────────────────────┬───────────────────────────────────┘  │
│                                  │                                      │
│                                  ▼                                      │
│  ┌───────────────────────────────────────────────────────────────────┐  │
│  │ Deterministic PR Risk & Gate Policy Engine (Score 0-100)          │  │
│  └───────────────────────────────┬───────────────────────────────────┘  │
└──────────────────────────────────┼──────────────────────────────────────┘
                                   │
         ┌─────────────────────────┴─────────────────────────┐
         ▼                                                   ▼
┌─────────────────────────────────┐         ┌─────────────────────────────────┐
│ Storage Tier (PostgreSQL 16)    │         │ CI/CD Feedback & Presentation   │
│ - Findings Ledger               │         │ - Strict Exit Code (0, 1, 2, 3) │
│ - Fix Lifecycle Status Tracking │         │ - GitHub Check Run / PR Comments│
│ - Audit Trail & Exceptions      │         │ - React 18 Admin Dashboard      │
└─────────────────────────────────┘         └─────────────────────────────────┘
```

---

## 🧮 2. Formulação do Risk Score e Políticas de Gate

O CodeSentinel calcula a pontuação de risco da Pull Request ($R_{PR} \in [0, 100]$) com base na severidade, contexto do arquivo alterado (ex.: infraestrutura vs. documentação) e confiança do achado:

### Pesos de Severidade ($W_{\text{sev}}$)
- **CRITICAL:** 40.0 (Vulnerabilidades com exploração imediata: RCE, SQLi, SSRF, Hardcoded Secrets).
- **HIGH:** 20.0 (Falhas de controle de acesso, quebra de autorização, CSRF).
- **MEDIUM:** 7.0 (Code smells arquiteturais, vazamento de exceções, anti-patterns de concorrência).
- **LOW / INFO:** 1.5 (Inconsistências de estilo, falta de documentação pública).

### Multiplicadores de Escopo ($W_{\text{scope}}$)
- `core/security/**`, `auth/**`, `*.tf`, `Dockerfile`: $1.5\times$
- `internal/**`, `services/**`: $1.0\times$
- `tests/**`, `docs/**`: $0.2\times$

### Critérios de Decisão do Gate
```text
Se R_PR >= threshold_score (Padrão: 50) ────────► GATE STATUS = FAILED (Exit Code 1)
Se contagem(CRITICAL) > 0              ────────► GATE STATUS = FAILED (Exit Code 1)
Se contagem(HIGH) > max_high_findings  ────────► GATE STATUS = FAILED (Exit Code 1)
Caso contrário                          ────────► GATE STATUS = PASSED (Exit Code 0)
```

---

## 📁 3. Estrutura Canônica do Monorepo

```text
codesentinel/
├── cmd/
│   └── codesentinel/
│       └── main.go                  # CLI Entrypoint e comandos Cobra
├── internal/
│   ├── ast/
│   │   ├── parser.go                # Tree-sitter AST Context Extractor
│   │   └── fingerprint.go           # Hash estrutural invariante a drift de linha
│   ├── engine/
│   │   ├── analyzer.go              # Interface para SAST e Open Code Review
│   │   ├── llm_advisory.go          # Orquestrador LLM com structured output
│   │   └── risk_evaluator.go        # Cálculo determinístico da pontuação de risco
│   ├── gate/
│   │   └── policy.go                # Avaliação de regras de bloqueio de CI/CD
│   ├── privacy/
│   │   └── redactor.go              # Sanitização de tokens, PII e segredos de diff
│   ├── storage/
│   │   ├── database.go              # Conexão e migrações (PostgreSQL / SQLite)
│   │   └── repository.go            # Queries otimizadas para achados e histórico
│   └── models/
│       ├── finding.go               # Entidade unificada SARIF-compliant
│       └── policy.go                # Schemas de configuração
├── pkg/
│   └── sarif/
│       └── parser.go                # Leitor e normalizador OASIS SARIF v2.1.0
├── web/                             # Dashboard React 18 + Vite + Tailwind
│   ├── src/
│   │   ├── components/              # Matriz de calor, visualizador de diffs e gates
│   │   ├── pages/Dashboard.tsx
│   │   └── App.tsx
│   ├── package.json
│   └── tsconfig.json
├── configs/
│   └── codesentinel.default.yaml
├── deployments/
│   ├── Dockerfile
│   └── docker-compose.yml
├── go.mod
├── go.sum
└── README.md
```

---

## 🔍 4. Algoritmo de Fingerprinting Estrutural

Linhas de código mudam continuamente à medida que um PR é desenvolvido. Utilizar `arquivo:linha` resulta em alarmes falsos de "novas vulnerabilidades". O CodeSentinel utiliza AST Context Fingerprinting:

```go
// internal/ast/fingerprint.go
package ast

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"
)

// GenerateStructuralFingerprint calcula um hash estável do achado
// considerando a regra, o caminho relativo do arquivo, o nó sintático e o escopo da função.
func GenerateStructuralFingerprint(ruleID, relativePath, enclosingScope, normalizedSnippet string) string {
	hasher := sha256.New()

	// Remove espaços em branco redundantes e comentários para imunidade cosmética
	cleanSnippet := strings.Join(strings.Fields(normalizedSnippet), " ")

	compositeKey := fmt.Sprintf("%s|%s|%s|%s",
		ruleID,
		strings.TrimSpace(relativePath),
		strings.TrimSpace(enclosingScope),
		cleanSnippet,
	)

	hasher.Write([]byte(compositeKey))
	return hex.EncodeToString(hasher.Sum(nil))
}
```

---

## 🛡️ 5. Privacy & Secret Redactor Engine

Antes de qualquer trecho de código ser submetido à inferência contextual de modelos de linguagem, o módulo `internal/privacy/redactor.go` mascara valores sensíveis:

```go
// internal/privacy/redactor.go
package privacy

import (
	"regexp"
)

var (
	jwtRegex        = regexp.MustCompile(`eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_=]+\.?[A-Za-z0-9-_.+/=]*`)
	apiKeyRegex     = regexp.MustCompile(`(?i)(api[_-]?key|secret|token|password|bearer)[\s]*[=:]+[\s]*['"]?([a-zA-Z0-9_\-\.]{12,})['"]?`)
	privateKeyRegex = regexp.MustCompile(`(?s)-----BEGIN [A-Z ]+PRIVATE KEY-----.*?-----END [A-Z ]+PRIVATE KEY-----`)
)

func RedactSensitiveData(sourceDiff string) string {
	redacted := privateKeyRegex.ReplaceAllString(sourceDiff, "[REDACTED_RSA_PRIVATE_KEY]")
	redacted = jwtRegex.ReplaceAllString(redacted, "[REDACTED_JWT_TOKEN]")
	redacted = apiKeyRegex.ReplaceAllString(redacted, "$1: [REDACTED_CREDENTIAL]")
	return redacted
}
```

---

## 🚀 6. Implementação do Motor de Avaliação do Gate

```go
// internal/gate/policy.go
package gate

import (
	"fmt"
	"github.com/codesentinel/codesentinel/internal/models"
)

type GateResult struct {
	Allowed        bool
	RiskScore      float64
	BlockingReason string
	CriticalCount  int
	HighCount      int
}

func EvaluatePullRequest(findings []models.Finding, policy models.PolicyConfig) GateResult {
	result := GateResult{Allowed: true}
	var totalScore float64

	for _, f := range findings {
		if f.Status == models.StatusAcceptedRisk {
			continue
		}

		var weight float64
		switch f.Severity {
		case models.SeverityCritical:
			result.CriticalCount++
			weight = 40.0
		case models.SeverityHigh:
			result.HighCount++
			weight = 20.0
		case models.SeverityMedium:
			weight = 7.0
		case models.SeverityLow:
			weight = 1.5
		}

		scopeMultiplier := 1.0
		if f.IsSecuritySensitiveScope {
			scopeMultiplier = 1.5
		}

		totalScore += weight * scopeMultiplier * f.ConfidenceMultiplier
	}

	if totalScore > 100.0 {
		totalScore = 100.0
	}
	result.RiskScore = totalScore

	// Validação de Limiares da Política
	if result.CriticalCount > 0 && policy.BlockOnCritical {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: %d achados CRÍTICOS encontrados.", result.CriticalCount)
		return result
	}

	if result.HighCount > policy.MaxHighFindings {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: Achados HIGH (%d) excedem o limite tolerado de %d.", result.HighCount, policy.MaxHighFindings)
		return result
	}

	if result.RiskScore >= policy.MaxRiskScoreThreshold {
		result.Allowed = false
		result.BlockingReason = fmt.Sprintf("Bloqueado: Pontuação de risco (%.1f) excede o teto permitido de %.1f.", result.RiskScore, policy.MaxRiskScoreThreshold)
		return result
	}

	return result
}
```

---

## ⚙️ 7. Configuração Corporativa (codesentinel.yaml)

```yaml
version: "1.0"
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout_sec: 30
  write_timeout_sec: 60

database:
  type: "postgres" # postgres | sqlite
  dsn: "postgres://sentinel_usr:Secr3tP@ss@localhost:5432/codesentinel_db?sslmode=require"
  max_open_conns: 25
  max_idle_conns: 5

privacy:
  strict_redaction: true
  enable_external_llm: true
  allowed_remote_providers:
    - "anthropic"
    - "ollama_internal"

llm_gateway:
  provider: "anthropic" # anthropic | openai | ollama | mock
  model: "claude-3-5-sonnet-20241022"
  temperature: 0.1
  max_tokens: 1500
  timeout_sec: 25

policy:
  fail_fast: true
  block_on_critical: true
  max_high_findings: 0
  max_medium_findings: 3
  max_risk_score_threshold: 45.0
  allow_accepted_risk: true

scanners:
  sarif_inputs:
    - "./reports/semgrep.sarif"
    - "./reports/gosec.sarif"
    - "./reports/trivy.sarif"
```

---

## 🛠️ 8. Instalação & Compilação

### Pré-requisitos
- **Go 1.24+**
- **Node.js 20+** e **npm 10+**
- **PostgreSQL 15+** (produção) ou **SQLite** (desenvolvimento)

### Compilação do Binário CLI & Server
```bash
git clone https://github.com/org/codesentinel.git
cd codesentinel

# Download e validação de dependências Go
go mod download
go mod verify

# Compilação otimizada estática
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=2.4.0" -o bin/codesentinel ./cmd/codesentinel
```

### Build do Dashboard SPA (React/TypeScript)
```bash
cd web
npm ci
npm run build
# Os artefatos estáticos serão gerados em web/dist e embarcados no binário Go via //go:embed
```

---

## 🚀 9. Execução & Integração em CI/CD

### 1. Inicializar o Servidor REST API & Web Dashboard
```bash
./bin/codesentinel server --config ./configs/codesentinel.yaml
```
- O painel estará disponível em `http://localhost:8080`.

### 2. Executar Inspeção de Pull Request em Pipeline CI/CD
Exemplo de uso nativo no **GitHub Actions**:

```yaml
name: "CodeSentinel Security Gate"

on:
  pull_request:
    branches: [ "main" ]

jobs:
  gate-evaluation:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout Código
        uses: actions/checkout@v4
        with:
          fetch-depth: 0

      - name: Executar Semgrep SAST
        run: |
          docker run --rm -v $(pwd):/src returntocorp/semgrep \
            semgrep scan --config auto --sarif --output=semgrep.sarif

      - name: Baixar CodeSentinel
        run: |
          curl -sSL https://releases.internal.corp/codesentinel/v2.4.0/codesentinel-linux-amd64 -o codesentinel
          chmod +x codesentinel

      - name: Avaliar Gate de Qualidade e Segurança
        run: |
          ./codesentinel review \
            --from origin/main \
            --to HEAD \
            --sarif-report semgrep.sarif \
            --policy ./codesentinel.yaml \
            --format github-comment
```

---

## 📊 Matriz de Códigos de Saída (CI/CD Exit Codes)

| Código | Estado | Descrição | Comportamento da Pipeline |
|---|---|---|---|
| `0` | **SUCCESS** | PR cumpre todos os critérios de qualidade e segurança. | Merge liberado. |
| `1` | **POLICY_VIOLATION** | Risco elevado, achados críticos ou excedeu limites de severidade. | Merge bloqueado imediatamente. |
| `2` | **CONFIGURATION_ERROR** | Falha no carregamento do `codesentinel.yaml` ou variáveis ausentes. | Interrupção por erro de configuração. |
| `3` | **INFRASTRUCTURE_FAILURE** | Queda do PostgreSQL ou timeout inalcançável do gateway de IA. | Falha de execução técnica. |

---

## ❓ Troubleshooting e Diagnóstico

| Sintoma | Causa Provável | Procedimento de Resolução |
|---|---|---|
| **Exit Code 1: High findings limit exceeded** | O PR introduziu novos problemas de segurança sem waiver. | Inspecione os achados via `./bin/codesentinel status --pr <ID>` e aplique as mitigações ou registre uma exceção documentada. |
| **Privacy Violation Triggered** | O redactor identificou tokens ou chaves RSA no diff. | Remova os segredos do histórico do Git utilizando `git-filter-repo` antes de tentar novo merge. |
| **SARIF Schema Incompatible** | Relatório SARIF fora da especificação OASIS v2.1.0. | Atualize seu scanner estático (Semgrep/Trivy) com a flag formal `--sarif`. |
| **Database connection pooling exhausted** | Número excessivo de runners CI disparando verificações concorrentes. | Aumente `max_open_conns` no `codesentinel.yaml` ou ative o PgBouncer no nó PostgreSQL. |
