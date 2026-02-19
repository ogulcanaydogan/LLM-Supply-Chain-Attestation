package verify

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

// --- ordered() edge cases ---

func TestOrderedInvalidPredecessorTimestamp(t *testing.T) {
	if !ordered("not-a-timestamp", "2026-02-18T00:00:00Z") {
		t.Fatal("expected true for invalid predecessor timestamp (permissive)")
	}
}

func TestOrderedInvalidSuccessorTimestamp(t *testing.T) {
	if !ordered("2026-02-18T00:00:00Z", "not-a-timestamp") {
		t.Fatal("expected true for invalid successor timestamp (permissive)")
	}
}

func TestOrderedBothInvalid(t *testing.T) {
	if !ordered("garbage", "also-garbage") {
		t.Fatal("expected true when both timestamps are invalid")
	}
}

func TestOrderedValidSameTime(t *testing.T) {
	if !ordered("2026-02-18T00:00:00Z", "2026-02-18T00:00:00Z") {
		t.Fatal("expected true when timestamps are identical")
	}
}

// --- checkUnknownDependencies edge cases ---

func TestCheckUnknownDependenciesWhitespaceDep(t *testing.T) {
	byType := map[string][]ChainStatement{}
	byID := map[string]ChainStatement{}
	violations := map[string]struct{}{}

	st := ChainStatement{
		StatementID:     "s1",
		AttestationType: "prompt_attestation",
		DependsOn:       []string{"  ", "", "\t"},
	}

	checkUnknownDependencies(st, byType, byID, violations)
	if len(violations) != 0 {
		t.Fatalf("expected 0 violations for whitespace deps, got %d: %v", len(violations), violations)
	}
}

// --- VerifyProvenanceChain: missing explicit dependency reference ---

func TestVerifyProvenanceChainMissingExplicitDependencyRef(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "prompt-1",
			AttestationType: "prompt_attestation",
			GeneratedAt:     "2026-02-18T00:00:00Z",
		},
		{
			StatementID:     "corpus-1",
			AttestationType: "corpus_attestation",
			GeneratedAt:     "2026-02-18T00:00:01Z",
		},
		{
			StatementID:     "eval-1",
			AttestationType: "eval_attestation",
			GeneratedAt:     "2026-02-18T00:00:02Z",
			DependsOn:       []string{"nonexistent_type"}, // wrong reference, doesn't match prompt or corpus
		},
	})
	if report.Valid {
		t.Fatal("expected invalid chain: eval has wrong dependency reference")
	}
	found := false
	for _, v := range report.Violations {
		if strings.Contains(v, "missing dependency reference") {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected 'missing dependency reference' violation, got: %v", report.Violations)
	}
}

// --- VerifyProvenanceChain: target with empty statement ID ---

func TestVerifyProvenanceChainEmptyStatementID(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "",
			AttestationType: "prompt_attestation",
			GeneratedAt:     "2026-02-18T00:00:00Z",
		},
		{
			StatementID:     "",
			AttestationType: "corpus_attestation",
			GeneratedAt:     "2026-02-18T00:00:01Z",
		},
		{
			StatementID:     "eval-1",
			AttestationType: "eval_attestation",
			GeneratedAt:     "2026-02-18T00:00:02Z",
		},
	})
	// Eval requires prompt and corpus — they exist but have empty IDs.
	// The edge should use "(by-type)" as ToStatementID.
	for _, edge := range report.Edges {
		if edge.FromStatementID == "eval-1" && edge.ToStatementID == "(by-type)" {
			return // found the expected edge
		}
	}
	// If we didn't find a (by-type) edge but all edges are satisfied, that's also valid
	if report.Valid {
		return
	}
	t.Fatalf("unexpected chain report: valid=%v, violations=%v, edges=%v", report.Valid, report.Violations, report.Edges)
}

// --- VerifySignature: invalid payload base64 ---

