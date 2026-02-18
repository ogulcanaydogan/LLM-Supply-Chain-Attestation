package types

import (
	"encoding/json"
	"testing"
)

func TestPredicateURI_AllTypes(t *testing.T) {
	tests := []struct {
		attestationType string
		wantURI         string
	}{
		{AttestationPrompt, "https://llmsa.dev/attestation/prompt/v1"},
		{AttestationCorpus, "https://llmsa.dev/attestation/corpus/v1"},
		{AttestationEval, "https://llmsa.dev/attestation/eval/v1"},
		{AttestationRoute, "https://llmsa.dev/attestation/route/v1"},
		{AttestationSLO, "https://llmsa.dev/attestation/slo/v1"},
	}
	for _, tt := range tests {
		got := PredicateURI(tt.attestationType)
		if got != tt.wantURI {
			t.Errorf("PredicateURI(%q) = %q, want %q", tt.attestationType, got, tt.wantURI)
		}
	}
}

func TestPredicateURI_UnknownType(t *testing.T) {
	got := PredicateURI("unknown_attestation")
	if got != "" {
		t.Errorf("PredicateURI(unknown) = %q, want empty", got)
	}
}

func TestPredicateURI_EmptyType(t *testing.T) {
	got := PredicateURI("")
	if got != "" {
		t.Errorf("PredicateURI('') = %q, want empty", got)
	}
}

func TestAttestationConstants(t *testing.T) {
	if AttestationPrompt != "prompt_attestation" {
		t.Errorf("AttestationPrompt = %q", AttestationPrompt)
	}
	if AttestationCorpus != "corpus_attestation" {
		t.Errorf("AttestationCorpus = %q", AttestationCorpus)
	}
	if AttestationEval != "eval_attestation" {
		t.Errorf("AttestationEval = %q", AttestationEval)
	}
	if AttestationRoute != "route_attestation" {
		t.Errorf("AttestationRoute = %q", AttestationRoute)
	}
	if AttestationSLO != "slo_attestation" {
		t.Errorf("AttestationSLO = %q", AttestationSLO)
	}
}

func TestStatementJSON_RoundTrip(t *testing.T) {
	s := Statement{
		SchemaVersion:   "1.0.0",
		StatementID:     "stmt-abc",
		AttestationType: AttestationPrompt,
		PredicateType:   "https://llmsa.dev/attestation/prompt/v1",
		GeneratedAt:     "2025-01-01T00:00:00Z",
		Generator: Generator{
			Name:    "llmsa",
			Version: "0.1.0",
			GitSHA:  "abc123",
		},
		Subject: []Subject{
			{
				Name: "system_prompt.txt",
				URI:  "file://system_prompt.txt",
				Digest: Digest{
					SHA256: "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855",
				},
				SizeBytes: 42,
			},
		},
		Privacy: Privacy{Mode: "hash_only"},
		Annotations: map[string]string{
			"depends_on": "corpus_attestation",
		},
	}

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got Statement
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.StatementID != s.StatementID {
		t.Errorf("statement_id = %q", got.StatementID)
	}
	if got.AttestationType != s.AttestationType {
		t.Errorf("attestation_type = %q", got.AttestationType)
	}
	if got.Generator.Name != "llmsa" {
		t.Errorf("generator.name = %q", got.Generator.Name)
	}
	if len(got.Subject) != 1 {
		t.Fatalf("subject count = %d", len(got.Subject))
	}
	if got.Subject[0].Digest.SHA256 != s.Subject[0].Digest.SHA256 {
		t.Errorf("subject digest mismatch")
	}
	if got.Privacy.Mode != "hash_only" {
		t.Errorf("privacy.mode = %q", got.Privacy.Mode)
	}
	if got.Annotations["depends_on"] != "corpus_attestation" {
		t.Errorf("annotations = %v", got.Annotations)
	}
}

func TestStatementJSON_OmitEmpty(t *testing.T) {
	s := Statement{
		SchemaVersion:   "1.0.0",
		StatementID:     "stmt-1",
		AttestationType: AttestationEval,
		PredicateType:   "https://llmsa.dev/attestation/eval/v1",
		GeneratedAt:     "2025-01-01T00:00:00Z",
		Generator:       Generator{Name: "llmsa", Version: "0.1.0"},
		Subject:         []Subject{},
		Privacy:         Privacy{Mode: "hash_only"},
	}

	raw, err := json.Marshal(s)
	if err != nil {
		t.Fatal(err)
	}
	str := string(raw)

	if contains(str, `"materials"`) {
		t.Error("materials should be omitted when empty")
	}
	if contains(str, `"annotations"`) {
		t.Error("annotations should be omitted when empty/nil")
	}
}

func TestPrivacyJSON_EncryptedPayload(t *testing.T) {
	p := Privacy{
		Mode:                           "encrypted_payload",
		EncryptedBlobDigest:            "sha256:abc123",
		EncryptionRecipientFingerprint: "fp123",
	}

	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}

	var got Privacy
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.Mode != "encrypted_payload" {
		t.Errorf("mode = %q", got.Mode)
	}
	if got.EncryptedBlobDigest != "sha256:abc123" {
		t.Errorf("encrypted_blob_digest = %q", got.EncryptedBlobDigest)
	}
	if got.EncryptionRecipientFingerprint != "fp123" {
		t.Errorf("encryption_recipient_fingerprint = %q", got.EncryptionRecipientFingerprint)
	}
}

