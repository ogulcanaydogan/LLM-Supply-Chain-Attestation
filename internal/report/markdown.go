package report

import (
	"fmt"
	"os"
	"strings"

	"github.com/ogulcanaydogan/llm-supply-chain-attestation/internal/verify"
)

func BuildMarkdown(r verify.Report) string {
	status := "PASS"
	if !r.Passed {
		status = "FAIL"
	}
	var b strings.Builder
	b.WriteString("# LLM Supply-Chain Verification Report\n\n")
	b.WriteString(fmt.Sprintf("- Status: **%s**\n", status))
	b.WriteString(fmt.Sprintf("- Exit Code: `%d`\n", r.ExitCode))
	b.WriteString(fmt.Sprintf("- Bundles Checked: `%d`\n\n", r.BundleCount))

	b.WriteString("## Checks\n\n")
	b.WriteString("| Bundle | Check | Passed | Message |\n")
	b.WriteString("|---|---|---:|---|\n")
	for _, c := range r.Checks {
		b.WriteString(fmt.Sprintf("| %s | %s | %t | %s |\n", c.Bundle, c.Check, c.Passed, strings.ReplaceAll(c.Message, "|", "\\|")))
	}

	if len(r.Violations) > 0 {
		b.WriteString("\n## Violations\n\n")
		for _, v := range r.Violations {
			b.WriteString("- " + v + "\n")
		}
	}

	if len(r.Statements) > 0 {
		b.WriteString("\n## Statements\n\n")
		b.WriteString("| Type | Statement ID | Privacy | Depends On |\n")
		b.WriteString("|---|---|---|---|\n")
		for _, s := range r.Statements {
			deps := "-"
			if len(s.DependsOn) > 0 {
				deps = strings.Join(s.DependsOn, ", ")
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n", s.AttestationType, s.StatementID, s.PrivacyMode, deps))
		}
	}

	b.WriteString("\n## Provenance Chain\n\n")
	b.WriteString(fmt.Sprintf("- Valid: **%t**\n", r.Chain.Valid))
	b.WriteString(fmt.Sprintf("- Nodes: `%d`\n", len(r.Chain.Nodes)))
	b.WriteString(fmt.Sprintf("- Edges: `%d`\n", len(r.Chain.Edges)))

	if len(r.Chain.Edges) > 0 {
		b.WriteString("\n| From Statement | From Type | To Type | To Statement | Satisfied | Detail |\n")
		b.WriteString("|---|---|---|---|---:|---|\n")
		for _, e := range r.Chain.Edges {
			toID := e.ToStatementID
			if toID == "" {
				toID = "-"
			}
			detail := e.Detail
			if detail == "" {
				detail = "ok"
			}
			b.WriteString(fmt.Sprintf("| %s | %s | %s | %s | %t | %s |\n", e.FromStatementID, e.FromType, e.ToType, toID, e.Satisfied, detail))
		}
	}

	if len(r.Chain.Violations) > 0 {
		b.WriteString("\n### Chain Violations\n\n")
		for _, v := range r.Chain.Violations {
			b.WriteString("- " + v + "\n")
		}
	}

	return b.String()
}

func WriteMarkdown(path string, r verify.Report) error {
	return os.WriteFile(path, []byte(BuildMarkdown(r)), 0o644)
}
