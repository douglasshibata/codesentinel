import React, { useState } from 'react';
import {
  ShieldAlert, GitPullRequest, Database, AlertTriangle, Activity,
  Settings, History, Lock, FileCode, CheckCircle2, XCircle, Search,
  Terminal, Sparkles, ChevronRight, Layers, ArrowUpRight, ArrowDownRight,
  Code2, Check, Copy, RefreshCw, Eye
} from 'lucide-react';
import { Finding, Review, Repository, Policy } from './types';

const mockRepos: Repository[] = [
  { id: 'repo-1', name: 'payment-gateway-service', type: 'github', url: 'https://github.com/org/payments', local_path: '/srv/payments', default_branch: 'main', profile: 'Security Strict' },
  { id: 'repo-2', name: 'user-auth-api', type: 'gitlab', url: 'https://gitlab.com/org/auth', local_path: '/srv/auth', default_branch: 'main', profile: 'Backend Microservice' },
  { id: 'repo-3', name: 'analytics-dashboard', type: 'local', url: '', local_path: '/local/analytics', default_branch: 'master', profile: 'Frontend React' },
];

const mockFindings: Finding[] = [
  {
    id: 'fnd-8a2b1c4d',
    review_id: 'rev-001',
    repository_id: 'repo-1',
    pull_request_id: 'PR-1842',
    commit_sha: 'a8f912c',
    branch: 'feature/payment-retry',
    file: 'payments/service.go',
    line: 42,
    end_line: 42,
    rule_id: 'SEC-001',
    category: 'Secrets',
    severity: 'CRITICAL',
    confidence: 'HIGH',
    title: 'Hardcoded Credentials or Secret Token Detected',
    description: 'A sensitive credential or API secret token was found hardcoded in source file.',
    impact: 'Attackers could extract credentials from source code and compromise financial gateway API.',
    suggested_fix: 'Store sensitive secrets in environment variables or vault.',
    fingerprint: 'a1b2c3d4e5f6',
    first_seen: '2026-03-10T10:00:00Z',
    last_seen: '2026-03-15T14:20:00Z',
    status: 'Open',
    is_new: true,
    author: 'dev-alex',
    code_snippet: 'apiKey := "sk_live_992128109a8f192b0012a"'
  },
  {
    id: 'fnd-9c3d2e1a',
    review_id: 'rev-001',
    repository_id: 'repo-1',
    pull_request_id: 'PR-1842',
    commit_sha: 'a8f912c',
    branch: 'feature/payment-retry',
    file: 'payments/webhook.go',
    line: 88,
    end_line: 88,
    rule_id: 'SEC-002',
    category: 'SQL injection',
    severity: 'CRITICAL',
    confidence: 'HIGH',
    title: 'Potential SQL Injection via String Concatenation',
    description: 'Dynamic SQL query constructed by string concatenation instead of parameterized placeholders.',
    impact: 'Malicious webhook payloads can inject arbitrary SQL commands and alter ledger data.',
    suggested_fix: 'Use parameterized queries db.QueryContext(ctx, "SELECT ... WHERE id = $1", id).',
    fingerprint: 'b2c3d4e5f6a1',
    first_seen: '2026-03-15T14:20:00Z',
    last_seen: '2026-03-15T14:20:00Z',
    status: 'Open',
    is_new: true,
    author: 'dev-alex',
    code_snippet: 'query := "SELECT * FROM webhooks WHERE tx_id = " + txID'
  },
  {
    id: 'fnd-1e2f3a4b',
    review_id: 'rev-001',
    repository_id: 'repo-1',
    pull_request_id: 'PR-1842',
    commit_sha: 'a8f912c',
    branch: 'feature/payment-retry',
    file: 'payments/retry.go',
    line: 115,
    end_line: 115,
    rule_id: 'SEC-005',
    category: 'SSRF',
    severity: 'HIGH',
    confidence: 'HIGH',
    title: 'Potential Server-Side Request Forgery (SSRF)',
    description: 'Making outbound HTTP requests directly to URLs provided in untrusted request parameters.',
    impact: 'Allows internal port scanning and unauthorized requests to cloud metadata endpoints.',
    suggested_fix: 'Validate destination domain against strict allowlist before executing http.Get().',
    fingerprint: 'c3d4e5f6a1b2',
    first_seen: '2026-03-14T09:00:00Z',
    last_seen: '2026-03-15T14:20:00Z',
    status: 'Open',
    is_new: false,
    author: 'dev-maria',
    code_snippet: 'resp, err := http.Get(targetWebhookURL)'
  }
];

