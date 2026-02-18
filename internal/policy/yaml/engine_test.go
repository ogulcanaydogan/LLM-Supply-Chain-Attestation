package yaml

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
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

// --- LoadStatements Tests ---

func writeStatement(t *testing.T, dir, name string, attType, stmtID, privacyMode string) string {
	t.Helper()
	stmt := map[string]any{
		"attestation_type": attType,
		"statement_id":     stmtID,
		"privacy": map[string]any{
			"mode": privacyMode,
		},
	}
	raw, _ := json.MarshalIndent(stmt, "", "  ")
	path := filepath.Join(dir, name)
	os.WriteFile(path, raw, 0o644)
	return path
}

func writeBundle(t *testing.T, dir, name string, attType, stmtID, privacyMode string) string {
	t.Helper()
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     stmtID,
		"attestation_type": attType,
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-01-01T00:00:00Z",
		"generator":        map[string]any{"name": "llmsa", "version": "0.1.0", "git_sha": "abc"},
		"subject":          []any{},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:abc",
			"system_prompt_digest": "sha256:def",
			"template_digests":     []any{},
			"tool_schema_digests":  []any{},
			"safety_policy_digest": "sha256:safe",
		},
		"privacy": map[string]any{"mode": privacyMode},
	}

	keyPath := filepath.Join(dir, "key.pem")
	if _, err := os.Stat(keyPath); os.IsNotExist(err) {
		sign.GeneratePEMPrivateKey(keyPath)
	}
	signer, _ := sign.NewPEMSigner(keyPath)
	canonical, _ := hash.CanonicalJSON(statement)
	material, _ := signer.Sign(canonical)
	bundle, _ := sign.CreateBundle(statement, material)

	bundlePath := filepath.Join(dir, name)
	sign.WriteBundle(bundlePath, bundle)
	return bundlePath
}

func TestLoadStatements_FromDirectory(t *testing.T) {
	dir := t.TempDir()
	writeStatement(t, dir, "prompt.json", "prompt_attestation", "s1", "hash_only")
	writeStatement(t, dir, "eval.json", "eval_attestation", "s2", "hash_only")

	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 statements, got %d", len(views))
	}
	types := map[string]bool{}
	for _, v := range views {
		types[v.AttestationType] = true
	}
	if !types["prompt_attestation"] || !types["eval_attestation"] {
		t.Errorf("missing expected types: %v", types)
	}
}

func TestLoadStatements_FromSingleFile(t *testing.T) {
	dir := t.TempDir()
	path := writeStatement(t, dir, "single.json", "corpus_attestation", "c1", "hash_only")

	views, err := LoadStatements(path)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 statement, got %d", len(views))
	}
	if views[0].AttestationType != "corpus_attestation" {
		t.Errorf("type = %q", views[0].AttestationType)
	}
}

func TestLoadStatements_BundleFile(t *testing.T) {
	dir := t.TempDir()
	writeBundle(t, dir, "prompt.bundle.json", "prompt_attestation", "b1", "hash_only")

	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1 statement from bundle, got %d", len(views))
	}
	if views[0].StatementID != "b1" {
		t.Errorf("statement_id = %q", views[0].StatementID)
	}
}

func TestLoadStatements_MixedFiles(t *testing.T) {
	dir := t.TempDir()
	writeStatement(t, dir, "eval.json", "eval_attestation", "s1", "hash_only")
	writeBundle(t, dir, "prompt.bundle.json", "prompt_attestation", "b1", "hash_only")

	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 2 {
		t.Fatalf("expected 2 statements from mixed dir, got %d", len(views))
	}
}

func TestLoadStatements_EmptyDir(t *testing.T) {
	dir := t.TempDir()
	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 0 {
		t.Errorf("expected 0 statements from empty dir, got %d", len(views))
	}
}

func TestLoadStatements_NonexistentPath(t *testing.T) {
	_, err := LoadStatements("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

func TestLoadStatements_IgnoresSubdirectories(t *testing.T) {
	dir := t.TempDir()
	writeStatement(t, dir, "prompt.json", "prompt_attestation", "s1", "hash_only")
	os.MkdirAll(filepath.Join(dir, "subdir"), 0o755)

	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Errorf("expected 1 statement, got %d (should skip subdir)", len(views))
	}
}

func TestLoadStatements_PrivacyModeExtracted(t *testing.T) {
	dir := t.TempDir()
	writeStatement(t, dir, "stmt.json", "prompt_attestation", "s1", "encrypted_payload")

	views, err := LoadStatements(dir)
	if err != nil {
		t.Fatal(err)
	}
	if len(views) != 1 {
		t.Fatalf("expected 1, got %d", len(views))
	}
	if views[0].PrivacyMode != "encrypted_payload" {
		t.Errorf("privacy_mode = %q, want encrypted_payload", views[0].PrivacyMode)
	}
}

// --- ChangedFiles Tests ---

func TestChangedFiles_DefaultRef(t *testing.T) {
	// Should not panic or error on empty ref.
	files, err := ChangedFiles("")
	if err != nil {
		t.Fatalf("ChangedFiles empty ref: %v", err)
	}
	// We don't assert on the result since it depends on git state,
	// but we verify it returns without error.
	_ = files
}

func TestChangedFiles_InvalidRef(t *testing.T) {
	// A nonsense ref should return empty (graceful degradation), not error.
	files, err := ChangedFiles("this-ref-definitely-does-not-exist-zzzz")
	if err != nil {
		t.Fatalf("ChangedFiles invalid ref: %v", err)
	}
	_ = files
}

// --- Evaluate Integration ---

func TestEvaluate_WithGitRef(t *testing.T) {
	// Test that Evaluate (which calls ChangedFiles internally) doesn't panic.
	policy := Policy{
		Gates: []Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"app/**"},
				RequiredAttestations: []string{"prompt_attestation"},
			},
		},
	}
	statements := []StatementView{
		{AttestationType: "prompt_attestation", StatementID: "s1"},
	}
	violations, err := Evaluate(policy, statements, "HEAD~1")
	if err != nil {
		t.Fatalf("Evaluate: %v", err)
	}
	// Don't assert specific count since it depends on git state.
	_ = violations
}

func TestLoadPolicy_InvalidYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.yaml")
	os.WriteFile(path, []byte("not: valid: yaml: [unclosed"), 0o644)
	_, err := LoadPolicy(path)
	if err == nil {
		t.Fatal("expected error for invalid YAML")
	}
}

func TestExtract_EmptyPayload(t *testing.T) {
	sv := extract(map[string]any{})
	if sv.AttestationType != "" {
		t.Errorf("expected empty type, got %q", sv.AttestationType)
	}
	if sv.PrivacyMode != "" {
		t.Errorf("expected empty privacy, got %q", sv.PrivacyMode)
	}
	if len(sv.DependsOn) != 0 {
		t.Errorf("expected empty depends_on, got %v", sv.DependsOn)
	}
}
