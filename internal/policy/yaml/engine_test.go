package yaml

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadPolicy(t *testing.T) {
	dir := t.TempDir()
	policyPath := filepath.Join(dir, "policy.yaml")
	content := `version: "1"
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/.+$'
plaintext_allowlist:
  - stmt-abc
gates:
  - id: G001
    trigger_paths:
      - app/**
    required_attestations:
      - prompt_attestation
    message: "prompt missing"
`
	if err := os.WriteFile(policyPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	p, err := LoadPolicy(policyPath)
	if err != nil {
		t.Fatalf("LoadPolicy: %v", err)
	}
	if p.Version != "1" {
		t.Errorf("version = %q, want %q", p.Version, "1")
	}
	if p.OIDCIssuer != "https://token.actions.githubusercontent.com" {
		t.Errorf("oidc_issuer = %q", p.OIDCIssuer)
	}
	if len(p.PlaintextAllowlist) != 1 || p.PlaintextAllowlist[0] != "stmt-abc" {
		t.Errorf("plaintext_allowlist = %v", p.PlaintextAllowlist)
	}
	if len(p.Gates) != 1 {
		t.Fatalf("gates count = %d, want 1", len(p.Gates))
	}
	if p.Gates[0].ID != "G001" {
		t.Errorf("gate id = %q", p.Gates[0].ID)
	}
}

func TestLoadPolicyFileNotFound(t *testing.T) {
	_, err := LoadPolicy("/nonexistent/path.yaml")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestEvaluateWithChanged_NoViolations(t *testing.T) {
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"app/**"},
				RequiredAttestations: []string{"prompt_attestation"},
				Message:              "prompt missing",
			},
		},
	}
	statements := []StatementView{
		{AttestationType: "prompt_attestation", StatementID: "s1"},
	}
	changed := []string{"app/main.go"}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations, got %v", violations)
	}
}

func TestEvaluateWithChanged_MissingAttestation(t *testing.T) {
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"app/**"},
				RequiredAttestations: []string{"prompt_attestation", "eval_attestation"},
				Message:              "missing attestations",
			},
		},
	}
	statements := []StatementView{
		{AttestationType: "prompt_attestation", StatementID: "s1"},
	}
	changed := []string{"app/main.go"}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	if violations[0] != "missing attestations" {
		t.Errorf("violation = %q", violations[0])
	}
}

func TestEvaluateWithChanged_GateNotTriggered(t *testing.T) {
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"app/**"},
				RequiredAttestations: []string{"prompt_attestation"},
				Message:              "prompt missing",
			},
		},
	}
	statements := []StatementView{}
	changed := []string{"docs/readme.md"}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for unmatched paths, got %v", violations)
	}
}

func TestEvaluateWithChanged_PlaintextBlocked(t *testing.T) {
	policy := Policy{
		PlaintextAllowlist: []string{},
		Gates:              []Gate{},
	}
	statements := []StatementView{
		{
			AttestationType: "prompt_attestation",
			StatementID:     "s1",
			PrivacyMode:     "plaintext_explicit",
		},
	}
	changed := []string{}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation for plaintext block, got %d", len(violations))
	}
	if violations[0] != "Sensitive payload exposure blocked by policy." {
		t.Errorf("violation = %q", violations[0])
	}
}

func TestEvaluateWithChanged_PlaintextAllowlisted(t *testing.T) {
	policy := Policy{
		PlaintextAllowlist: []string{"s1"},
		Gates:              []Gate{},
	}
	statements := []StatementView{
		{
			AttestationType: "prompt_attestation",
			StatementID:     "s1",
			PrivacyMode:     "plaintext_explicit",
		},
	}
	changed := []string{}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 0 {
		t.Errorf("expected 0 violations for allowlisted plaintext, got %v", violations)
	}
}

func TestEvaluateWithChanged_DefaultMessageFormat(t *testing.T) {
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"data/**"},
				RequiredAttestations: []string{"corpus_attestation"},
				Message:              "",
			},
		},
	}
	statements := []StatementView{}
	changed := []string{"data/doc.txt"}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 1 {
		t.Fatalf("expected 1 violation, got %d", len(violations))
	}
	want := "G001 missing attestations: corpus_attestation"
	if violations[0] != want {
		t.Errorf("violation = %q, want %q", violations[0], want)
	}
}

func TestMatch_DoubleStarSuffix(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"app/main.go", "app/**", true},
		{"app/sub/deep.go", "app/**", true},
		{"app", "app/**", true},
		{"other/main.go", "app/**", false},
		{"application/main.go", "app/**", false},
	}
	for _, tt := range tests {
		got := match(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("match(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestMatch_ExactGlob(t *testing.T) {
	tests := []struct {
		path    string
		pattern string
		want    bool
	}{
		{"config.yaml", "*.yaml", true},
		{"config.json", "*.yaml", false},
	}
	for _, tt := range tests {
		got := match(tt.path, tt.pattern)
		if got != tt.want {
			t.Errorf("match(%q, %q) = %v, want %v", tt.path, tt.pattern, got, tt.want)
		}
	}
}

func TestExtract(t *testing.T) {
	payload := map[string]any{
		"attestation_type": "prompt_attestation",
		"statement_id":     "stmt-1",
		"privacy": map[string]any{
			"mode": "hash_only",
		},
		"annotations": map[string]any{
			"depends_on": "prompt_attestation, corpus_attestation",
		},
	}
	sv := extract(payload)
	if sv.AttestationType != "prompt_attestation" {
		t.Errorf("attestation_type = %q", sv.AttestationType)
	}
	if sv.StatementID != "stmt-1" {
		t.Errorf("statement_id = %q", sv.StatementID)
	}
	if sv.PrivacyMode != "hash_only" {
		t.Errorf("privacy_mode = %q", sv.PrivacyMode)
	}
	if len(sv.DependsOn) != 2 {
		t.Fatalf("depends_on len = %d, want 2", len(sv.DependsOn))
	}
	if sv.DependsOn[0] != "prompt_attestation" || sv.DependsOn[1] != "corpus_attestation" {
		t.Errorf("depends_on = %v", sv.DependsOn)
	}
}

func TestExtract_NoPrivacy(t *testing.T) {
	payload := map[string]any{
		"attestation_type": "eval_attestation",
		"statement_id":     "stmt-2",
	}
	sv := extract(payload)
	if sv.PrivacyMode != "" {
		t.Errorf("privacy_mode = %q, want empty", sv.PrivacyMode)
	}
	if len(sv.DependsOn) != 0 {
		t.Errorf("depends_on = %v, want empty", sv.DependsOn)
	}
}

func TestEvaluateWithChanged_MultipleGates(t *testing.T) {
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"app/**"},
				RequiredAttestations: []string{"prompt_attestation"},
				Message:              "prompt gate failed",
			},
			{
				ID:                   "G002",
				TriggerPaths:         []string{"data/**"},
				RequiredAttestations: []string{"corpus_attestation"},
				Message:              "corpus gate failed",
			},
		},
	}
	statements := []StatementView{}
	changed := []string{"app/main.go", "data/doc.txt"}

	violations, err := EvaluateWithChanged(policy, statements, changed)
	if err != nil {
		t.Fatal(err)
	}
	if len(violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(violations), violations)
	}
}
