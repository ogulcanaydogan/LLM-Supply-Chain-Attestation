package verify

import (
	"strings"
	"testing"
)

const schemaDir = "../../schemas/v1"

func validPromptStatement() map[string]any {
	return map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "stmt-test",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2025-01-01T00:00:00Z",
		"generator": map[string]any{
			"name":    "llmsa",
			"version": "0.1.0",
			"git_sha": "abc123",
		},
		"subject": []any{
			map[string]any{
				"name": "system_prompt.txt",
				"uri":  "file://system_prompt.txt",
				"digest": map[string]any{
					"sha256": "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				"size_bytes": 42,
			},
		},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:abc",
			"system_prompt_digest": "sha256:def",
			"template_digests":     []any{"sha256:t1"},
			"tool_schema_digests":  []any{"sha256:s1"},
			"safety_policy_digest": "sha256:safety",
		},
		"privacy": map[string]any{
			"mode": "hash_only",
		},
	}
}

func TestVerifySchemas_ValidPrompt(t *testing.T) {
	err := VerifySchemas(schemaDir, validPromptStatement())
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifySchemas_MissingAttestationType(t *testing.T) {
	stmt := validPromptStatement()
	delete(stmt, "attestation_type")

	err := VerifySchemas(schemaDir, stmt)
	if err == nil {
		t.Fatal("expected error for missing attestation_type")
	}
}

func TestVerifySchemas_MissingPredicate(t *testing.T) {
	stmt := validPromptStatement()
	delete(stmt, "predicate")

	err := VerifySchemas(schemaDir, stmt)
	if err == nil {
		t.Fatal("expected error for missing predicate")
	}
}

func TestVerifySchemas_InvalidPredicate(t *testing.T) {
	stmt := validPromptStatement()
	stmt["predicate"] = map[string]any{
		"prompt_bundle_digest": "sha256:abc",
		// Missing system_prompt_digest which is required
	}

	err := VerifySchemas(schemaDir, stmt)
	if err == nil {
		t.Fatal("expected error for invalid predicate")
	}
	if !strings.Contains(err.Error(), "schema") {
		t.Errorf("error = %q", err)
	}
}

func TestVerifySchemas_InvalidBaseSchema(t *testing.T) {
	stmt := validPromptStatement()
	delete(stmt, "schema_version")
	delete(stmt, "statement_id")

	err := VerifySchemas(schemaDir, stmt)
	if err == nil {
		t.Fatal("expected error for invalid base schema")
	}
}
