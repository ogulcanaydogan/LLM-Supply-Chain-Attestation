package types

type TimeWindow struct {
	Start string `json:"start"`
	End   string `json:"end"`
}

type SLOPredicate struct {
	SLOProfileID             string     `json:"slo_profile_id"`
	Window                   TimeWindow `json:"window"`
	TTFTMSP50                float64    `json:"ttft_ms_p50"`
	TTFTMSP95                float64    `json:"ttft_ms_p95"`
	TokensPerSecP50          float64    `json:"tokens_per_sec_p50"`
	CostPer1KTokensCapUSD    float64    `json:"cost_per_1k_tokens_cap_usd"`
	ErrorRateCap             float64    `json:"error_rate_cap"`
	ErrorBudgetRemaining     float64    `json:"error_budget_remaining"`
	ObservabilityQueryDigest string     `json:"observability_query_digest,omitempty"`
}
