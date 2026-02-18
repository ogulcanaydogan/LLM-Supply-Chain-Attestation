package types

type EvalPredicate struct {
	EvalSuiteID           string             `json:"eval_suite_id"`
	TestsetDigest         string             `json:"testset_digest"`
	ScoringConfigDigest   string             `json:"scoring_config_digest"`
	BaselineResultDigest  string             `json:"baseline_result_digest"`
	CandidateResultDigest string             `json:"candidate_result_digest"`
	Metrics               map[string]float64 `json:"metrics"`
	Thresholds            map[string]float64 `json:"thresholds"`
	RegressionDetected    bool               `json:"regression_detected"`
	RunEnvironmentDigest  string             `json:"run_environment_digest,omitempty"`
}
