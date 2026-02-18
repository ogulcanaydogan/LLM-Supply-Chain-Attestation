package verify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

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
