package attest

import (
	"fmt"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type EvalConfig struct {
	EvalSuiteID      string             `yaml:"eval_suite_id"`
	Testset          string             `yaml:"testset"`
	ScoringConfig    string             `yaml:"scoring_config"`
	BaselineResults  string             `yaml:"baseline_results"`
	CandidateResults string             `yaml:"candidate_results"`
	Metrics          map[string]float64 `yaml:"metrics"`
	Thresholds       map[string]float64 `yaml:"thresholds"`
	RunEnvironment   string             `yaml:"run_environment"`
}

func CollectEval(configPath string) (types.Statement, error) {
	cfg := EvalConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return types.Statement{}, err
	}
	cfg.Testset = resolvePath(configPath, cfg.Testset)
	cfg.ScoringConfig = resolvePath(configPath, cfg.ScoringConfig)
	cfg.BaselineResults = resolvePath(configPath, cfg.BaselineResults)
	cfg.CandidateResults = resolvePath(configPath, cfg.CandidateResults)
	cfg.RunEnvironment = resolvePath(configPath, cfg.RunEnvironment)
	if cfg.EvalSuiteID == "" {
		return types.Statement{}, fmt.Errorf("eval_suite_id is required")
	}
	for _, req := range []struct {
		path string
		name string
	}{{cfg.Testset, "testset"}, {cfg.ScoringConfig, "scoring_config"}, {cfg.BaselineResults, "baseline_results"}, {cfg.CandidateResults, "candidate_results"}} {
		if err := requirePath(req.path, req.name); err != nil {
			return types.Statement{}, err
		}
	}

	testsetDigest, _, _ := hash.DigestFile(cfg.Testset)
	scoreDigest, _, _ := hash.DigestFile(cfg.ScoringConfig)
	baselineDigest, _, _ := hash.DigestFile(cfg.BaselineResults)
	candidateDigest, _, _ := hash.DigestFile(cfg.CandidateResults)

	regression := false
	for thresholdKey, thresholdValue := range cfg.Thresholds {
		if strings.HasSuffix(thresholdKey, "_min") {
			metric := strings.TrimSuffix(thresholdKey, "_min")
			if cfg.Metrics[metric] < thresholdValue {
				regression = true
			}
		}
		if strings.HasSuffix(thresholdKey, "_max") {
			metric := strings.TrimSuffix(thresholdKey, "_max")
			if cfg.Metrics[metric] > thresholdValue {
				regression = true
			}
		}
	}

	predicate := types.EvalPredicate{
		EvalSuiteID:           cfg.EvalSuiteID,
		TestsetDigest:         testsetDigest,
		ScoringConfigDigest:   scoreDigest,
		BaselineResultDigest:  baselineDigest,
		CandidateResultDigest: candidateDigest,
		Metrics:               cfg.Metrics,
		Thresholds:            cfg.Thresholds,
		RegressionDetected:    regression,
	}
	if cfg.RunEnvironment != "" {
		d, _, err := hash.DigestFile(cfg.RunEnvironment)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.RunEnvironmentDigest = d
	}

	subjects := make([]types.Subject, 0, 4)
	for _, p := range []string{cfg.Testset, cfg.ScoringConfig, cfg.BaselineResults, cfg.CandidateResults} {
		s, err := subjectFromPath(p)
		if err != nil {
			return types.Statement{}, err
		}
		subjects = append(subjects, s)
	}
	return newStatement(types.AttestationEval, predicate, subjects, nil), nil
}
