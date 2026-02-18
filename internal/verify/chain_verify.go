package verify

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

type ChainStatement struct {
	Bundle          string
	StatementID     string
	AttestationType string
	GeneratedAt     string
	DependsOn       []string
}

var requiredChainDeps = map[string][]string{
	"eval_attestation":  {"prompt_attestation", "corpus_attestation"},
	"route_attestation": {"eval_attestation"},
	"slo_attestation":   {"route_attestation"},
}

func VerifyBasicChainConstraints(statement map[string]any) error {
	generatedAt, _ := statement["generated_at"].(string)
	if generatedAt == "" {
		return fmt.Errorf("generated_at is required")
	}
	if _, err := time.Parse(time.RFC3339, generatedAt); err != nil {
		return fmt.Errorf("invalid generated_at: %w", err)
	}
	return nil
}

func VerifyProvenanceChain(statements []ChainStatement) ChainReport {
	report := ChainReport{
		Valid: true,
		Nodes: make([]ChainNode, 0, len(statements)),
		Edges: make([]ChainEdge, 0),
	}
	if len(statements) == 0 {
		return report
	}

	byType := make(map[string][]ChainStatement)
	byID := make(map[string]ChainStatement)
	for _, st := range statements {
		byType[st.AttestationType] = append(byType[st.AttestationType], st)
		if st.StatementID != "" {
			byID[st.StatementID] = st
		}
		report.Nodes = append(report.Nodes, ChainNode{
			Bundle:          st.Bundle,
			StatementID:     st.StatementID,
			AttestationType: st.AttestationType,
			GeneratedAt:     st.GeneratedAt,
			DependsOn:       append([]string(nil), st.DependsOn...),
		})
	}

	violations := map[string]struct{}{}
	for _, st := range statements {
		required := requiredChainDeps[st.AttestationType]
		if len(required) == 0 {
			checkUnknownDependencies(st, byType, byID, violations)
			continue
		}

		// Keep single-bundle verification usable while still enforcing explicit references.
		strict := len(statements) > 1 || len(st.DependsOn) > 0
		if !strict {
			continue
		}

		for _, reqType := range required {
			preds := byType[reqType]
			edge := ChainEdge{
				FromStatementID: st.StatementID,
				FromType:        st.AttestationType,
				ToType:          reqType,
				Satisfied:       true,
			}

			if len(preds) == 0 {
				edge.Satisfied = false
				edge.Detail = "missing_required_attestation_type"
				report.Edges = append(report.Edges, edge)
				violations[fmt.Sprintf("missing chain predecessor: %s requires %s", st.AttestationType, reqType)] = struct{}{}
				continue
			}

			target := preds[0]
			if len(st.DependsOn) > 0 {
				var matched bool
				if contains(st.DependsOn, reqType) {
					matched = true
				} else {
					for _, dep := range st.DependsOn {
						pred, ok := byID[dep]
						if ok && pred.AttestationType == reqType {
							target = pred
							matched = true
							break
						}
					}
				}
				if !matched {
					edge.Satisfied = false
					edge.Detail = "missing_dependency_reference"
					report.Edges = append(report.Edges, edge)
					violations[fmt.Sprintf("missing dependency reference: %s should reference %s", st.StatementID, reqType)] = struct{}{}
					continue
				}
			}

			edge.ToStatementID = target.StatementID
			if target.StatementID == "" {
				edge.ToStatementID = "(by-type)"
			}
			if !ordered(target.GeneratedAt, st.GeneratedAt) {
				edge.Satisfied = false
				edge.Detail = "predecessor_generated_after_successor"
				report.Edges = append(report.Edges, edge)
				violations[fmt.Sprintf("invalid chain order: predecessor %s generated after %s", target.StatementID, st.StatementID)] = struct{}{}
				continue
			}

			report.Edges = append(report.Edges, edge)
		}

		checkUnknownDependencies(st, byType, byID, violations)
	}

	report.Violations = make([]string, 0, len(violations))
	for v := range violations {
		report.Violations = append(report.Violations, v)
	}
	sort.Strings(report.Violations)
	sort.Slice(report.Nodes, func(i, j int) bool {
		if report.Nodes[i].AttestationType == report.Nodes[j].AttestationType {
			return report.Nodes[i].StatementID < report.Nodes[j].StatementID
		}
		return report.Nodes[i].AttestationType < report.Nodes[j].AttestationType
	})
	sort.Slice(report.Edges, func(i, j int) bool {
		if report.Edges[i].FromStatementID == report.Edges[j].FromStatementID {
			return report.Edges[i].ToType < report.Edges[j].ToType
		}
		return report.Edges[i].FromStatementID < report.Edges[j].FromStatementID
	})
	report.Valid = len(report.Violations) == 0
	return report
}

func checkUnknownDependencies(st ChainStatement, byType map[string][]ChainStatement, byID map[string]ChainStatement, violations map[string]struct{}) {
	for _, dep := range st.DependsOn {
		dep = strings.TrimSpace(dep)
		if dep == "" {
			continue
		}
		if _, ok := byType[dep]; ok {
			continue
		}
		if _, ok := byID[dep]; ok {
			continue
		}
		violations[fmt.Sprintf("unknown dependency reference: %s -> %s", st.StatementID, dep)] = struct{}{}
	}
}

func ordered(predecessorGeneratedAt string, successorGeneratedAt string) bool {
	predecessor, err := time.Parse(time.RFC3339, predecessorGeneratedAt)
	if err != nil {
		return true
	}
	successor, err := time.Parse(time.RFC3339, successorGeneratedAt)
	if err != nil {
		return true
	}
	return !predecessor.After(successor)
}

func contains(items []string, want string) bool {
	for _, item := range items {
		if strings.TrimSpace(item) == want {
			return true
		}
	}
	return false
}
