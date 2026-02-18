package verify

import (
	"strings"
	"testing"
)

func TestVerifyProvenanceChainValid(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "prompt-1",
			AttestationType: "prompt_attestation",
			GeneratedAt:     "2026-02-17T20:10:11Z",
		},
		{
			StatementID:     "corpus-1",
			AttestationType: "corpus_attestation",
			GeneratedAt:     "2026-02-17T20:10:12Z",
		},
		{
			StatementID:     "eval-1",
			AttestationType: "eval_attestation",
			GeneratedAt:     "2026-02-17T20:10:13Z",
			DependsOn:       []string{"prompt_attestation", "corpus_attestation"},
		},
		{
			StatementID:     "route-1",
			AttestationType: "route_attestation",
			GeneratedAt:     "2026-02-17T20:10:14Z",
			DependsOn:       []string{"eval_attestation"},
		},
		{
			StatementID:     "slo-1",
			AttestationType: "slo_attestation",
			GeneratedAt:     "2026-02-17T20:10:15Z",
			DependsOn:       []string{"route_attestation"},
		},
	})

	if !report.Valid {
		t.Fatalf("expected valid chain, got violations: %v", report.Violations)
	}
	if len(report.Edges) == 0 {
		t.Fatalf("expected chain edges")
	}
}

func TestVerifyProvenanceChainMissingPredecessor(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "route-1",
			AttestationType: "route_attestation",
			GeneratedAt:     "2026-02-17T20:10:14Z",
			DependsOn:       []string{"eval_attestation"},
		},
	})

	if report.Valid {
		t.Fatalf("expected invalid chain")
	}
	if !containsViolation(report.Violations, "missing chain predecessor") {
		t.Fatalf("expected missing predecessor violation, got %v", report.Violations)
	}
}

func TestVerifyProvenanceChainUnknownDependency(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "prompt-1",
			AttestationType: "prompt_attestation",
			GeneratedAt:     "2026-02-17T20:10:11Z",
			DependsOn:       []string{"unknown-ref"},
		},
	})

	if report.Valid {
		t.Fatalf("expected invalid chain")
	}
	if !containsViolation(report.Violations, "unknown dependency reference") {
		t.Fatalf("expected unknown dependency violation, got %v", report.Violations)
	}
}

func TestVerifyProvenanceChainTemporalOrder(t *testing.T) {
	report := VerifyProvenanceChain([]ChainStatement{
		{
			StatementID:     "eval-1",
			AttestationType: "eval_attestation",
			GeneratedAt:     "2026-02-17T20:10:15Z",
			DependsOn:       []string{"prompt-1", "corpus-1"},
		},
		{
			StatementID:     "prompt-1",
			AttestationType: "prompt_attestation",
			GeneratedAt:     "2026-02-17T20:10:16Z",
		},
		{
			StatementID:     "corpus-1",
			AttestationType: "corpus_attestation",
			GeneratedAt:     "2026-02-17T20:10:14Z",
		},
	})

	if report.Valid {
		t.Fatalf("expected invalid chain due to ordering")
	}
	if !containsViolation(report.Violations, "invalid chain order") {
		t.Fatalf("expected ordering violation, got %v", report.Violations)
	}
}

func containsViolation(violations []string, needle string) bool {
	for _, v := range violations {
		if strings.Contains(v, needle) {
			return true
		}
	}
	return false
}
