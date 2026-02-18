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
	DependsOn       []string `json:"depends_on,omitempty"`
	GeneratedAt     string   `json:"generated_at,omitempty"`
}

type ChainNode struct {
	Bundle          string   `json:"bundle"`
	StatementID     string   `json:"statement_id"`
	AttestationType string   `json:"attestation_type"`
	GeneratedAt     string   `json:"generated_at"`
	DependsOn       []string `json:"depends_on,omitempty"`
}

type ChainEdge struct {
	FromStatementID string `json:"from_statement_id"`
	FromType        string `json:"from_type"`
	ToStatementID   string `json:"to_statement_id,omitempty"`
	ToType          string `json:"to_type"`
	Satisfied       bool   `json:"satisfied"`
	Detail          string `json:"detail,omitempty"`
}

type ChainReport struct {
	Valid      bool      `json:"valid"`
	Nodes      []ChainNode `json:"nodes,omitempty"`
	Edges      []ChainEdge `json:"edges,omitempty"`
	Violations []string  `json:"violations,omitempty"`
}

type Report struct {
	Passed      bool               `json:"passed"`
	ExitCode    int                `json:"exit_code"`
	BundleCount int                `json:"bundle_count"`
	Checks      []CheckResult      `json:"checks"`
	Violations  []string           `json:"violations"`
	Statements  []StatementSummary `json:"statements"`
	Chain       ChainReport        `json:"chain"`
}
