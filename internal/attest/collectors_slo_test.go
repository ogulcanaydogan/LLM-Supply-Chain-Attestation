package attest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestCollectSLO(t *testing.T) {
	st, err := CollectSLO("../../examples/tiny-rag/configs/slo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.AttestationType != types.AttestationSLO {
		t.Fatalf("type = %q, want slo_attestation", st.AttestationType)
	}
	pred, ok := st.Predicate.(types.SLOPredicate)
	if !ok {
		t.Fatal("predicate is not SLOPredicate")
	}
	if pred.SLOProfileID == "" {
		t.Error("slo_profile_id is empty")
	}
	if pred.Window.Start == "" || pred.Window.End == "" {
		t.Error("window start/end is empty")
	}
}

func TestCollectSLO_DependsOnRoute(t *testing.T) {
	st, err := CollectSLO("../../examples/tiny-rag/configs/slo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := st.Annotations["depends_on"]
	if !strings.Contains(deps, types.AttestationRoute) {
		t.Errorf("depends_on should contain route_attestation, got %q", deps)
	}
}

func TestCollectSLO_NumericValues(t *testing.T) {
	st, err := CollectSLO("../../examples/tiny-rag/configs/slo.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pred := st.Predicate.(types.SLOPredicate)
	if pred.TTFTMSP50 <= 0 {
		t.Error("ttft_ms_p50 should be positive")
	}
	if pred.TTFTMSP95 <= 0 {
		t.Error("ttft_ms_p95 should be positive")
	}
	if pred.TokensPerSecP50 <= 0 {
		t.Error("tokens_per_sec_p50 should be positive")
	}
	if pred.CostPer1KTokensCapUSD <= 0 {
		t.Error("cost_per_1k_tokens_cap_usd should be positive")
	}
	if pred.ErrorRateCap <= 0 {
		t.Error("error_rate_cap should be positive")
	}
}

func TestCollectSLO_MissingProfileID(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "slo.yaml")
	content := `window_start: "2025-01-01T00:00:00Z"
window_end: "2025-01-31T23:59:59Z"
ttft_ms_p50: 100
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectSLO(cfg)
	if err == nil {
		t.Fatal("expected error for missing slo_profile_id")
	}
	if !strings.Contains(err.Error(), "slo_profile_id") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectSLO_MissingWindowStart(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "slo.yaml")
	content := `slo_profile_id: test-slo
window_end: "2025-01-31T23:59:59Z"
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectSLO(cfg)
	if err == nil {
		t.Fatal("expected error for missing window_start")
	}
}

func TestCollectSLO_OptionalObservability(t *testing.T) {
	dir := t.TempDir()
	cfg := filepath.Join(dir, "slo.yaml")
	content := `slo_profile_id: test-slo
window_start: "2025-01-01T00:00:00Z"
window_end: "2025-01-31T23:59:59Z"
ttft_ms_p50: 100
ttft_ms_p95: 300
`
	os.WriteFile(cfg, []byte(content), 0o644)

	st, err := CollectSLO(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if len(st.Subject) != 0 {
		t.Errorf("expected 0 subjects when no observability_query, got %d", len(st.Subject))
	}
}
