package rego

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/policy/yaml"
	oparego "github.com/open-policy-agent/opa/rego"
)

type Input struct {
	ChangedFiles       []string             `json:"changed_files"`
	Statements         []yaml.StatementView `json:"statements"`
	Gates              []yaml.Gate          `json:"gates"`
	PlaintextAllowlist []string             `json:"plaintext_allowlist"`
}

type Result struct {
	Allow      bool     `json:"allow"`
	Violations []string `json:"violations"`
}

func BuildInput(policy yaml.Policy, statements []yaml.StatementView, changed []string) Input {
	return Input{
		ChangedFiles:       changed,
		Statements:         statements,
		Gates:              policy.Gates,
		PlaintextAllowlist: policy.PlaintextAllowlist,
	}
}

func Evaluate(policyPath string, input Input) (Result, error) {
	raw, err := os.ReadFile(policyPath)
	if err != nil {
		return Result{}, fmt.Errorf("read rego policy: %w", err)
	}

	query, err := oparego.New(
		oparego.Query("data.llmsa.gates.result"),
		oparego.Module(filepath.Base(policyPath), string(raw)),
		oparego.Input(input),
	).PrepareForEval(context.Background())
	if err != nil {
		return Result{}, fmt.Errorf("prepare rego query: %w", err)
	}

	rs, err := query.Eval(context.Background())
	if err != nil {
		return Result{}, fmt.Errorf("eval rego policy: %w", err)
	}
	if len(rs) == 0 || len(rs[0].Expressions) == 0 {
		return Result{}, fmt.Errorf("rego policy returned no result")
	}

	out, err := decodeResult(rs[0].Expressions[0].Value)
	if err != nil {
		return Result{}, err
	}
	return out, nil
}

func decodeResult(v any) (Result, error) {
	obj, ok := v.(map[string]any)
	if !ok {
		return Result{}, fmt.Errorf("rego result must be object")
	}
	allow, _ := obj["allow"].(bool)
	violations := decodeViolations(obj["violations"])
	sort.Strings(violations)
	return Result{Allow: allow, Violations: violations}, nil
}

func decodeViolations(v any) []string {
	out := []string{}
	switch raw := v.(type) {
	case []any:
		for _, item := range raw {
			if s, ok := item.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	case map[string]any:
		for key := range raw {
			if key != "" {
				out = append(out, key)
			}
		}
	case map[any]any:
		for key := range raw {
			if s, ok := key.(string); ok && s != "" {
				out = append(out, s)
			}
		}
	}
	return out
}
