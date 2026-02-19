package rego

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	policyyaml "github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/policy/yaml"
)

// --- Evaluate error paths ---

func TestEvaluate_NonexistentPolicyFile(t *testing.T) {
	_, err := Evaluate("/nonexistent/policy.rego", Input{})
	if err == nil {
		t.Fatal("expected error for nonexistent policy file")
	}
	if !strings.Contains(err.Error(), "read rego policy") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluate_InvalidRegoSyntax(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "bad.rego")
	os.WriteFile(path, []byte("package bad\n!!!invalid syntax here"), 0o644)

	_, err := Evaluate(path, Input{})
	if err == nil {
		t.Fatal("expected error for invalid rego syntax")
	}
	if !strings.Contains(err.Error(), "prepare rego query") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluate_NoResultReturned(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "empty.rego")
	// Valid rego but under a different package — data.llmsa.gates.result won't exist
	os.WriteFile(path, []byte("package other\ndefault allow = true\n"), 0o644)

	_, err := Evaluate(path, Input{})
	if err == nil {
		t.Fatal("expected error for no result")
	}
	if !strings.Contains(err.Error(), "rego policy returned no result") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestEvaluate_ValidPolicyAllowAll(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "allow.rego")
	policy := `package llmsa.gates

result := {"allow": true, "violations": []}
`
	os.WriteFile(path, []byte(policy), 0o644)

	r, err := Evaluate(path, Input{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !r.Allow {
		t.Fatal("expected allow=true")
	}
	if len(r.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(r.Violations))
	}
}

func TestEvaluate_ValidPolicyDeny(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "deny.rego")
	policy := `package llmsa.gates

result := {"allow": false, "violations": ["missing prompt attestation", "missing eval attestation"]}
`
	os.WriteFile(path, []byte(policy), 0o644)

	r, err := Evaluate(path, Input{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Allow {
		t.Fatal("expected allow=false")
	}
	if len(r.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(r.Violations), r.Violations)
	}
}

// --- decodeResult edge cases ---

func TestDecodeResult_NonObjectInput(t *testing.T) {
	_, err := decodeResult("not an object")
	if err == nil {
		t.Fatal("expected error for non-object input")
	}
	if !strings.Contains(err.Error(), "rego result must be object") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDecodeResult_NilInput(t *testing.T) {
	_, err := decodeResult(nil)
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestDecodeResult_NumberInput(t *testing.T) {
	_, err := decodeResult(42)
	if err == nil {
		t.Fatal("expected error for number input")
	}
}

func TestDecodeResult_EmptyMap(t *testing.T) {
	r, err := decodeResult(map[string]any{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Allow {
		t.Fatal("expected allow=false for empty map (zero-value)")
	}
	if len(r.Violations) != 0 {
		t.Fatalf("expected 0 violations, got %d", len(r.Violations))
	}
}

func TestDecodeResult_WithViolationsSlice(t *testing.T) {
	r, err := decodeResult(map[string]any{
		"allow":      false,
		"violations": []any{"error1", "error2"},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if r.Allow {
		t.Fatal("expected allow=false")
	}
	if len(r.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d", len(r.Violations))
	}
}

func TestDecodeResult_WithViolationsMap(t *testing.T) {
	r, err := decodeResult(map[string]any{
		"allow":      false,
		"violations": map[string]any{"violation_a": true, "violation_b": true},
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(r.Violations) != 2 {
		t.Fatalf("expected 2 violations, got %d: %v", len(r.Violations), r.Violations)
	}
}

// --- decodeViolations edge cases ---

func TestDecodeViolations_NilInput(t *testing.T) {
	got := decodeViolations(nil)
	if len(got) != 0 {
		t.Fatalf("expected empty slice for nil, got %v", got)
	}
}

func TestDecodeViolations_EmptySlice(t *testing.T) {
	got := decodeViolations([]any{})
	if len(got) != 0 {
		t.Fatalf("expected empty slice, got %v", got)
	}
}

func TestDecodeViolations_SliceWithEmptyStrings(t *testing.T) {
	got := decodeViolations([]any{"real", "", "another"})
	if len(got) != 2 {
		t.Fatalf("expected 2 items (empty filtered), got %v", got)
	}
}

func TestDecodeViolations_SliceWithNonStrings(t *testing.T) {
	got := decodeViolations([]any{"real", 42, true})
	if len(got) != 1 || got[0] != "real" {
		t.Fatalf("expected ['real'], got %v", got)
	}
}

func TestDecodeViolations_MapStringAny(t *testing.T) {
	got := decodeViolations(map[string]any{"a": true, "b": true})
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %v", got)
	}
}

func TestDecodeViolations_MapStringAnyEmptyKey(t *testing.T) {
	got := decodeViolations(map[string]any{"": true, "real": true})
	if len(got) != 1 || got[0] != "real" {
		t.Fatalf("expected ['real'], got %v", got)
	}
}

func TestDecodeViolations_MapAnyAny(t *testing.T) {
	got := decodeViolations(map[any]any{"violation1": true, "violation2": true})
	if len(got) != 2 {
		t.Fatalf("expected 2 items, got %v", got)
	}
}

func TestDecodeViolations_MapAnyAny_NonStringKey(t *testing.T) {
	got := decodeViolations(map[any]any{123: true, "valid": true})
	if len(got) != 1 || got[0] != "valid" {
		t.Fatalf("expected ['valid'], got %v", got)
	}
}

func TestDecodeViolations_MapAnyAny_EmptyStringKey(t *testing.T) {
	got := decodeViolations(map[any]any{"": true, "real": true})
	if len(got) != 1 || got[0] != "real" {
		t.Fatalf("expected ['real'], got %v", got)
	}
}

func TestDecodeViolations_UnsupportedType(t *testing.T) {
	got := decodeViolations(42)
	if len(got) != 0 {
		t.Fatalf("expected empty slice for unsupported type, got %v", got)
	}
}

// --- BuildInput ---

func TestBuildInput_FieldMapping(t *testing.T) {
	input := BuildInput(
		policyyaml.Policy{PlaintextAllowlist: []string{"x"}},
		nil,
		[]string{"file.txt"},
	)
	if len(input.ChangedFiles) != 1 || input.ChangedFiles[0] != "file.txt" {
		t.Fatalf("unexpected changed_files: %v", input.ChangedFiles)
	}
	if len(input.PlaintextAllowlist) != 1 {
		t.Fatalf("unexpected plaintext_allowlist: %v", input.PlaintextAllowlist)
	}
}
