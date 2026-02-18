package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/sign"
)

func TestVerifySignaturePath(t *testing.T) {
	tmp := t.TempDir()
	keyPath := filepath.Join(tmp, "dev.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	statement := map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "id-1",
		"attestation_type": "slo_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/slo/v1",
		"generated_at":     "2026-02-17T20:10:11Z",
		"generator":        map[string]any{"name": "llmsa", "version": "0.1.0", "git_sha": "local"},
		"subject":          []any{map[string]any{"name": "x", "uri": "../../examples/tiny-rag/slo/profile.json", "digest": map[string]any{"sha256": "d4f0cca6f47f5d13df8ba6ebf8fc4b6ba6ec7ca6b6f6aeac21358f6d79e55f6f"}, "size_bytes": 0}},
		"predicate": map[string]any{
			"slo_profile_id":             "prod-rag-api",
			"window":                     map[string]any{"start": "2026-02-17T00:00:00Z", "end": "2026-02-17T23:59:59Z"},
			"ttft_ms_p50":                700,
			"ttft_ms_p95":                1200,
			"tokens_per_sec_p50":         30,
			"cost_per_1k_tokens_cap_usd": 0.15,
			"error_rate_cap":             0.02,
			"error_budget_remaining":     0.7,
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}

	canonical, err := json.Marshal(statement)
	if err != nil {
		t.Fatal(err)
	}
	mat, err := signer.Sign(canonical)
	if err != nil {
		t.Fatal(err)
	}
	bundle, err := sign.CreateBundle(statement, mat)
	if err != nil {
		t.Fatal(err)
	}
	bundlePath := filepath.Join(tmp, "s.bundle.json")
	if err := sign.WriteBundle(bundlePath, bundle); err != nil {
		t.Fatal(err)
	}

	report := Run(Options{SourcePath: bundlePath, SchemaDir: "../../schemas/v1"})
	if report.ExitCode == ExitSignatureFail {
		t.Fatalf("unexpected signature fail: %+v", report)
	}
	if len(report.Checks) == 0 {
		t.Fatalf("expected checks")
	}
	_ = os.Remove(bundlePath)
}

func TestRunEmptyDirectory(t *testing.T) {
	report := Run(Options{
		SourcePath: t.TempDir(),
		SchemaDir:  "../../schemas/v1",
	})
	if report.ExitCode != ExitMissing {
		t.Fatalf("expected exit code %d, got %d", ExitMissing, report.ExitCode)
	}
}

func TestRunChainGraphFailure(t *testing.T) {
	tmp := t.TempDir()
	keyPath := filepath.Join(tmp, "dev.pem")
	if err := sign.GeneratePEMPrivateKey(keyPath); err != nil {
		t.Fatal(err)
	}
	signer, err := sign.NewPEMSigner(keyPath)
	if err != nil {
		t.Fatal(err)
	}

	if _, err := writeBundleForStatement(tmp, signer, map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "prompt-1",
		"attestation_type": "prompt_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/prompt/v1",
		"generated_at":     "2026-02-18T00:00:00Z",
		"generator":        map[string]any{"name": "llmsa", "version": "1.0.0", "git_sha": "abc"},
		"subject":          []any{},
		"predicate": map[string]any{
			"prompt_bundle_digest": "sha256:bundle",
			"system_prompt_digest": "sha256:system",
			"template_digests":     []any{"sha256:template"},
			"tool_schema_digests":  []any{"sha256:tool"},
			"safety_policy_digest": "sha256:safety",
		},
		"privacy": map[string]any{"mode": "hash_only"},
	}); err != nil {
		t.Fatal(err)
	}
	if _, err := writeBundleForStatement(tmp, signer, map[string]any{
		"schema_version":   "1.0.0",
		"statement_id":     "slo-1",
		"attestation_type": "slo_attestation",
		"predicate_type":   "https://llmsa.dev/attestation/slo/v1",
		"generated_at":     "2026-02-18T00:00:01Z",
		"generator":        map[string]any{"name": "llmsa", "version": "1.0.0", "git_sha": "abc"},
		"subject":          []any{},
		"predicate": map[string]any{
			"slo_profile_id":             "prod",
			"window":                     map[string]any{"start": "2026-02-18T00:00:00Z", "end": "2026-02-18T01:00:00Z"},
			"ttft_ms_p50":                1.0,
			"ttft_ms_p95":                2.0,
			"tokens_per_sec_p50":         3.0,
			"cost_per_1k_tokens_cap_usd": 0.1,
			"error_rate_cap":             0.01,
			"error_budget_remaining":     0.8,
		},
		"privacy": map[string]any{"mode": "hash_only"},
		"annotations": map[string]any{
			"depends_on": "route_attestation",
		},
	}); err != nil {
		t.Fatal(err)
	}

	report := Run(Options{
		SourcePath: tmp,
		SchemaDir:  "../../schemas/v1",
	})
	if report.ExitCode != ExitSchemaFail {
		t.Fatalf("expected schema/chain failure exit code %d, got %d", ExitSchemaFail, report.ExitCode)
	}
	if report.Chain.Valid {
		t.Fatalf("expected invalid chain report")
	}
	if len(report.Chain.Violations) == 0 {
		t.Fatalf("expected chain violations")
	}
}

func TestBundlePathsSingleFileAndHelpers(t *testing.T) {
	tmp := t.TempDir()
	file := filepath.Join(tmp, "x.bundle.json")
	if err := os.WriteFile(file, []byte("{}"), 0o644); err != nil {
		t.Fatal(err)
	}
	paths, err := bundlePaths(file)
	if err != nil {
		t.Fatal(err)
	}
	if len(paths) != 1 || paths[0] != file {
		t.Fatalf("unexpected bundle paths: %v", paths)
	}

	statement := map[string]any{
		"privacy": map[string]any{"mode": "hash_only"},
		"annotations": map[string]any{
			"depends_on": "route_attestation, eval_attestation",
		},
	}
	deps := dependsOn(statement)
	sort.Strings(deps)
	if strings.Join(deps, ",") != "eval_attestation,route_attestation" {
		t.Fatalf("unexpected depends_on parsing: %v", deps)
	}
	if privacyMode(statement) != "hash_only" {
		t.Fatalf("unexpected privacy mode")
	}
}

func TestWriteJSON(t *testing.T) {
	out := filepath.Join(t.TempDir(), "verify.json")
	report := Report{
		Passed:   true,
		ExitCode: ExitPass,
		Checks:   []CheckResult{{Bundle: "b", Check: "schema", Passed: true, Message: "ok"}},
	}
	if err := WriteJSON(out, report); err != nil {
		t.Fatal(err)
	}
	raw, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(raw), "\"exit_code\": 0") {
		t.Fatalf("unexpected json output: %s", string(raw))
	}
}

func writeBundleForStatement(dir string, signer *sign.PEMSigner, statement map[string]any) (string, error) {
	canonical, err := hash.CanonicalJSON(statement)
	if err != nil {
		return "", err
	}
	mat, err := signer.Sign(canonical)
	if err != nil {
		return "", err
	}
	bundle, err := sign.CreateBundle(statement, mat)
	if err != nil {
		return "", err
	}
	path := filepath.Join(dir, "test-"+statement["statement_id"].(string)+"-"+time.Now().Format("150405.000000")+".bundle.json")
	if err := sign.WriteBundle(path, bundle); err != nil {
		return "", err
	}
	return path, nil
}