export function App() {
  const [activeTab, setActiveTab] = useState<string>('overview');
  const [selectedFinding, setSelectedFinding] = useState<Finding | null>(mockFindings[0]);
  const [aiQuestion, setAiQuestion] = useState<string>('');
  const [aiResponse, setAiResponse] = useState<string | null>(null);
  const [patchGenerated, setPatchGenerated] = useState<boolean>(false);

  const renderNav = () => (
    <nav style={{ width: '240px', backgroundColor: '#0f172a', borderRight: '1px solid #1e293b', padding: '1.5rem 1rem', display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
      <div style={{ display: 'flex', alignItems: 'center', gap: '0.75rem', marginBottom: '2rem', paddingLeft: '0.5rem' }}>
        <ShieldAlert style={{ color: '#10b981', width: '28px', height: '28px' }} />
        <div>
          <div style={{ fontWeight: 700, fontSize: '1.1rem', letterSpacing: '-0.5px', color: '#f8fafc' }}>CodeSentinel</div>
          <div style={{ fontSize: '0.7rem', color: '#64748b', fontFamily: 'Fira Code' }}>v1.2.0-GA</div>
        </div>
      </div>

      {[
        { id: 'overview', label: 'Overview', icon: Activity },
        { id: 'repositories', label: 'Repositories', icon: Database },
        { id: 'pull-requests', label: 'Pull Requests', icon: GitPullRequest },
        { id: 'findings', label: 'Findings', icon: ShieldAlert },
        { id: 'security', label: 'Security', icon: Lock },
        { id: 'trends', label: 'Trends & Health', icon: Layers },
        { id: 'policies', label: 'Policies', icon: AlertTriangle },
        { id: 'history', label: 'Review History', icon: History },
        { id: 'settings', label: 'Settings & AI', icon: Settings },
      ].map((item) => {
        const Icon = item.icon;
        const isActive = activeTab === item.id;
        return (
          <button
            key={item.id}
            onClick={() => setActiveTab(item.id)}
            style={{
              display: 'flex',
              alignItems: 'center',
              gap: '0.75rem',
              padding: '0.65rem 0.85rem',
              borderRadius: '6px',
              border: 'none',
              backgroundColor: isActive ? '#1e293b' : 'transparent',
              color: isActive ? '#10b981' : '#94a3b8',
              fontWeight: isActive ? 600 : 400,
              cursor: 'pointer',
              textAlign: 'left',
              fontSize: '0.9rem',
              transition: 'all 0.15s ease'
            }}
          >
            <Icon style={{ width: '18px', height: '18px' }} />
            {item.label}
          </button>
        );
      })}

      <div style={{ marginTop: 'auto', padding: '1rem 0.5rem', borderTop: '1px solid #1e293b', fontSize: '0.75rem', color: '#64748b' }}>
        <div>Engine: <span style={{ color: '#38bdf8' }}>Open Code Review v2.4</span></div>
        <div>Mode: <span style={{ color: '#10b981' }}>Continuous Gate</span></div>
      </div>
    </nav>
  );

  const renderOverview = () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
        <div>
          <h1 style={{ fontSize: '1.5rem', fontWeight: 700, margin: 0, color: '#f8fafc' }}>Engineering Quality Overview</h1>
          <p style={{ margin: '0.25rem 0 0 0', color: '#64748b', fontSize: '0.875rem' }}>Continuous code quality and automated security gate monitoring</p>
        </div>
        <button style={{ backgroundColor: '#10b981', color: '#090d16', fontWeight: 600, border: 'none', padding: '0.5rem 1rem', borderRadius: '6px', cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
          <RefreshCw style={{ width: '16px', height: '16px' }} /> Run Full Scan
        </button>
      </div>

      {/* Top Metrics Cards */}
      <div style={{ display: 'grid', gridTemplateColumns: 'repeat(4, 1fr)', gap: '1rem' }}>
        {[
          { title: 'Engineering Health', value: '87 / 100', change: '+4.2%', isPos: true, color: '#10b981' },
          { title: 'Security Score', value: '82 / 100', change: '+6.1%', isPos: true, color: '#38bdf8' },
          { title: 'Open Critical Findings', value: '2', change: '-31%', isPos: true, color: '#ef4444' },
          { title: 'PRs Reviewed (30d)', value: '1,284', change: 'Avg 2m 14s', isPos: true, color: '#a855f7' },
        ].map((m, idx) => (
          <div key={idx} style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '1.25rem' }}>
            <div style={{ fontSize: '0.8rem', color: '#64748b', fontWeight: 500 }}>{m.title}</div>
            <div style={{ fontSize: '1.75rem', fontWeight: 700, margin: '0.5rem 0', color: '#f8fafc', fontFamily: 'Fira Code' }}>{m.value}</div>
            <div style={{ fontSize: '0.75rem', color: m.color, display: 'flex', alignItems: 'center', gap: '0.25rem' }}>
              {m.change}
            </div>
          </div>
        ))}
      </div>

      {/* Code Health Gauges */}
      <div style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '1.5rem' }}>
        <h3 style={{ margin: '0 0 1.25rem 0', fontSize: '1rem', fontWeight: 600, color: '#f8fafc' }}>Domain Quality Scores</h3>
        <div style={{ display: 'flex', flexDirection: 'column', gap: '1rem' }}>
          {[
            { name: 'Overall Health', score: 87, color: '#10b981' },
            { name: 'Security', score: 82, color: '#ef4444' },
            { name: 'Reliability', score: 86, color: '#f59e0b' },
            { name: 'Maintainability', score: 91, color: '#38bdf8' },
          ].map((cat, i) => (
            <div key={i}>
              <div style={{ display: 'flex', justifyContent: 'space-between', fontSize: '0.85rem', marginBottom: '0.35rem' }}>
                <span style={{ fontWeight: 500, color: '#e2e8f0' }}>{cat.name}</span>
                <span style={{ fontFamily: 'Fira Code', fontWeight: 600, color: cat.color }}>{cat.score} / 100</span>
              </div>
              <div style={{ width: '100%', height: '8px', backgroundColor: '#1e293b', borderRadius: '4px', overflow: 'hidden' }}>
                <div style={{ width: `${cat.score}%`, height: '100%', backgroundColor: cat.color, borderRadius: '4px' }} />
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* Active Pull Requests Under Review */}
      <div style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '1.5rem' }}>
        <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem' }}>
          <h3 style={{ margin: 0, fontSize: '1rem', fontWeight: 600, color: '#f8fafc' }}>Recent PR Risk Evaluations</h3>
          <span style={{ fontSize: '0.8rem', color: '#10b981', cursor: 'pointer' }}>View all PRs →</span>
        </div>
        <table style={{ width: '100%', borderCollapse: 'collapse', textAlign: 'left', fontSize: '0.875rem' }}>
          <thead>
            <tr style={{ borderBottom: '1px solid #1e293b', color: '#64748b' }}>
              <th style={{ padding: '0.75rem 0' }}>PR</th>
              <th>Repository</th>
              <th>Author</th>
              <th>Risk Score</th>
              <th>Status</th>
              <th>Gate</th>
            </tr>
          </thead>
          <tbody>
            {[
              { pr: 'PR #1842', title: 'Refactor payment processing', repo: 'payment-gateway-service', author: 'dev-alex', score: '78 / 100', level: 'HIGH', status: 'Blocked' },
              { pr: 'PR #1841', title: 'Update user session cache key', repo: 'user-auth-api', author: 'dev-maria', score: '12 / 100', level: 'LOW', status: 'Passed' },
              { pr: 'PR #1840', title: 'Migrate DB migration hooks', repo: 'analytics-dashboard', author: 'dev-john', score: '35 / 100', level: 'MODERATE', status: 'Passed' },
            ].map((row, idx) => (
              <tr key={idx} style={{ borderBottom: '1px solid #1e293b' }}>
                <td style={{ padding: '0.85rem 0', fontFamily: 'Fira Code', fontWeight: 600, color: '#f8fafc' }}>{row.pr}</td>
                <td style={{ color: '#cbd5e1' }}>{row.repo}</td>
                <td style={{ color: '#94a3b8' }}>{row.author}</td>
                <td>
                  <span style={{
                    fontFamily: 'Fira Code',
                    fontWeight: 700,
                    color: row.level === 'HIGH' ? '#ef4444' : row.level === 'MODERATE' ? '#f59e0b' : '#10b981'
                  }}>
                    {row.score} [{row.level}]
                  </span>
                </td>
                <td>
                  <span style={{
                    padding: '0.2rem 0.5rem',
                    borderRadius: '4px',
                    fontSize: '0.75rem',
                    fontWeight: 600,
                    backgroundColor: row.status === 'Blocked' ? '#450a0a' : '#064e3b',
                    color: row.status === 'Blocked' ? '#f87171' : '#34d399'
                  }}>
                    {row.status}
                  </span>
                </td>
                <td style={{ cursor: 'pointer', color: '#38bdf8' }} onClick={() => setActiveTab('pull-requests')}>Inspect →</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );

  const renderPRReview = () => (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', backgroundColor: '#0f172a', border: '1px solid #1e293b', padding: '1.25rem', borderRadius: '8px' }}>
        <div>
          <div style={{ fontSize: '0.8rem', color: '#64748b', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
            <span>payment-gateway-service</span> / <span style={{ fontFamily: 'Fira Code', color: '#38bdf8' }}>PR #1842</span>
          </div>
          <h2 style={{ margin: '0.25rem 0', fontSize: '1.25rem', fontWeight: 700, color: '#f8fafc' }}>Refactor payment processing logic</h2>
          <div style={{ fontSize: '0.8rem', color: '#94a3b8' }}>Branch: <code style={{ backgroundColor: '#1e293b', padding: '2px 6px', borderRadius: '4px', color: '#f8fafc' }}>feature/payment-retry</code> → <code style={{ backgroundColor: '#1e293b', padding: '2px 6px', borderRadius: '4px', color: '#f8fafc' }}>main</code></div>
        </div>
        <div style={{ textAlign: 'right' }}>
          <div style={{ fontSize: '0.75rem', color: '#64748b' }}>PULL REQUEST RISK</div>
          <div style={{ fontSize: '1.75rem', fontWeight: 800, color: '#ef4444', fontFamily: 'Fira Code' }}>78 / 100 [HIGH]</div>
          <div style={{ fontSize: '0.75rem', color: '#ef4444', fontWeight: 600 }}>MERGE BLOCKED BY POLICY</div>
        </div>
      </div>

      {/* PR Inline Findings Inspector */}
      <div style={{ display: 'grid', gridTemplateColumns: '320px 1fr', gap: '1.25rem' }}>
        {/* Findings List */}
        <div style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '1rem', display: 'flex', flexDirection: 'column', gap: '0.75rem' }}>
          <div style={{ fontWeight: 600, fontSize: '0.9rem', color: '#f8fafc', marginBottom: '0.25rem' }}>3 Active PR Findings</div>
          {mockFindings.map((f) => {
            const isSelected = selectedFinding?.id === f.id;
            return (
              <div
                key={f.id}
                onClick={() => { setSelectedFinding(f); setAiResponse(null); setPatchGenerated(false); }}
                style={{
                  padding: '0.85rem',
                  borderRadius: '6px',
                  backgroundColor: isSelected ? '#1e293b' : '#0b0f19',
                  border: isSelected ? '1px solid #38bdf8' : '1px solid #1e293b',
                  cursor: 'pointer',
                  transition: 'all 0.15s ease'
                }}
              >
                <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '0.25rem' }}>
                  <span style={{
                    fontSize: '0.65rem',
                    fontWeight: 700,
                    padding: '2px 6px',
                    borderRadius: '4px',
                    backgroundColor: f.severity === 'CRITICAL' ? '#450a0a' : '#422006',
                    color: f.severity === 'CRITICAL' ? '#f87171' : '#fba518'
                  }}>
                    {f.severity}
                  </span>
                  <span style={{ fontSize: '0.7rem', fontFamily: 'Fira Code', color: '#64748b' }}>{f.rule_id}</span>
                </div>
                <div style={{ fontSize: '0.85rem', fontWeight: 600, color: '#f8fafc', marginBottom: '0.25rem' }}>{f.title}</div>
                <div style={{ fontSize: '0.75rem', fontFamily: 'Fira Code', color: '#94a3b8' }}>{f.file}:{f.line}</div>
              </div>
            );
          })}
        </div>

        {/* Inspection Panel */}
        {selectedFinding && (
          <div style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '1.5rem', display: 'flex', flexDirection: 'column', gap: '1.25rem' }}>
            <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'flex-start' }}>
              <div>
                <span style={{ fontSize: '0.75rem', color: '#38bdf8', fontFamily: 'Fira Code', fontWeight: 600 }}>{selectedFinding.category}</span>
                <h3 style={{ margin: '0.25rem 0', fontSize: '1.1rem', fontWeight: 700, color: '#f8fafc' }}>{selectedFinding.title}</h3>
                <div style={{ fontSize: '0.8rem', color: '#94a3b8', fontFamily: 'Fira Code' }}>File: {selectedFinding.file} (Line {selectedFinding.line})</div>
              </div>
              <button
                onClick={() => setPatchGenerated(!patchGenerated)}
                style={{ backgroundColor: '#10b981', color: '#090d16', border: 'none', padding: '0.5rem 1rem', borderRadius: '6px', fontWeight: 600, cursor: 'pointer', display: 'flex', alignItems: 'center', gap: '0.4rem' }}
              >
                <Sparkles style={{ width: '16px', height: '16px' }} /> Generate Fix Patch
              </button>
            </div>

            {/* Affected Code Snippet */}
            <div>
              <div style={{ fontSize: '0.75rem', color: '#64748b', fontWeight: 600, marginBottom: '0.4rem' }}>AFFECTED CODE LINE</div>
              <div style={{ backgroundColor: '#0b0f19', border: '1px solid #1e293b', padding: '1rem', borderRadius: '6px', fontFamily: 'Fira Code', fontSize: '0.85rem', color: '#f87171' }}>
                <span style={{ color: '#64748b', marginRight: '1rem' }}>{selectedFinding.line} |</span>
                {selectedFinding.code_snippet}
              </div>
            </div>

            {/* Impact & Fix */}
            <div style={{ display: 'grid', gridTemplateColumns: '1fr 1fr', gap: '1rem' }}>
              <div style={{ backgroundColor: '#0b0f19', border: '1px solid #1e293b', padding: '1rem', borderRadius: '6px' }}>
                <div style={{ fontSize: '0.75rem', color: '#ef4444', fontWeight: 600, marginBottom: '0.25rem' }}>SECURITY / QUALITY IMPACT</div>
                <div style={{ fontSize: '0.85rem', color: '#cbd5e1' }}>{selectedFinding.impact}</div>
              </div>
              <div style={{ backgroundColor: '#0b0f19', border: '1px solid #1e293b', padding: '1rem', borderRadius: '6px' }}>
                <div style={{ fontSize: '0.75rem', color: '#10b981', fontWeight: 600, marginBottom: '0.25rem' }}>RECOMMENDED REMEDIATION</div>
                <div style={{ fontSize: '0.85rem', color: '#cbd5e1' }}>{selectedFinding.suggested_fix}</div>
              </div>
            </div>

            {/* Patch Display */}
            {patchGenerated && (
              <div style={{ backgroundColor: '#022c22', border: '1px solid #059669', padding: '1rem', borderRadius: '6px' }}>
                <div style={{ fontSize: '0.8rem', fontWeight: 700, color: '#34d399', marginBottom: '0.5rem', display: 'flex', alignItems: 'center', gap: '0.5rem' }}>
                  <Code2 style={{ width: '16px', height: '16px' }} /> AI Fix Patch Proposal (BEFORE / AFTER)
                </div>
                <pre style={{ margin: 0, fontFamily: 'Fira Code', fontSize: '0.8rem', color: '#e2e8f0', whiteSpace: 'pre-wrap' }}>
{`--- a/${selectedFinding.file}
+++ b/${selectedFinding.file}
@@ -${selectedFinding.line},1 +${selectedFinding.line},1 @@
- ${selectedFinding.code_snippet}
+ apiKey := os.Getenv("PAYMENT_API_SECRET_KEY") // Sanitized via vault env`}
                </pre>
              </div>
            )}

            {/* AI Assistant Question Box */}
            <div style={{ backgroundColor: '#0b0f19', border: '1px solid #1e293b', padding: '1rem', borderRadius: '6px' }}>
              <div style={{ fontSize: '0.8rem', fontWeight: 600, color: '#38bdf8', marginBottom: '0.5rem' }}>AI Explanation Assistant</div>
              <div style={{ display: 'flex', gap: '0.5rem' }}>
                <input
                  type="text"
                  value={aiQuestion}
                  onChange={(e) => setAiQuestion(e.target.value)}
                  placeholder="Ask e.g. 'Why is this dangerous?' or 'What tests should I add?'"
                  style={{ flex: 1, backgroundColor: '#0f172a', border: '1px solid #334155', borderRadius: '4px', padding: '0.5rem', color: '#f8fafc', fontSize: '0.85rem' }}
                />
                <button
                  onClick={() => setAiResponse(`The finding in ${selectedFinding.file} triggers ${selectedFinding.rule_id} because static analysis confirmed dynamic input inclusion without sanitization. To prevent exploitability, wrap parameters with prepared statement context or environment variable access.`)}
                  style={{ backgroundColor: '#38bdf8', color: '#090d16', border: 'none', padding: '0.5rem 1rem', borderRadius: '4px', fontWeight: 600, cursor: 'pointer' }}
                >
                  Ask AI
                </button>
              </div>
              {aiResponse && (
                <div style={{ marginTop: '0.75rem', fontSize: '0.85rem', color: '#e2e8f0', backgroundColor: '#0f172a', padding: '0.75rem', borderRadius: '4px', borderLeft: '3px solid #38bdf8' }}>
                  {aiResponse}
                </div>
              )}
            </div>
          </div>
        )}
      </div>
    </div>
  );

  return (
    <div style={{ display: 'flex', minHeight: '100vh', backgroundColor: '#0b0f19', color: '#e2e8f0' }}>
      {renderNav()}
      <main style={{ flex: 1, padding: '2rem', overflowY: 'auto' }}>
        {activeTab === 'overview' && renderOverview()}
        {activeTab === 'pull-requests' && renderPRReview()}
        {activeTab !== 'overview' && activeTab !== 'pull-requests' && (
          <div style={{ backgroundColor: '#0f172a', border: '1px solid #1e293b', borderRadius: '8px', padding: '2rem', textAlign: 'center' }}>
            <h2 style={{ fontSize: '1.25rem', color: '#f8fafc' }}>{activeTab.toUpperCase()} Panel Active</h2>
            <p style={{ color: '#64748b' }}>Connected to CodeSentinel backend REST API running on port 8080.</p>
          </div>
        )}
      </main>
    </div>
  );
}
