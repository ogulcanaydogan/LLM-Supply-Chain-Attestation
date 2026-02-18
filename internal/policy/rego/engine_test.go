package rego

import (
	"path/filepath"
	"runtime"
	"sort"
	"testing"

	policyyaml "github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/policy/yaml"
)

func TestRegoParityWithYAML(t *testing.T) {
	policy := policyyaml.Policy{
		Version:       "1",
		OIDCIssuer:    "https://token.actions.githubusercontent.com",
		IdentityRegex: "^https://github\\.com/.+$",
		Gates: []policyyaml.Gate{
			{
				ID:                   "G001",
				TriggerPaths:         []string{"examples/tiny-rag/app/**"},
				RequiredAttestations: []string{"prompt_attestation", "eval_attestation"},
				Message:              "Prompt changed without passing eval attestation.",
			},
			{
				ID:                   "G003",
				TriggerPaths:         []string{"examples/tiny-rag/route/**"},
				RequiredAttestations: []string{"route_attestation", "slo_attestation"},
				Message:              "Route changed without valid SLO attestation.",
			},
		},
	}

	regoPath := filepath.Join(repoRoot(t), "policy", "examples", "rego-gates.rego")

	cases := []struct {
		name       string
		changed    []string
		statements []policyyaml.StatementView
	}{
		{
			name:    "prompt change with prompt+eval present passes",
			changed: []string{"examples/tiny-rag/app/system_prompt.txt"},
			statements: []policyyaml.StatementView{
				{AttestationType: "prompt_attestation", StatementID: "s1", PrivacyMode: "hash_only"},
				{AttestationType: "eval_attestation", StatementID: "s2", PrivacyMode: "hash_only"},
			},
		},
		{
			name:    "route change missing slo fails",
			changed: []string{"examples/tiny-rag/route/route.yaml"},
			statements: []policyyaml.StatementView{
				{AttestationType: "route_attestation", StatementID: "s1", PrivacyMode: "hash_only"},
			},
		},
		{
			name:    "plaintext exposure blocked",
			changed: []string{"examples/tiny-rag/app/system_prompt.txt"},
			statements: []policyyaml.StatementView{
				{AttestationType: "prompt_attestation", StatementID: "s1", PrivacyMode: "plaintext_explicit"},
				{AttestationType: "eval_attestation", StatementID: "s2", PrivacyMode: "hash_only"},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			yamlViolations, err := policyyaml.EvaluateWithChanged(policy, tc.statements, tc.changed)
			if err != nil {
				t.Fatalf("yaml evaluate: %v", err)
			}
			sort.Strings(yamlViolations)

			regoResult, err := Evaluate(regoPath, BuildInput(policy, tc.statements, tc.changed))
			if err != nil {
				t.Fatalf("rego evaluate: %v", err)
			}

			if (len(yamlViolations) == 0) != regoResult.Allow {
				t.Fatalf("allow mismatch: yaml=%v rego=%v", len(yamlViolations) == 0, regoResult.Allow)
			}
			if len(yamlViolations) != len(regoResult.Violations) {
				t.Fatalf("violation length mismatch: yaml=%v rego=%v", yamlViolations, regoResult.Violations)
			}
			for i := range yamlViolations {
				if yamlViolations[i] != regoResult.Violations[i] {
					t.Fatalf("violation mismatch at %d: yaml=%q rego=%q", i, yamlViolations[i], regoResult.Violations[i])
				}
			}
		})
	}
}

func repoRoot(t *testing.T) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatalf("cannot resolve test file path")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}
