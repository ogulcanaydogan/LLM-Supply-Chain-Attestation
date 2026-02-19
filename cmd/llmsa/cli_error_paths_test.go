package main

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

// --- Root Command ---

func TestNewRootCommand_SubcommandRegistration(t *testing.T) {
	root := newRootCommand()
	cmds := root.Commands()
	want := map[string]bool{
		"init": false, "attest": false, "sign": false, "publish": false,
		"verify": false, "gate": false, "report": false, "demo": false, "webhook": false,
	}
	for _, c := range cmds {
		want[c.Name()] = true
	}
	for name, found := range want {
		if !found {
			t.Errorf("missing subcommand: %s", name)
		}
	}
}

// --- Sign Command Error Paths ---

func TestSignCommand_MissingInFlag(t *testing.T) {
	cmd := newSignCommand()
	cmd.SetArgs([]string{"--provider", "pem"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --in")
	}
	if !strings.Contains(err.Error(), "--in is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignCommand_NonExistentInput(t *testing.T) {
	cmd := newSignCommand()
	cmd.SetArgs([]string{"--in", "/nonexistent/statement.json", "--provider", "pem", "--key", "/dev/null"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent input")
	}
}

func TestSignCommand_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	badJSON := filepath.Join(tmp, "bad.json")
	os.WriteFile(badJSON, []byte("{not valid json"), 0o644)

	cmd := newSignCommand()
	cmd.SetArgs([]string{"--in", badJSON, "--provider", "pem", "--key", "/dev/null"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestSignCommand_PEMMissingKey(t *testing.T) {
	tmp := t.TempDir()
	stmt := filepath.Join(tmp, "stmt.json")
	os.WriteFile(stmt, []byte(`{"schema_version":"1.0.0","statement_id":"test","attestation_type":"prompt_attestation"}`), 0o644)

	cmd := newSignCommand()
	cmd.SetArgs([]string{"--in", stmt, "--provider", "pem"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --key")
	}
	if !strings.Contains(err.Error(), "--key is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignCommand_UnsupportedProvider(t *testing.T) {
	tmp := t.TempDir()
	stmt := filepath.Join(tmp, "stmt.json")
	os.WriteFile(stmt, []byte(`{"schema_version":"1.0.0"}`), 0o644)

	cmd := newSignCommand()
	cmd.SetArgs([]string{"--in", stmt, "--provider", "aws"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported provider")
	}
	if !strings.Contains(err.Error(), "unsupported provider") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignCommand_KMSProvider(t *testing.T) {
	tmp := t.TempDir()
	stmt := filepath.Join(tmp, "stmt.json")
	os.WriteFile(stmt, []byte(`{"schema_version":"1.0.0"}`), 0o644)

	cmd := newSignCommand()
	cmd.SetArgs([]string{"--in", stmt, "--provider", "kms"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for KMS provider (not implemented)")
	}
	if !strings.Contains(err.Error(), "not implemented") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestSignCommand_AutoOutputPath(t *testing.T) {
	root := repoRoot(t)
	tmp := t.TempDir()

	// Create a valid statement
	outDir := t.TempDir()
	createCmd := newAttestCommand()
	createCmd.SetArgs([]string{
		"create",
		"--type", "prompt_attestation",
		"--config", filepath.Join(root, "examples", "tiny-rag", "configs", "prompt.yaml"),
		"--out", outDir,
	})
	if err := createCmd.Execute(); err != nil {
		t.Fatal(err)
	}

	entries, _ := os.ReadDir(outDir)
	var stmtPath string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "statement_") {
			stmtPath = filepath.Join(outDir, e.Name())
			break
		}
	}

	keyPath := filepath.Join(tmp, "key.pem")
	sign.GeneratePEMPrivateKey(keyPath)

	// Sign without --out (auto-generate path)
	signCmd := newSignCommand()
	signCmd.SetArgs([]string{"--in", stmtPath, "--provider", "pem", "--key", keyPath})
	if err := signCmd.Execute(); err != nil {
		t.Fatalf("sign auto-path: %v", err)
	}

	// Verify bundle was created in the statement's directory
	outEntries, _ := os.ReadDir(outDir)
	found := false
	for _, e := range outEntries {
		if strings.HasSuffix(e.Name(), ".bundle.json") {
			found = true
		}
	}
	if !found {
		t.Error("expected .bundle.json in statement's directory when --out is omitted")
	}
}

// --- Verify Command Error Paths ---

func TestVerifyCommand_UnsupportedSource(t *testing.T) {
	cmd := newVerifyCommand()
	cmd.SetArgs([]string{"--source", "s3", "--attestations", "ref"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported source")
	}
	if !strings.Contains(err.Error(), "unsupported source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyCommand_UnsupportedFormat(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")

	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "local",
		"--attestations", filepath.Dir(bundlePath),
		"--schema-dir", schemaDir,
		"--format", "xml",
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported format")
	}
	if !strings.Contains(err.Error(), "unsupported format") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestVerifyCommand_MarkdownFormat(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")
	outPath := filepath.Join(tmp, "verify.md")

	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "local",
		"--attestations", filepath.Dir(bundlePath),
		"--schema-dir", schemaDir,
		"--format", "md",
		"--out", outPath,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify md format: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty markdown output")
	}
}

func TestVerifyCommand_FailedVerification(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")

	// Corrupt the bundle's signature
	bundle, err := sign.ReadBundle(bundlePath)
	if err != nil {
		t.Fatal(err)
	}
	bundle.Envelope.Signatures[0].Sig = "AAAA"
	if err := sign.WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}

	outPath := filepath.Join(tmp, "verify-fail.json")
	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "local",
		"--attestations", filepath.Dir(bundlePath),
		"--schema-dir", schemaDir,
		"--format", "json",
		"--out", outPath,
	})
	err = cmd.Execute()
	if err == nil {
		t.Fatal("expected error for failed verification")
	}
	var ce cliError
	if !errors.As(err, &ce) {
		t.Fatalf("expected cliError, got %T: %v", err, err)
	}
	if ce.code == 0 {
		t.Fatal("expected non-zero exit code")
	}
}

func TestVerifyCommand_OCIPullError(t *testing.T) {
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")

	originalPull := ociPullFunc
	t.Cleanup(func() { ociPullFunc = originalPull })
	ociPullFunc = func(_ string, _ string) error {
		return errors.New("mock OCI pull failed")
	}

	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "oci",
		"--attestations", "ghcr.io/test/repo:tag",
		"--schema-dir", schemaDir,
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for OCI pull failure")
	}
	if !strings.Contains(err.Error(), "mock OCI pull failed") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- Gate Command Error Paths ---

func TestGateCommand_MissingPolicy(t *testing.T) {
	cmd := newGateCommand()
	cmd.SetArgs([]string{"--attestations", t.TempDir()})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --policy")
	}
	if !strings.Contains(err.Error(), "--policy is required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateCommand_UnsupportedSource(t *testing.T) {
	tmp := t.TempDir()
	policyPath := filepath.Join(tmp, "p.yaml")
	os.WriteFile(policyPath, []byte("version: 1\ngates: []\n"), 0o644)

	cmd := newGateCommand()
	cmd.SetArgs([]string{"--source", "s3", "--policy", policyPath})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported source")
	}
	if !strings.Contains(err.Error(), "unsupported source") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateCommand_UnsupportedEngine(t *testing.T) {
	tmp := t.TempDir()
	policyPath := filepath.Join(tmp, "policy.yaml")
	os.WriteFile(policyPath, []byte("version: 1\ngates: []\n"), 0o644)
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")

	cmd := newGateCommand()
	cmd.SetArgs([]string{
		"--engine", "wasm",
		"--policy", policyPath,
		"--attestations", filepath.Dir(bundlePath),
	})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for unsupported engine")
	}
	if !strings.Contains(err.Error(), "unsupported policy engine") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestGateCommand_InvalidPolicyFile(t *testing.T) {
	cmd := newGateCommand()
	cmd.SetArgs([]string{"--policy", "/nonexistent/policy.yaml", "--attestations", t.TempDir()})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for invalid policy file")
	}
}

// --- Report Command Error Paths ---

func TestReportCommand_MissingFlags(t *testing.T) {
	cmd := newReportCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --in and --out")
	}
	if !strings.Contains(err.Error(), "--in and --out are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestReportCommand_NonExistentInput(t *testing.T) {
	cmd := newReportCommand()
	cmd.SetArgs([]string{"--in", "/nonexistent/report.json", "--out", "/tmp/out.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for nonexistent input")
	}
}

func TestReportCommand_MalformedJSON(t *testing.T) {
	tmp := t.TempDir()
	badJSON := filepath.Join(tmp, "bad.json")
	os.WriteFile(badJSON, []byte("{not valid"), 0o644)

	cmd := newReportCommand()
	cmd.SetArgs([]string{"--in", badJSON, "--out", filepath.Join(tmp, "out.md")})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestReportCommand_WriteError(t *testing.T) {
	tmp := t.TempDir()
	reportPath := filepath.Join(tmp, "report.json")
	r := verify.Report{Passed: true, ExitCode: 0, BundleCount: 1}
	raw, _ := json.MarshalIndent(r, "", "  ")
	os.WriteFile(reportPath, raw, 0o644)

	cmd := newReportCommand()
	cmd.SetArgs([]string{"--in", reportPath, "--out", "/nonexistent/dir/report.md"})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for write failure")
	}
}

// --- Publish Command Error Paths ---

func TestPublishCommand_MissingFlags(t *testing.T) {
	cmd := newPublishCommand()
	cmd.SetArgs([]string{})
	err := cmd.Execute()
	if err == nil {
		t.Fatal("expected error for missing --in and --oci")
	}
	if !strings.Contains(err.Error(), "--in and --oci are required") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// --- CLI helper functions ---

func TestSplitCSV(t *testing.T) {
	tests := []struct {
		input string
		want  int
	}{
		{"a,b,c", 3},
		{"a, b, c", 3},
		{"single", 1},
		{"", 0},
		{"a,,b", 2},
	}
	for _, tt := range tests {
		got := splitCSV(tt.input)
		if len(got) != tt.want {
			t.Errorf("splitCSV(%q) = %d items, want %d", tt.input, len(got), tt.want)
		}
	}
}

func TestAsString(t *testing.T) {
	if asString("hello") != "hello" {
		t.Error("expected 'hello'")
	}
	if asString(nil) != "" {
		t.Error("expected empty for nil")
	}
	if asString(42) != "" {
		t.Error("expected empty for non-string")
	}
}

func TestFileExists(t *testing.T) {
	tmp := t.TempDir()
	f := filepath.Join(tmp, "exists.txt")
	os.WriteFile(f, []byte("x"), 0o644)
	if !fileExists(f) {
		t.Error("expected true for existing file")
	}
	if fileExists("/nonexistent/file") {
		t.Error("expected false for nonexistent file")
	}
}

func TestCliError_ErrorString(t *testing.T) {
	ce := cliError{code: 11, err: errors.New("sig fail")}
	if ce.Error() != "sig fail" {
		t.Fatalf("unexpected error string: %q", ce.Error())
	}
}
