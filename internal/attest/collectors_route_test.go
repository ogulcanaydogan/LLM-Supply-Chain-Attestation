package attest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestCollectRoute(t *testing.T) {
	st, err := CollectRoute("../../examples/tiny-rag/configs/route.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.AttestationType != types.AttestationRoute {
		t.Fatalf("type = %q, want route_attestation", st.AttestationType)
	}
	if len(st.Subject) == 0 {
		t.Fatal("expected subjects")
	}
	pred, ok := st.Predicate.(types.RoutePredicate)
	if !ok {
		t.Fatal("predicate is not RoutePredicate")
	}
	if pred.RouteConfigDigest == "" {
		t.Error("route_config_digest is empty")
	}
	if pred.BudgetPolicyDigest == "" {
		t.Error("budget_policy_digest is empty")
	}
	if pred.FallbackGraphDigest == "" {
		t.Error("fallback_graph_digest is empty")
	}
	if pred.RoutingStrategy == "" {
		t.Error("routing_strategy is empty")
	}
	if len(pred.ProviderSet) == 0 {
		t.Error("expected provider_set entries")
	}
}

func TestCollectRoute_DependsOnEval(t *testing.T) {
	st, err := CollectRoute("../../examples/tiny-rag/configs/route.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := st.Annotations["depends_on"]
	if !strings.Contains(deps, types.AttestationEval) {
		t.Errorf("depends_on should contain eval_attestation, got %q", deps)
	}
}

func TestCollectRoute_OptionalCanary(t *testing.T) {
	st, err := CollectRoute("../../examples/tiny-rag/configs/route.yaml")
	if err != nil {
		t.Fatal(err)
	}
	pred := st.Predicate.(types.RoutePredicate)
	if pred.CanaryConfigDigest == "" {
		t.Error("expected canary_config_digest to be populated from tiny-rag example")
	}
}

func TestCollectRoute_MissingStrategy(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"route.yaml", "budget.yaml", "fallback.yaml"} {
		os.WriteFile(filepath.Join(dir, f), []byte("---"), 0o644)
	}
	cfg := filepath.Join(dir, "config.yaml")
	content := `route_config: route.yaml
provider_set:
  - provider: openai
    model: gpt-4
budget_policy: budget.yaml
fallback_graph: fallback.yaml
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectRoute(cfg)
	if err == nil {
		t.Fatal("expected error for missing routing_strategy")
	}
	if !strings.Contains(err.Error(), "routing_strategy") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectRoute_EmptyProviderSet(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"route.yaml", "budget.yaml", "fallback.yaml"} {
		os.WriteFile(filepath.Join(dir, f), []byte("---"), 0o644)
	}
	cfg := filepath.Join(dir, "config.yaml")
	content := `route_config: route.yaml
routing_strategy: latency_aware
provider_set: []
budget_policy: budget.yaml
fallback_graph: fallback.yaml
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectRoute(cfg)
	if err == nil {
		t.Fatal("expected error for empty provider_set")
	}
	if !strings.Contains(err.Error(), "provider_set") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectRoute_MissingRouteConfig(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"budget.yaml", "fallback.yaml"} {
		os.WriteFile(filepath.Join(dir, f), []byte("---"), 0o644)
	}
	cfg := filepath.Join(dir, "config.yaml")
	content := `route_config: nonexistent.yaml
routing_strategy: latency_aware
provider_set:
  - provider: openai
    model: gpt-4
budget_policy: budget.yaml
fallback_graph: fallback.yaml
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectRoute(cfg)
	if err == nil {
		t.Fatal("expected error for missing route_config file")
	}
}
