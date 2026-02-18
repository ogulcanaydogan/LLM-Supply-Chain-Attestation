package schema

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestValidateStatementSchema(t *testing.T) {
	doc := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "abc",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-02-17T20:10:11Z",
		"generator": map[string]any{
			"name":    "llmsa",
			"version": "0.1.0",
			"git_sha": "local",
		},
		"subject": []any{map[string]any{
			"name": "system_prompt.txt",
			"uri":  "examples/tiny-rag/app/system_prompt.txt",
			"digest": map[string]any{
				"sha256": "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			},
			"size_bytes": 10,
		}},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:abc",
			"system_prompt_digest": "sha256:abc",
			"template_digests":     []any{"sha256:abc"},
			"tool_schema_digests":  []any{"sha256:abc"},
			"safety_policy_digest": "sha256:abc",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}
	errs, err := Validate("../../schemas/v1/statement.schema.json", doc)
	if err != nil {
		t.Fatal(err)
	}
	if len(errs) != 0 {
		t.Fatalf("schema should pass: %v", errs)
	}
}

func TestValidateInvalidDocument(t *testing.T) {
	doc := map[string]any{
		"attestation_type": "prompt_attestation",
	}
	errs, err := Validate("../../schemas/v1/statement.schema.json", doc)
	if err != nil {
		t.Fatalf("unexpected validation error: %v", err)
	}
	if len(errs) == 0 {
		t.Fatal("expected schema violations")
	}
}

func TestValidateMissingSchemaFile(t *testing.T) {
	_, err := Validate(filepath.Join(t.TempDir(), "missing.schema.json"), map[string]any{})
	if err == nil {
		t.Fatal("expected schema loader error")
	}
	if !strings.Contains(err.Error(), "validate") {
		t.Fatalf("unexpected error: %v", err)
	}
}
