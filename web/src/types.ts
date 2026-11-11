export interface Finding {
  id: string;
  review_id: string;
  repository_id: string;
  pull_request_id: string;
  commit_sha: string;
  branch: string;
  file: string;
  line: number;
  end_line: number;
  rule_id: string;
  category: string;
  severity: 'CRITICAL' | 'HIGH' | 'MEDIUM' | 'LOW' | 'INFO';
  confidence: string;
  title: string;
  description: string;
  impact: string;
  suggested_fix: string;
  fingerprint: string;
  first_seen: string;
  last_seen: string;
  status: 'Open' | 'Confirmed' | 'False Positive' | 'Fixed' | 'Ignored' | 'Accepted Risk';
  is_new: boolean;
  author?: string;
  code_snippet?: string;
}

export interface Review {
  id: string;
  repository_id: string;
  pull_request_id: string;
  commit_sha: string;
  base_commit: string;
  risk_score: number;
  risk_level: string;
  status: string;
  total_findings: number;
  critical_count: number;
  high_count: number;
  medium_count: number;
  low_count: number;
  duration_ms: number;
  created_at: string;
}

export interface Repository {
  id: string;
  name: string;
  type: string;
  url: string;
  local_path: string;
  default_branch: string;
  profile: string;
}

export interface Policy {
  id: string;
  repository_id: string;
  block_on_severity: string;
  block_security: boolean;
  max_high_findings: number;
  max_medium_findings: number;
  allow_accepted_risk: boolean;
}
