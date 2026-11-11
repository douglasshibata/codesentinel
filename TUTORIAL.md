# 📘 Tutorial de Operação e Instalação: CodeSentinel - Continuous AI Code-Quality & Security Gate

Guia passo a passo para configuração, inicialização, execução de testes, pipelines SARIF e resolução de problemas no projeto **CodeSentinel - Continuous AI Code-Quality & Security Gate** (`codesentinel`).

---

## 🛠️ 1. Visão Geral e Arquitetura

> **"Deterministic gates. Context-aware AI code reviews. Zero secret leakage."**

O **CodeSentinel** é uma plataforma corporativa e portão de qualidade contínuo (*Continuous AI Code-Quality & Security Gate*) desenvolvida em Go e React/TypeScript. Ele orquestra motores de análise estática e pipelines de LLM para transformar achados brutos (SARIF v2.1.0/JSON) em um grafo relacional de auditoria com cálculo determinístico de risco (0–100), bloqueio preemptivo de merges via CLI/CI/CD, fingerprinting sintático invariante a mudanças de linha e verificação de ciclo de vida de correções (*Fix Lifecycle Tracking*).

- **Backend / CLI:** Go 1.24+ (Standard Go Project Layout, Cobra CLI, Gin REST API, GORM)
- **Frontend SPA:** React 18, TypeScript, TailwindCSS, Vite
- **Padrão de Ingestão:** OASIS SARIF v2.1.0
- **Segurança & Privacidade:** Privacy Layer com Regex-based High Entropy & Secret Redactor
- **Persistência:** PostgreSQL 16 (produção) / SQLite 3 (desenvolvimento/testes)

---

## 📋 2. Pré-requisitos do Sistema

Antes de iniciar a configuração local, certifique-se de ter os seguintes componentes instalados:

- **Git:** 2.40+ para controle de versão e extração de hunks/diffs.
- **Go:** 1.24+ (suporte a módulos Go e toolchains modernas).
- **Node.js & npm:** Node.js 20+ LTS e npm 10+.
- **Docker & Docker Compose:** Opcional, para execução do banco PostgreSQL e infraestrutura de CI.

---

## 🚀 3. Passo a Passo de Setup e Instalação

### Passo 1: Acessar a Pasta do Projeto
```bash
cd codesentinel
```

### Passo 2: Baixar e Validar Dependências Go
```bash
go mod download
go mod verify
```

### Passo 3: Compilação do Binário Go
```bash
# Compilar binário da CLI e Servidor CodeSentinel
CGO_ENABLED=0 go build -ldflags="-s -w -X main.version=2.4.0" -o bin/codesentinel ./cmd/codesentinel
```

### Passo 4: Instalar Dependências e Compilar o Dashboard React
```bash
cd web
npm ci
npm run build
cd ..
```

### Passo 5: Executar Infraestrutura via Docker Compose (Opcional)
```bash
docker compose -f deployments/docker-compose.yml up -d
```

---

## ⚙️ 4. Configuração do Ambiente

O CodeSentinel utiliza o arquivo de configuração corporativo `codesentinel.yaml` (ou template em `configs/codesentinel.default.yaml`):

```yaml
version: "1.0"
server:
  host: "0.0.0.0"
  port: 8080
  read_timeout_sec: 30
  write_timeout_sec: 60

database:
  type: "sqlite" # "sqlite" ou "postgres"
  dsn: "codesentinel.db"
  max_open_conns: 25
  max_idle_conns: 5

privacy:
  strict_redaction: true
  enable_external_llm: true
  allowed_remote_providers:
    - "anthropic"
    - "ollama_internal"

llm_gateway:
  provider: "mock" # "anthropic", "openai", "ollama", "mock"
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

## 💻 5. Guia de Execução

### 1. Inicializar o Servidor REST API & Dashboard
```bash
./bin/codesentinel server --config ./codesentinel.yaml
```
Acesse o painel web em: `http://localhost:8080`.

### 2. Avaliar Pull Request no CI/CD com Relatórios SARIF
```bash
./bin/codesentinel review \
  --from origin/main \
  --to HEAD \
  --policy ./codesentinel.yaml \
  --format json
```

### 3. Códigos de Saída do CI/CD Gate
- `0`: **SUCCESS** (Merge aprovado).
- `1`: **POLICY_VIOLATION** (Bloqueio preemptivo por achado crítico ou pontuação de risco).
- `2`: **CONFIGURATION_ERROR** (Arquivo de configuração ou variáveis inválidas).
- `3`: **INFRASTRUCTURE_FAILURE** (Falha de conectividade ou indisponibilidade de serviços).

---

## 🧪 6. Testes e Validação

Execute a suíte de testes unitários e de integração:

```bash
cd codesentinel
go test -v ./...
```

Para verificar o build do frontend:
```bash
cd codesentinel/web
npm run build
```

---

## 🔍 7. Troubleshooting e Resolução de Problemas

1. **Exit Code 1: Bloqueio do Gate:**
   - Execute `./bin/codesentinel findings` para listar os achados impeditivos e proceda à correção do código ou configure uma política com waiver justificado.
2. **Privacy Redactor ativado:**
   - Verifique se segredos como JWT ou tokens de API estão presentes nos diffs de commits recentes e remova-os com `git-filter-repo` antes de abrir a PR.
3. **Erros de Parseamento SARIF:**
   - Assegure-se de que a ferramenta de SAST (Semgrep, Gosec, Trivy) gerou a saída na versão padrão OASIS SARIF v2.1.0 (`"$schema": "https://raw.githubusercontent.com/oasis-tcs/sarif-spec/master/Schemata/sarif-schema-2.1.0.json"`).
