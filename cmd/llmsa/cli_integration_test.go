package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

// --- Init Command ---

func TestInitCommand_CreatesConfigAndStore(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	cmd := newInitCommand()
	if err := cmd.Execute(); err != nil {
		t.Fatalf("init: %v", err)
	}

	// Verify all expected files and dirs exist.
	for _, p := range []string{
		".llmsa/attestations",
		"llmsa.yaml",
		"policy/examples/mvp-gates.yaml",
		".llmsa/dev_ed25519.pem",
	} {
		if _, err := os.Stat(p); err != nil {
			t.Errorf("init missing %q: %v", p, err)
		}
	}
}

func TestInitCommand_Idempotent(t *testing.T) {
	orig, _ := os.Getwd()
	tmp := t.TempDir()
	os.Chdir(tmp)
	defer os.Chdir(orig)

	if err := newInitCommand().Execute(); err != nil {
		t.Fatal(err)
	}
	// Running again should not error.
	if err := newInitCommand().Execute(); err != nil {
		t.Fatalf("second init: %v", err)
	}
}

// --- Attest Command ---

func TestAttestCreateCommand_PromptAttestation(t *testing.T) {
	root := repoRoot(t)
	outDir := t.TempDir()

	cmd := newAttestCommand()
	cmd.SetArgs([]string{
		"create",
		"--type", "prompt_attestation",
		"--config", filepath.Join(root, "examples", "tiny-rag", "configs", "prompt.yaml"),
		"--out", outDir,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("attest create: %v", err)
	}

	// Should produce a statement file.
	entries, _ := os.ReadDir(outDir)
	found := false
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "statement_") && strings.HasSuffix(e.Name(), ".json") {
			found = true
		}
	}
	if !found {
		t.Error("expected statement_*.json file in output")
	}
}

func TestAttestCreateCommand_MissingFlags(t *testing.T) {
	cmd := newAttestCommand()
	cmd.SetArgs([]string{"create"})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for missing --type and --config")
	}
}

func TestAttestCreateCommand_InvalidType(t *testing.T) {
	cmd := newAttestCommand()
	cmd.SetArgs([]string{
		"create",
		"--type", "nonexistent_attestation",
		"--config", "/dev/null",
		"--out", t.TempDir(),
	})
	if err := cmd.Execute(); err == nil {
		t.Fatal("expected error for invalid attestation type")
	}
}

// --- Sign Command ---

func TestSignCommand_PEMProvider(t *testing.T) {
	root := repoRoot(t)
	tmp := t.TempDir()

	// First create a statement to sign.
	outDir := t.TempDir()
	createCmd := newAttestCommand()
	createCmd.SetArgs([]string{
		"create",
		"--type", "prompt_attestation",
		"--config", filepath.Join(root, "examples", "tiny-rag", "configs", "prompt.yaml"),
		"--out", outDir,
	})
	if err := createCmd.Execute(); err != nil {
		t.Fatalf("attest create: %v", err)
	}

	// Find the statement file.
	entries, _ := os.ReadDir(outDir)
	var stmtPath string
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), "statement_") {
			stmtPath = filepath.Join(outDir, e.Name())
			break
		}
	}
	if stmtPath == "" {
		t.Fatal("no statement file found")
	}

	// Generate a PEM key before signing.
	keyPath := filepath.Join(tmp, "key.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatalf("generate key: %v", err)
	}
	signCmd := newSignCommand()
	signCmd.SetArgs([]string{
		"--in", stmtPath,
		"--provider", "pem",
		"--key", keyPath,
		"--out", tmp,
	})
	if err := signCmd.Execute(); err != nil {
		t.Fatalf("sign command: %v", err)
	}

	// Verify a bundle file was produced.
	entries, _ = os.ReadDir(tmp)
	found := false
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".bundle.json") {
			found = true
		}
	}
	if !found {
		t.Error("expected .bundle.json in output")
	}
}

// --- Report Command ---

func TestReportCommand_GeneratesMarkdown(t *testing.T) {
	tmp := t.TempDir()

	// Create a minimal verify report.
	report := verify.Report{
		Passed:      true,
		ExitCode:    0,
		BundleCount: 1,
		Checks: []verify.CheckResult{
			{Bundle: "test.bundle.json", Check: "signature", Passed: true},
		},
	}
	reportPath := filepath.Join(tmp, "report.json")
	raw, _ := json.MarshalIndent(report, "", "  ")
	os.WriteFile(reportPath, raw, 0o644)

	outPath := filepath.Join(tmp, "report.md")
	cmd := newReportCommand()
	cmd.SetArgs([]string{
		"--in", reportPath,
		"--out", outPath,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("report command: %v", err)
	}

	data, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty report output")
	}
}

// --- Verify Command Local Source ---

func TestVerifyCommandLocal(t *testing.T) {
	tmp := t.TempDir()
	bundlePath := writeSignedPromptBundle(t, tmp, "hash_only")
	schemaDir := filepath.Join(repoRoot(t), "schemas", "v1")
	outPath := filepath.Join(tmp, "verify-local.json")

	cmd := newVerifyCommand()
	cmd.SetArgs([]string{
		"--source", "local",
		"--attestations", filepath.Dir(bundlePath),
		"--schema-dir", schemaDir,
		"--format", "json",
		"--out", outPath,
	})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("verify local: %v", err)
	}

	raw, _ := os.ReadFile(outPath)
	var r verify.Report
	json.Unmarshal(raw, &r)
	if !r.Passed {
		t.Errorf("expected verify pass, got exit %d: %v", r.ExitCode, r.Violations)
	}
}

// --- Webhook Command ---

func TestWebhookCommand_Help(t *testing.T) {
	cmd := newWebhookCommand()
	cmd.SetArgs([]string{"--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("webhook help: %v", err)
	}
}

func TestWebhookServeCommand_Help(t *testing.T) {
	cmd := newWebhookCommand()
	cmd.SetArgs([]string{"serve", "--help"})
	if err := cmd.Execute(); err != nil {
		t.Fatalf("webhook serve help: %v", err)
	}
}
