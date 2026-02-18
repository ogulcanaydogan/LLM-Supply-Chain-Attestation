package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

func TestVerifyCommandWithOCISource(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")
	outPath := filepath.Join(tmp, "verify.json")

	originalPull := ociPullFunc
	t.Cleanup(func() { ociPullFunc = originalPull })

	called := false
	ociPullFunc = func(_ string, outPath string) error {
		called = true
		raw, err := os.ReadFile(bundlePath)
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, raw, 0o644)
	}

	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "oci",
		"--attestations", "ghcr.io/example/repo/attestations:test",
		"--schema-dir", schemaDir,
		"--format", "json",
		"--out", outPath,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify command failed: %v", err)
	}
	if !called {
		t.Fatalf("expected oci pull function to be called")
	}

	raw, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	var r verify.Report
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatal(err)
	}
	if !r.Passed || r.ExitCode != verify.ExitPass {
		t.Fatalf("expected verify pass, got %+v", r)
	}
}

func TestGateCommandWithOCISource(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "plaintext_explicit")
	policyPath := filepath.Join(tmp, "policy.yaml")
	if err := os.WriteFile(policyPath, []byte(`version: 1
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/.+/.+/.github/workflows/.+@refs/.+$'
plaintext_allowlist: []
gates: []
`), 0o644); err != nil {
		t.Fatal(err)
	}

	originalPull := ociPullFunc
	t.Cleanup(func() { ociPullFunc = originalPull })

	ociPullFunc = func(_ string, outPath string) error {
		raw, err := os.ReadFile(bundlePath)
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, raw, 0o644)
	}

	cmd := newGateCommand()
	cmd.SetArgs([]string{
		"--source", "oci",
		"--policy", policyPath,
		"--attestations", "ghcr.io/example/repo/attestations:test",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected policy failure for plaintext exposure")
	}

	var ce cliError
	if !errors.As(err, &ce) {
		t.Fatalf("expected cliError, got %T: %v", err, err)
	}
	if ce.code != verify.ExitPolicyFail {
		t.Fatalf("expected exit code %d, got %d", verify.ExitPolicyFail, ce.code)
	}
}

func TestGateCommandWithRegoEngine(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "plaintext_explicit")
	policyPath := filepath.Join(tmp, "policy.yaml")
	if err := os.WriteFile(policyPath, []byte(`version: 1
oidc_issuer: https://token.actions.githubusercontent.com
identity_regex: '^https://github\.com/.+/.+/.github/workflows/.+@refs/.+$'
plaintext_allowlist: []
gates: []
`), 0o644); err != nil {
		t.Fatal(err)
	}
	regoPath := filepath.Join(repoRoot(t), "policy", "examples", "rego-gates.rego")

	originalPull := ociPullFunc
	t.Cleanup(func() { ociPullFunc = originalPull })
	ociPullFunc = func(_ string, outPath string) error {
		raw, err := os.ReadFile(bundlePath)
		if err != nil {
			return err
		}
		return os.WriteFile(outPath, raw, 0o644)
	}

	cmd := newGateCommand()
	cmd.SetArgs([]string{
		"--engine", "rego",
		"--rego-policy", regoPath,
		"--source", "oci",
		"--policy", policyPath,
		"--attestations", "ghcr.io/example/repo/attestations:test",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatalf("expected policy failure for plaintext exposure")
	}

	var ce cliError
	if !errors.As(err, &ce) {
		t.Fatalf("expected cliError, got %T: %v", err, err)
	}
	if ce.code != verify.ExitPolicyFail {
		t.Fatalf("expected exit code %d, got %d", verify.ExitPolicyFail, ce.code)
	}
}

func TestDefaultBundlePathContract(t *testing.T) {
	statement := map[string]any{
		"attestation_type": "prompt_attestation",
		"statement_id":     "abc-123",
		"generator": map[string]any{
			"git_sha": "deadbeef",
		},
	}
	got := defaultBundlePath(filepath.Join("/tmp", "statement.json"), statement)
	want := filepath.Join("/tmp", "attestation_prompt_attestation_deadbeef_abc-123.bundle.json")
	if got != want {
		t.Fatalf("unexpected bundle path: got %q want %q", got, want)
	}
}

func TestDefaultBundlePathSanitizesGitSHA(t *testing.T) {
	statement := map[string]any{
		"attestation_type": "prompt_attestation",
		"statement_id":     "stmt-1",
		"generator": map[string]any{
			"git_sha": "refs/heads/main:sha with spaces",
		},
	}
	got := filepath.Base(defaultBundlePath(filepath.Join("/tmp", "statement.json"), statement))
	if strings.ContainsAny(got, " /:") {
		t.Fatalf("bundle name should sanitize git sha, got %q", got)
	}
	if !strings.HasPrefix(got, "attestation_prompt_attestation_refs_heads_main_sha_with_spaces_") {
		t.Fatalf("unexpected sanitized prefix: %q", got)
	}
}

func TestPublishCommandUsesOCIPublisher(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := filepath.Join(tmp, "bundle.json")
	if err := os.WriteFile(bundlePath, []byte(`{"envelope":{"payloadType":"application/json","payload":"e30=","signatures":[]},"metadata":{"bundle_version":"1","created_at":"2026-01-01T00:00:00Z","statement_hash":"abc"}}`), 0o644); err != nil {
		t.Fatal(err)
	}

	original := ociPublishFunc
	t.Cleanup(func() { ociPublishFunc = original })

	called := false
	ociPublishFunc = func(inPath string, ociRef string) (string, error) {
		called = true
		if inPath != bundlePath {
			t.Fatalf("unexpected in path: %s", inPath)
		}
		if ociRef != "ghcr.io/acme/llmsa/attestations:test" {
			t.Fatalf("unexpected ref: %s", ociRef)
		}
		return "ghcr.io/acme/llmsa/attestations@sha256:deadbeef", nil
	}

	cmd := newPublishCommand()
	cmd.SetArgs([]string{
		"--in", bundlePath,
		"--oci", "ghcr.io/acme/llmsa/attestations:test",
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("publish command failed: %v", err)
	}
	if !called {
		t.Fatalf("expected oci publisher to be called")
	}
}

func writeSignedPromptBundle(t *testing.T, dir string, privacyMode string) string {
	t.Helper()

	subjectPath := filepath.Join(dir, "subject.txt")
	if err := os.WriteFile(subjectPath, []byte("subject-bytes"), 0o644); err != nil {
		t.Fatal(err)
	}
	digest, size, err := hash.DigestFile(subjectPath)
	if err != nil {
		t.Fatal(err)
	}

	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "test-statement-1",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-02-17T20:10:11Z",
		"generator": map[string]any{
			"name":    "llmsa",
			"version": "0.1.0",
			"git_sha": "local",
		},
		"subject": []any{
			map[string]any{
				"name": "subject.txt",
				"uri":  subjectPath,
				"digest": map[string]any{
					"sha256": strings.TrimPrefix(digest, "sha256:"),
				},
				"size_bytes": size,
			},
		},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:bundle",
			"system_prompt_digest": "sha256:system",
			"template_digests":     []string{"sha256:template"},
			"tool_schema_digests":  []string{"sha256:tool"},
			"safety_policy_digest": "sha256:safety",
		},
		"privacy": map[string]any{
			"mode": privacyMode,
		},
	}

	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		t.Fatal(err)
	}
	keyPath := filepath.Join(dir, "dev.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
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
	bundle, err := sign.CreateBundle(statement, material)
	if err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(dir, "prompt.bundle.json")
	if err := sign.WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}
	return bundlePath
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", ".."))
}
