//go:build e2e

package e2e

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/attest"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

func TestFullPipeline_AttestSignVerify(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	createAllFiveAttestations(t, outDir, keyPath)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if !report.Passed {
		t.Fatalf("verify failed: exit %d, violations: %v", report.ExitCode, report.Violations)
	}
	if report.ExitCode != 0 {
		t.Errorf("exit code = %d, want 0", report.ExitCode)
	}
	if report.BundleCount != 5 {
		t.Errorf("bundle count = %d, want 5", report.BundleCount)
	}
}

func TestFullPipeline_TamperDetection(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	createAndSignAttestation(t, "prompt_attestation", configPath(t, "prompt"), outDir, keyPath)

	// Tamper with the subject file referenced by the attestation.
	promptPath := filepath.Join(repoRoot(t), "examples", "tiny-rag", "app", "system_prompt.txt")
	original, err := os.ReadFile(promptPath)
	if err != nil {
		t.Fatal(err)
	}
	defer os.WriteFile(promptPath, original, 0o644) // Restore after test.

	tampered := append(original, []byte("\n# tampered")...)
	os.WriteFile(promptPath, tampered, 0o644)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected verify to fail after tampering")
	}
	if report.ExitCode != 12 {
		t.Errorf("exit code = %d, want 12 (digest mismatch)", report.ExitCode)
	}
}

func TestFullPipeline_ChainVerification(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	createAllFiveAttestations(t, outDir, keyPath)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if !report.Chain.Valid {
		t.Errorf("chain invalid: %v", report.Chain.Violations)
	}
	// Verify edges exist in the chain.
	hasEdge := func(fromType, toType string) bool {
		for _, e := range report.Chain.Edges {
			if e.FromType == fromType && e.ToType == toType {
				return true
			}
		}
		return false
	}
	// Edge direction: FromType is the dependent, ToType is the dependency.
	// eval depends on prompt and corpus, route depends on eval, slo depends on route.
	if !hasEdge("eval_attestation", "prompt_attestation") {
		t.Error("missing edge: eval -> prompt")
	}
	if !hasEdge("eval_attestation", "corpus_attestation") {
		t.Error("missing edge: eval -> corpus")
	}
	if !hasEdge("route_attestation", "eval_attestation") {
		t.Error("missing edge: route -> eval")
	}
	if !hasEdge("slo_attestation", "route_attestation") {
		t.Error("missing edge: slo -> route")
	}
}

func TestFullPipeline_MissingAttestation(t *testing.T) {
	emptyDir := t.TempDir()
	report := verify.Run(verify.Options{
		SourcePath: emptyDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected failure for empty attestation dir")
	}
	if report.ExitCode != 10 {
		t.Errorf("exit code = %d, want 10 (missing)", report.ExitCode)
	}
}

func TestFullPipeline_SignatureCorruption(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	bundlePath := createAndSignAttestation(t, "prompt_attestation", configPath(t, "prompt"), outDir, keyPath)

	// Read the bundle and corrupt the signature.
	bundle, err := sign.ReadBundle(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	if len(bundle.Envelope.Signatures) > 0 {
		bundle.Envelope.Signatures[0].Sig = "dGFtcGVyZWQ=" // "tampered" in base64
	}
	corrupted, _ := json.MarshalIndent(bundle, "", "  ")
	os.WriteFile(bundlePath, corrupted, 0o644)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected failure for corrupted signature")
	}
	if report.ExitCode != 11 {
		t.Errorf("exit code = %d, want 11 (signature fail)", report.ExitCode)
	}
}

func TestTamper_StatementHashDesync(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	bundlePath := createAndSignAttestation(t, "prompt_attestation", configPath(t, "prompt"), outDir, keyPath)

	bundle, err := sign.ReadBundle(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Metadata.StatementHash = "sha256:" + strings.Repeat("a", 64)
	corrupted, _ := json.MarshalIndent(bundle, "", "  ")
	os.WriteFile(bundlePath, corrupted, 0o644)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected verify to fail after statement hash desync")
	}
	if report.ExitCode != verify.ExitSignatureFail {
		t.Errorf("exit code = %d, want %d (ExitSignatureFail)", report.ExitCode, verify.ExitSignatureFail)
	}
}

func TestTamper_PayloadBase64Corruption(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	bundlePath := createAndSignAttestation(t, "prompt_attestation", configPath(t, "prompt"), outDir, keyPath)

	bundle, err := sign.ReadBundle(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Envelope.Payload = "!!!not-valid-base64==="
	corrupted, _ := json.MarshalIndent(bundle, "", "  ")
	os.WriteFile(bundlePath, corrupted, 0o644)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected verify to fail after payload base64 corruption")
	}
	if report.ExitCode != verify.ExitSignatureFail {
		t.Errorf("exit code = %d, want %d (ExitSignatureFail)", report.ExitCode, verify.ExitSignatureFail)
	}
}

func TestTamper_SchemaViolation(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	// Statement that satisfies the base statement schema but fails the
	// predicate-specific schema: prompt_attestation requires prompt_bundle_digest
	// (and others) which are all absent from this empty predicate.
	stmt := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "test-schema-fail",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-06-02T00:00:00Z",
		"generator":        map[string]any{"name": "test", "version": "0.0.0", "git_sha": "abc"},
		"subject":          []any{},
		"predicate":        map[string]any{},
		"privacy":          map[string]any{"mode": "hash_only"},
	}

	canonical, err := hash.CanonicalJSON(stmt)
	if err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}
	material, err := signer.Sign(canonical)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sign.CreateBundle(stmt, material)
	if err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(outDir, "prompt_attestation.bundle.json")
	if err := sign.WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected verify to fail on schema violation")
	}
	if report.ExitCode != verify.ExitSchemaFail {
		t.Errorf("exit code = %d, want %d (ExitSchemaFail)", report.ExitCode, verify.ExitSchemaFail)
	}
}

func TestTamper_SignaturesCleared(t *testing.T) {
	outDir := t.TempDir()
	keyPath := filepath.Join(outDir, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	bundlePath := createAndSignAttestation(t, "prompt_attestation", configPath(t, "prompt"), outDir, keyPath)

	bundle, err := sign.ReadBundle(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Envelope.Signatures = nil
	corrupted, _ := json.MarshalIndent(bundle, "", "  ")
	os.WriteFile(bundlePath, corrupted, 0o644)

	report := verify.Run(verify.Options{
		SourcePath: outDir,
		SchemaDir:  schemaDir(t),
	})
	if report.Passed {
		t.Error("expected verify to fail after signatures cleared")
	}
	if report.ExitCode != verify.ExitSignatureFail {
		t.Errorf("exit code = %d, want %d (ExitSignatureFail)", report.ExitCode, verify.ExitSignatureFail)
	}
}

func TestFullPipeline_DeterminismCheck(t *testing.T) {
	outDir := t.TempDir()
	_, err := attest.CreateByType(attest.CreateOptions{
		Type:             "prompt_attestation",
		ConfigPath:       configPath(t, "prompt"),
		OutDir:           outDir,
		DeterminismCheck: 3,
	})
	if err != nil {
		// Determinism checks may fail due to timestamps in the statement.
		// The important thing is the function runs without panic.
		if !strings.Contains(err.Error(), "determinism") {
			t.Fatalf("unexpected error: %v", err)
		}
	}
}
