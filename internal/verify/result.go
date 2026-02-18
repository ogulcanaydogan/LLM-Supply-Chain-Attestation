package verify

const (
	ExitPass           = 0
	ExitMissing        = 10
	ExitSignatureFail  = 11
	ExitDigestMismatch = 12
	ExitPolicyFail     = 13
	ExitSchemaFail     = 14
)

type CheckResult struct {
	Bundle  string `json:"bundle"`
	Check   string `json:"check"`
	Passed  bool   `json:"passed"`
	Message string `json:"message"`
}

type StatementSummary struct {
	AttestationType string `json:"attestation_type"`
	StatementID     string `json:"statement_id"`
	PrivacyMode     string `json:"privacy_mode"`
}

type Report struct {
	Passed      bool               `json:"passed"`
	ExitCode    int                `json:"exit_code"`
	BundleCount int                `json:"bundle_count"`
	Checks      []CheckResult      `json:"checks"`
	Violations  []string           `json:"violations"`
	Statements  []StatementSummary `json:"statements"`
}