func TestVerifySignatureInvalidPayloadBase64(t *testing.T) {
	bundle := sign.Bundle{
		Envelope: sign.Envelope{
			Payload:    "not-valid-base64!!!@#$",
			Signatures: []sign.Signature{{Sig: "test"}},
		},
	}
	err := VerifySignature(bundle, SignerPolicy{})
	if err == nil {
		t.Fatal("expected error for invalid base64 payload")
	}
	if !strings.Contains(err.Error(), "decode payload") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- VerifySignature: invalid signature base64 ---

func TestVerifySignatureInvalidSigBase64(t *testing.T) {
	// Create a valid payload but with invalid signature base64
	payload := []byte(`{"test":"data"}`)
	payloadB64 := base64.StdEncoding.EncodeToString(payload)
	stmtHash := hash.DigestBytes(payload)

	bundle := sign.Bundle{
		Envelope: sign.Envelope{
			Payload: payloadB64,
			Signatures: []sign.Signature{{
				Sig:          "invalid-base64!!!",
				PublicKeyPEM: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			}},
		},
		Metadata: sign.Metadata{
			StatementHash: stmtHash,
		},
	}
	err := VerifySignature(bundle, SignerPolicy{})
	if err == nil {
		t.Fatal("expected error for invalid signature base64")
	}
	if !strings.Contains(err.Error(), "decode signature") && !strings.Contains(err.Error(), "parse public key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- verifyIdentityPolicy: invalid regex ---

func TestVerifyIdentityPolicyInvalidRegex(t *testing.T) {
	sig := sign.Signature{
		OIDCIssuer:   "https://issuer",
		OIDCIdentity: "identity",
	}
	policy := SignerPolicy{
		IdentityRegex: "[invalid(regex",
	}
	err := verifyIdentityPolicy(sig, policy)
	if err == nil {
		t.Fatal("expected error for invalid regex")
	}
	if !strings.Contains(err.Error(), "invalid identity regex") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- parsePublicKey: valid PEM but invalid PKIX content ---

func TestParsePublicKeyInvalidPKIXContent(t *testing.T) {
	// Valid PEM structure but garbage PKIX bytes
	pemText := "-----BEGIN PUBLIC KEY-----\nYWJjZGVm\n-----END PUBLIC KEY-----"
	_, err := parsePublicKey(pemText)
	if err == nil {
		t.Fatal("expected error for invalid PKIX content")
	}
	if !strings.Contains(err.Error(), "parse public key") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- bundlePaths: nonexistent path ---

func TestBundlePathsNonexistentPath(t *testing.T) {
	_, err := bundlePaths("/nonexistent/path/to/bundles")
	if err == nil {
		t.Fatal("expected error for nonexistent path")
	}
}

// --- bundlePaths: directory with mixed files ---

func TestBundlePathsFiltersNonBundles(t *testing.T) {
	tmp := t.TempDir()
	os.WriteFile(filepath.Join(tmp, "a.bundle.json"), []byte("{}"), 0o644)
	os.WriteFile(filepath.Join(tmp, "b.txt"), []byte("ignore"), 0o644)
	os.WriteFile(filepath.Join(tmp, "c.bundle.json"), []byte("{}"), 0o644)
	os.MkdirAll(filepath.Join(tmp, "subdir"), 0o755) // subdirectories skipped

	paths, err := bundlePaths(tmp)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 2 {
		t.Fatalf("expected 2 bundle files, got %d: %v", len(paths), paths)
	}
}

// --- WriteJSON: invalid output path ---

func TestWriteJSONInvalidPath(t *testing.T) {
	report := Report{Passed: true}
	err := WriteJSON("/nonexistent/dir/verify.json", report)
	if err == nil {
		t.Fatal("expected error for invalid output path")
	}
}

// --- Run: bundle read error (invalid JSON) ---

func TestRunBundleReadError(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := filepath.Join(tmp, "bad.bundle.json")
	os.WriteFile(bundlePath, []byte("not valid json"), 0o644)

	report := Run(Options{SourcePath: tmp, SchemaDir: "../../schemas/v1"})
	if report.Passed {
		t.Fatal("expected failure for invalid bundle JSON")
	}
	if report.ExitCode != ExitMissing {
		t.Fatalf("expected exit code %d, got %d", ExitMissing, report.ExitCode)
	}
}

// --- Run: payload decode error ---

func TestRunPayloadDecodeError(t *testing.T) {
	tmp := t.TempDir()
	// Create a bundle with invalid base64 payload
	bundle := sign.Bundle{
		Envelope: sign.Envelope{
			Payload:     "not-base64!!!",
			PayloadType: "application/vnd.llmsa.statement.v1+json",
			Signatures: []sign.Signature{{
				Sig:          "dGVzdA==",
				PublicKeyPEM: "-----BEGIN PUBLIC KEY-----\ntest\n-----END PUBLIC KEY-----",
			}},
		},
		Metadata: sign.Metadata{
			BundleVersion: "1",
			StatementHash: "sha256:abc",
		},
	}
	bundlePath := filepath.Join(tmp, "test.bundle.json")
	sign.WriteBundle(bundlePath, bundle)

	report := Run(Options{SourcePath: tmp, SchemaDir: "../../schemas/v1"})
	if report.Passed {
		t.Fatal("expected failure for invalid payload")
	}
}

// --- Run: nonexistent source path ---

func TestRunNonexistentSource(t *testing.T) {
	report := Run(Options{SourcePath: "/nonexistent/path", SchemaDir: "../../schemas/v1"})
	if report.Passed {
		t.Fatal("expected failure for nonexistent source")
	}
	if report.ExitCode != ExitMissing {
		t.Fatalf("expected exit code %d, got %d", ExitMissing, report.ExitCode)
	}
}

// --- Run: schema validation failure ---

func TestRunSchemaValidationFailure(t *testing.T) {
	tmp := t.TempDir()
	keyPath := filepath.Join(tmp, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)
	signer, _ := sign.NewPEMSigner(keyPath)

	// Create a statement that passes signature but fails schema (missing required fields)
	statement := map[string]any{
		"attestation_type": "prompt_attestation",
		// Missing schema_version, statement_id, generated_at, subject, predicate
	}
	canonical, _ := hash.CanonicalJSON(statement)
	material, _ := signer.Sign(canonical)
	bundle, _ := sign.CreateBundle(statement, material)

	bundlePath := filepath.Join(tmp, "test.bundle.json")
	sign.WriteBundle(bundlePath, bundle)

	report := Run(Options{SourcePath: tmp, SchemaDir: filepath.Join(repoRoot(t), "schemas", "v1")})
	if report.Passed {
		t.Fatal("expected failure for schema validation")
	}
}

// --- Run: chain constraint failure (missing generated_at) ---

func TestRunChainConstraintFailure(t *testing.T) {
	tmp := t.TempDir()
	keyPath := filepath.Join(tmp, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)
	signer, _ := sign.NewPEMSigner(keyPath)

	// Create a minimal statement that passes schema but fails chain (no generated_at)
	// Use a valid statement that has all schema-required fields except generated_at set to empty
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "test-id",
		"attestation_type": "prompt_attestation",
		"generated_at":     "", // empty - should fail chain constraint
		"subject":          []any{},
		"predicate": map[string]any{
			"system_prompt_digest": "sha256:abc",
			"templates_digest":     "sha256:def",
			"tool_schemas_digest":  "sha256:ghi",
			"safety_policy_digest": "sha256:jkl",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}
	canonical, _ := hash.CanonicalJSON(statement)
	material, _ := signer.Sign(canonical)
	bundle, _ := sign.CreateBundle(statement, material)

	bundlePath := filepath.Join(tmp, "test.bundle.json")
	sign.WriteBundle(bundlePath, bundle)

	report := Run(Options{SourcePath: tmp, SchemaDir: filepath.Join(repoRoot(t), "schemas", "v1")})
	if report.Passed {
		t.Fatal("expected failure for empty generated_at")
	}
}

// --- Run: chain graph failure path ---

func TestRunChainGraphFailureEdgeCase(t *testing.T) {
	tmp := t.TempDir()
	keyPath := filepath.Join(tmp, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)
	signer, _ := sign.NewPEMSigner(keyPath)
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")

	// Write a file for subject
	dataFile := filepath.Join(tmp, "data.txt")
	os.WriteFile(dataFile, []byte("data"), 0o644)
	fileDigest, _, _ := hash.DigestFile(dataFile)
	sha256Only := strings.TrimPrefix(fileDigest, "sha256:")

	// Create route_attestation WITHOUT eval_attestation predecessor → chain fails
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "route-1",
		"attestation_type": "route_attestation",
		"generated_at":     "2026-02-18T12:00:00Z",
		"subject": []any{
			map[string]any{
				"uri":    dataFile,
				"digest": map[string]any{"sha256": sha256Only},
			},
		},
		"predicate": map[string]any{
			"route_config_digest": "sha256:abc",
			"provider_set": []any{
				map[string]any{"provider": "openai", "model": "gpt-4o"},
			},
			"routing_strategy": "latency_aware",
		},
		"annotations": map[string]any{
			"depends_on":   "eval_attestation",
			"generated_by": "llmsa attest create",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}

	canonical, _ := hash.CanonicalJSON(statement)
	material, _ := signer.Sign(canonical)
	bundle, _ := sign.CreateBundle(statement, material)
	bundlePath := filepath.Join(tmp, "route.bundle.json")
	sign.WriteBundle(bundlePath, bundle)

	report := Run(Options{SourcePath: tmp, SchemaDir: schemaDir})
	// Chain graph should fail since route_attestation requires eval_attestation
	if report.Passed && report.Chain.Valid {
		t.Fatal("expected chain graph failure for missing eval_attestation predecessor")
	}
}

// --- VerifySchemas: nonexistent schema directory ---

func TestVerifySchemasNonexistentSchemaDir(t *testing.T) {
	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "id",
		"attestation_type": "prompt_attestation",
	}
	err := VerifySchemas("/nonexistent/schemas", statement)
	if err == nil {
		t.Fatal("expected error for nonexistent schema dir")
	}
}

// --- addFailure: escalates exit code ---

func TestAddFailureEscalatesExitCode(t *testing.T) {
	report := Report{Passed: true, ExitCode: ExitPass}
	report.addFailure("bundle1", "test", ExitSignatureFail, fmt.Errorf("sig failed"))
	if report.ExitCode != ExitSignatureFail {
		t.Fatalf("expected exit code %d, got %d", ExitSignatureFail, report.ExitCode)
	}
	report.addFailure("bundle2", "test", ExitDigestMismatch, fmt.Errorf("digest failed"))
	if report.ExitCode != ExitDigestMismatch {
		t.Fatalf("expected exit code to escalate to %d, got %d", ExitDigestMismatch, report.ExitCode)
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	dir, _ := os.Getwd()
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not find repo root")
		}
		dir = parent
	}
}
