package report

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

func sampleReport() verify.Report {
	return verify.Report{
		Passed:      true,
		ExitCode:    0,
		BundleCount: 2,
		Checks: []verify.CheckResult{
			{Bundle: "b1.bundle.json", Check: "signature", Passed: true, Message: "ok"},
			{Bundle: "b1.bundle.json", Check: "digest", Passed: true, Message: "ok"},
			{Bundle: "b2.bundle.json", Check: "signature", Passed: true, Message: "ok"},
		},
		Violations: nil,
		Statements: []verify.StatementSummary{
			{
				AttestationType: "prompt_attestation",
				StatementID:     "stmt-1",
				PrivacyMode:     "hash_only",
			},
			{
				AttestationType: "eval_attestation",
				StatementID:     "stmt-2",
				PrivacyMode:     "hash_only",
				DependsOn:       []string{"prompt_attestation", "corpus_attestation"},
			},
		},
		Chain: verify.ChainReport{
			Valid: true,
			Nodes: []verify.ChainNode{
				{Bundle: "b1", StatementID: "stmt-1", AttestationType: "prompt_attestation", GeneratedAt: "2025-01-01T00:00:00Z"},
			},
			Edges: []verify.ChainEdge{
				{FromStatementID: "stmt-2", FromType: "eval_attestation", ToStatementID: "stmt-1", ToType: "prompt_attestation", Satisfied: true},
			},
		},
	}
}

func TestBuildMarkdown_PassingReport(t *testing.T) {
	md := BuildMarkdown(sampleReport())

	if !strings.Contains(md, "# LLM Supply-Chain Verification Report") {
		t.Error("missing title")
	}
	if !strings.Contains(md, "Status: **PASS**") {
		t.Error("missing PASS status")
	}
	if !strings.Contains(md, "Exit Code: `0`") {
		t.Error("missing exit code")
	}
	if !strings.Contains(md, "Bundles Checked: `2`") {
		t.Error("missing bundle count")
	}
	if !strings.Contains(md, "## Checks") {
		t.Error("missing checks section")
	}
	if !strings.Contains(md, "b1.bundle.json") {
		t.Error("missing bundle name in checks table")
	}
	if !strings.Contains(md, "## Statements") {
		t.Error("missing statements section")
	}
	if !strings.Contains(md, "prompt_attestation") {
		t.Error("missing attestation type in statements")
	}
	if !strings.Contains(md, "## Provenance Chain") {
		t.Error("missing provenance chain section")
	}
	if !strings.Contains(md, "Valid: **true**") {
		t.Error("missing chain validity")
	}
}

func TestBuildMarkdown_FailingReport(t *testing.T) {
	r := sampleReport()
	r.Passed = false
	r.ExitCode = 12
	r.Violations = []string{"digest mismatch on prompt"}

	md := BuildMarkdown(r)

	if !strings.Contains(md, "Status: **FAIL**") {
		t.Error("missing FAIL status")
	}
	if !strings.Contains(md, "## Violations") {
		t.Error("missing violations section")
	}
	if !strings.Contains(md, "digest mismatch on prompt") {
		t.Error("missing violation text")
	}
}

func TestBuildMarkdown_ChainViolations(t *testing.T) {
	r := sampleReport()
	r.Chain.Valid = false
	r.Chain.Violations = []string{"missing chain predecessor: eval requires prompt"}

	md := BuildMarkdown(r)

	if !strings.Contains(md, "### Chain Violations") {
		t.Error("missing chain violations section")
	}
	if !strings.Contains(md, "missing chain predecessor") {
		t.Error("missing chain violation text")
	}
}

func TestBuildMarkdown_DependsOnRendering(t *testing.T) {
	md := BuildMarkdown(sampleReport())

	if !strings.Contains(md, "prompt_attestation, corpus_attestation") {
		t.Error("depends_on not joined correctly")
	}
}

func TestBuildMarkdown_NoDependsOnShowsDash(t *testing.T) {
	r := verify.Report{
		Passed:      true,
		ExitCode:    0,
		BundleCount: 1,
		Checks:      []verify.CheckResult{},
		Statements: []verify.StatementSummary{
			{AttestationType: "prompt_attestation", StatementID: "s1", PrivacyMode: "hash_only"},
		},
		Chain: verify.ChainReport{Valid: true},
	}
	md := BuildMarkdown(r)
	if !strings.Contains(md, "| - |") {
		t.Error("expected dash for empty depends_on")
	}
}

func TestBuildMarkdown_EdgeEmptyToID(t *testing.T) {
	r := verify.Report{
		Passed:      true,
		ExitCode:    0,
		BundleCount: 1,
		Chain: verify.ChainReport{
			Valid: true,
			Edges: []verify.ChainEdge{
				{FromStatementID: "s1", FromType: "eval_attestation", ToType: "prompt_attestation", Satisfied: true},
			},
		},
	}
	md := BuildMarkdown(r)
	if !strings.Contains(md, "| - |") {
		t.Error("expected dash for empty to_statement_id")
	}
}

func TestBuildMarkdown_PipeInMessage(t *testing.T) {
	r := verify.Report{
		Passed:      true,
		ExitCode:    0,
		BundleCount: 1,
		Checks: []verify.CheckResult{
			{Bundle: "b1", Check: "test", Passed: true, Message: "a|b"},
		},
		Chain: verify.ChainReport{Valid: true},
	}
	md := BuildMarkdown(r)
	if !strings.Contains(md, `a\|b`) {
		t.Error("pipe in message not escaped")
	}
}

func TestWriteMarkdown(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.md")

	if err := WriteMarkdown(path, sampleReport()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "# LLM Supply-Chain Verification Report") {
		t.Error("written file missing title")
	}
}

func TestWriteJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "report.json")

	if err := WriteJSON(path, sampleReport()); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var r verify.Report
	if err := json.Unmarshal(raw, &r); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !r.Passed {
		t.Error("expected passed=true")
	}
	if r.BundleCount != 2 {
		t.Errorf("bundle_count = %d, want 2", r.BundleCount)
	}
	if len(r.Checks) != 3 {
		t.Errorf("checks count = %d, want 3", len(r.Checks))
	}
}
