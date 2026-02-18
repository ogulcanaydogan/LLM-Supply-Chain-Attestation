package attest

import (
	"fmt"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/hash"
	"github.com/ogulcanaydogan/llm-supply-chain-attestation/pkg/types"
)

type SLOConfig struct {
	SLOProfileID          string  `yaml:"slo_profile_id"`
	WindowStart           string  `yaml:"window_start"`
	WindowEnd             string  `yaml:"window_end"`
	TTFTMSP50             float64 `yaml:"ttft_ms_p50"`
	TTFTMSP95             float64 `yaml:"ttft_ms_p95"`
	TokensPerSecP50       float64 `yaml:"tokens_per_sec_p50"`
	CostPer1KTokensCapUSD float64 `yaml:"cost_per_1k_tokens_cap_usd"`
	ErrorRateCap          float64 `yaml:"error_rate_cap"`
	ErrorBudgetRemaining  float64 `yaml:"error_budget_remaining"`
	ObservabilityQuery    string  `yaml:"observability_query"`
}

func CollectSLO(configPath string) (types.Statement, error) {
	cfg := SLOConfig{}
	if err := LoadConfig(configPath, &cfg); err != nil {
		return types.Statement{}, err
	}
	cfg.ObservabilityQuery = resolvePath(configPath, cfg.ObservabilityQuery)
	if cfg.SLOProfileID == "" || cfg.WindowStart == "" || cfg.WindowEnd == "" {
		return types.Statement{}, fmt.Errorf("slo_profile_id, window_start and window_end are required")
	}

	predicate := types.SLOPredicate{
		SLOProfileID:          cfg.SLOProfileID,
		Window:                types.TimeWindow{Start: cfg.WindowStart, End: cfg.WindowEnd},
		TTFTMSP50:             cfg.TTFTMSP50,
		TTFTMSP95:             cfg.TTFTMSP95,
		TokensPerSecP50:       cfg.TokensPerSecP50,
		CostPer1KTokensCapUSD: cfg.CostPer1KTokensCapUSD,
		ErrorRateCap:          cfg.ErrorRateCap,
		ErrorBudgetRemaining:  cfg.ErrorBudgetRemaining,
	}
	subjects := []types.Subject{}
	if cfg.ObservabilityQuery != "" {
		if err := requirePath(cfg.ObservabilityQuery, "observability_query"); err != nil {
			return types.Statement{}, err
		}
		d, _, err := hash.DigestFile(cfg.ObservabilityQuery)
		if err != nil {
			return types.Statement{}, err
		}
		predicate.ObservabilityQueryDigest = d
		s, err := subjectFromPath(cfg.ObservabilityQuery)
		if err != nil {
			return types.Statement{}, err
		}
			subjects = append(subjects, s)
		}
	statement := newStatement(types.AttestationSLO, predicate, subjects, nil)
	setDependsOn(&statement, types.AttestationRoute)
	return statement, nil
}
