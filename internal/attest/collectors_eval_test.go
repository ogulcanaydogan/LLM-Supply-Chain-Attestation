package attest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

func TestCollectEval(t *testing.T) {
	st, err := CollectEval("../../examples/tiny-rag/configs/eval.yaml")
	if err != nil {
		t.Fatal(err)
	}
	if st.AttestationType != types.AttestationEval {
		t.Fatalf("type = %q, want eval_attestation", st.AttestationType)
	}
	if len(st.Subject) == 0 {
		t.Fatal("expected subjects")
	}
	pred, ok := st.Predicate.(types.EvalPredicate)
	if !ok {
		t.Fatal("predicate is not EvalPredicate")
	}
	if pred.EvalSuiteID == "" {
		t.Error("eval_suite_id is empty")
	}
	if pred.TestsetDigest == "" {
		t.Error("testset_digest is empty")
	}
	if pred.ScoringConfigDigest == "" {
		t.Error("scoring_config_digest is empty")
	}
	if pred.BaselineResultDigest == "" {
		t.Error("baseline_result_digest is empty")
	}
	if pred.CandidateResultDigest == "" {
		t.Error("candidate_result_digest is empty")
	}
}

func TestCollectEval_DependsOnPromptAndCorpus(t *testing.T) {
	st, err := CollectEval("../../examples/tiny-rag/configs/eval.yaml")
	if err != nil {
		t.Fatal(err)
	}
	deps := st.Annotations["depends_on"]
	if !strings.Contains(deps, types.AttestationPrompt) {
		t.Errorf("depends_on should contain prompt_attestation, got %q", deps)
	}
	if !strings.Contains(deps, types.AttestationCorpus) {
		t.Errorf("depends_on should contain corpus_attestation, got %q", deps)
	}
}

func TestCollectEval_RegressionDetected(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"testset.json", "scoring.yaml", "baseline.json", "candidate.json"} {
		os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0o644)
	}
	cfg := filepath.Join(dir, "eval.yaml")
	content := `eval_suite_id: regression-test
testset: testset.json
scoring_config: scoring.yaml
baseline_results: baseline.json
candidate_results: candidate.json
metrics:
  accuracy: 0.80
thresholds:
  accuracy_min: 0.90
`
	os.WriteFile(cfg, []byte(content), 0o644)

	st, err := CollectEval(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pred := st.Predicate.(types.EvalPredicate)
	if !pred.RegressionDetected {
		t.Error("expected regression_detected=true when metric below min threshold")
	}
}

func TestCollectEval_NoRegression(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"testset.json", "scoring.yaml", "baseline.json", "candidate.json"} {
		os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0o644)
	}
	cfg := filepath.Join(dir, "eval.yaml")
	content := `eval_suite_id: pass-test
testset: testset.json
scoring_config: scoring.yaml
baseline_results: baseline.json
candidate_results: candidate.json
metrics:
  accuracy: 0.95
thresholds:
  accuracy_min: 0.90
`
	os.WriteFile(cfg, []byte(content), 0o644)

	st, err := CollectEval(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pred := st.Predicate.(types.EvalPredicate)
	if pred.RegressionDetected {
		t.Error("expected regression_detected=false when metric above min threshold")
	}
}

func TestCollectEval_MissingSuiteID(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"testset.json", "scoring.yaml", "baseline.json", "candidate.json"} {
		os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0o644)
	}
	cfg := filepath.Join(dir, "eval.yaml")
	content := `testset: testset.json
scoring_config: scoring.yaml
baseline_results: baseline.json
candidate_results: candidate.json
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectEval(cfg)
	if err == nil {
		t.Fatal("expected error for missing eval_suite_id")
	}
	if !strings.Contains(err.Error(), "eval_suite_id") {
		t.Errorf("error = %q", err)
	}
}

func TestCollectEval_MissingTestset(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"scoring.yaml", "baseline.json", "candidate.json"} {
		os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0o644)
	}
	cfg := filepath.Join(dir, "eval.yaml")
	content := `eval_suite_id: suite-1
testset: missing.json
scoring_config: scoring.yaml
baseline_results: baseline.json
candidate_results: candidate.json
`
	os.WriteFile(cfg, []byte(content), 0o644)

	_, err := CollectEval(cfg)
	if err == nil {
		t.Fatal("expected error for missing testset file")
	}
}

func TestCollectEval_MaxThresholdRegression(t *testing.T) {
	dir := t.TempDir()
	for _, f := range []string{"testset.json", "scoring.yaml", "baseline.json", "candidate.json"} {
		os.WriteFile(filepath.Join(dir, f), []byte(`{}`), 0o644)
	}
	cfg := filepath.Join(dir, "eval.yaml")
	content := `eval_suite_id: max-test
testset: testset.json
scoring_config: scoring.yaml
baseline_results: baseline.json
candidate_results: candidate.json
metrics:
  latency: 500
thresholds:
  latency_max: 200
`
	os.WriteFile(cfg, []byte(content), 0o644)

	st, err := CollectEval(cfg)
	if err != nil {
		t.Fatal(err)
	}
	pred := st.Predicate.(types.EvalPredicate)
	if !pred.RegressionDetected {
		t.Error("expected regression_detected=true when metric above max threshold")
	}
}