func TestPrivacyJSON_OmitEncryptedFields(t *testing.T) {
	p := Privacy{Mode: "hash_only"}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	str := string(raw)
	if contains(str, `"encrypted_blob_digest"`) {
		t.Error("encrypted_blob_digest should be omitted for hash_only mode")
	}
	if contains(str, `"encryption_recipient_fingerprint"`) {
		t.Error("encryption_recipient_fingerprint should be omitted for hash_only mode")
	}
}

func TestPromptPredicateJSON(t *testing.T) {
	p := PromptPredicate{
		PromptBundleDigest: "sha256:abc",
		SystemPromptDigest: "sha256:def",
		TemplateDigests:    []string{"sha256:t1", "sha256:t2"},
		ToolSchemaDigests:  []string{"sha256:s1"},
		SafetyPolicyDigest: "sha256:safety",
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got PromptPredicate
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.SystemPromptDigest != "sha256:def" {
		t.Errorf("system_prompt_digest = %q", got.SystemPromptDigest)
	}
	if len(got.TemplateDigests) != 2 {
		t.Errorf("template_digests len = %d", len(got.TemplateDigests))
	}
}

func TestCorpusPredicateJSON(t *testing.T) {
	p := CorpusPredicate{
		CorpusSnapshotID: "snap-1",
		ConnectorConfigDigests: []NamedDigest{
			{Name: "local", Digest: "sha256:conn"},
		},
		DocumentManifestDigest: "sha256:dm",
		ChunkingConfigDigest:   "sha256:chunk",
		EmbeddingModel:         "text-embedding-3-small",
		EmbeddingInputDigest:   "sha256:emb",
		VectorIndexDigest:      "sha256:vec",
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got CorpusPredicate
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.CorpusSnapshotID != "snap-1" {
		t.Errorf("corpus_snapshot_id = %q", got.CorpusSnapshotID)
	}
	if len(got.ConnectorConfigDigests) != 1 {
		t.Errorf("connector count = %d", len(got.ConnectorConfigDigests))
	}
	if got.EmbeddingModel != "text-embedding-3-small" {
		t.Errorf("embedding_model = %q", got.EmbeddingModel)
	}
}

func TestEvalPredicateJSON(t *testing.T) {
	p := EvalPredicate{
		EvalSuiteID:           "suite-1",
		TestsetDigest:         "sha256:ts",
		ScoringConfigDigest:   "sha256:sc",
		BaselineResultDigest:  "sha256:br",
		CandidateResultDigest: "sha256:cr",
		Metrics:               map[string]float64{"accuracy": 0.95},
		Thresholds:            map[string]float64{"accuracy": 0.90},
		RegressionDetected:    false,
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got EvalPredicate
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.EvalSuiteID != "suite-1" {
		t.Errorf("eval_suite_id = %q", got.EvalSuiteID)
	}
	if got.Metrics["accuracy"] != 0.95 {
		t.Errorf("metrics = %v", got.Metrics)
	}
	if got.RegressionDetected {
		t.Error("regression_detected should be false")
	}
}

func TestRoutePredicateJSON(t *testing.T) {
	p := RoutePredicate{
		RouteConfigDigest: "sha256:rc",
		ProviderSet: []ProviderModel{
			{Provider: "openai", Model: "gpt-4"},
			{Provider: "anthropic", Model: "claude-3"},
		},
		BudgetPolicyDigest:  "sha256:bp",
		FallbackGraphDigest: "sha256:fg",
		RoutingStrategy:     "cost-optimized",
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got RoutePredicate
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if len(got.ProviderSet) != 2 {
		t.Errorf("provider_set len = %d", len(got.ProviderSet))
	}
	if got.RoutingStrategy != "cost-optimized" {
		t.Errorf("routing_strategy = %q", got.RoutingStrategy)
	}
}

func TestSLOPredicateJSON(t *testing.T) {
	p := SLOPredicate{
		SLOProfileID:          "slo-1",
		Window:                TimeWindow{Start: "2025-01-01", End: "2025-01-31"},
		TTFTMSP50:             150.0,
		TTFTMSP95:             350.0,
		TokensPerSecP50:       45.0,
		CostPer1KTokensCapUSD: 0.005,
		ErrorRateCap:          0.01,
		ErrorBudgetRemaining:  0.98,
	}
	raw, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	var got SLOPredicate
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatal(err)
	}
	if got.SLOProfileID != "slo-1" {
		t.Errorf("slo_profile_id = %q", got.SLOProfileID)
	}
	if got.Window.Start != "2025-01-01" || got.Window.End != "2025-01-31" {
		t.Errorf("window = %v", got.Window)
	}
	if got.TTFTMSP50 != 150.0 {
		t.Errorf("ttft_ms_p50 = %f", got.TTFTMSP50)
	}
	if got.ErrorBudgetRemaining != 0.98 {
		t.Errorf("error_budget_remaining = %f", got.ErrorBudgetRemaining)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && searchString(s, substr)
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
